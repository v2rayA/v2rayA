package multiobservatory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/xtls/xray-core/app/dispatcher"
	"github.com/xtls/xray-core/app/observatory"
	"github.com/xtls/xray-core/app/observatory/burst"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/serial"
	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
	"github.com/xtls/xray-core/infra/conf"
	_ "github.com/xtls/xray-core/proxy/blackhole"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/proxy/http"
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"
	_ "github.com/xtls/xray-core/transport/internet/tcp"
)

type groupProxy struct {
	server   *httptest.Server
	down     atomic.Bool
	requests atomic.Int64
}

func newGroupProxy(t *testing.T, dataAddress string) *groupProxy {
	t.Helper()
	p := &groupProxy{}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.down.Load() || r.Method != http.MethodConnect {
			http.Error(w, "unavailable", http.StatusBadGateway)
			return
		}
		upstream, err := net.DialTimeout("tcp", r.Host, time.Second)
		if err != nil {
			http.Error(w, "unavailable", http.StatusBadGateway)
			return
		}
		defer upstream.Close()
		client, buffered, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		defer client.Close()
		if r.Host == dataAddress {
			p.requests.Add(1)
		}
		_, _ = client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		go func() { _, _ = io.Copy(upstream, buffered); _ = upstream.Close() }()
		_, _ = io.Copy(client, upstream)
	}))
	t.Cleanup(p.server.Close)
	return p
}

func TestNativeGroupFallback(t *testing.T) {
	for _, strategy := range []string{"leastping", "leastload", "random", "roundrobin"} {
		t.Run(strategy, func(t *testing.T) {
			var delivered atomic.Int64
			data := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				delivered.Add(1)
				_, _ = io.WriteString(w, "proxied")
			}))
			defer data.Close()
			first := newGroupProxy(t, data.Listener.Addr().String())
			second := newGroupProxy(t, data.Listener.Addr().String())
			var probeDown atomic.Bool
			gate := make(chan struct{})
			var release sync.Once
			probe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-gate:
				case <-r.Context().Done():
					return
				}
				if probeDown.Load() {
					conn, _, err := w.(http.Hijacker).Hijack()
					if err == nil {
						_ = conn.Close()
					}
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer func() { release.Do(func() { close(gate) }); probe.Close() }()

			proxyJSON := func(tag string, p *groupProxy) string {
				addr := p.server.Listener.Addr().(*net.TCPAddr)
				return fmt.Sprintf(`{"tag":%q,"protocol":"http","settings":{"servers":[{"address":"127.0.0.1","port":%d}]}}`, tag, addr.Port)
			}
			// A usable direct default makes an accidental fallthrough observable.
			raw := fmt.Sprintf(`{
				"log":{"loglevel":"none"},
				"outbounds":[{"tag":"direct","protocol":"freedom"},%s,%s,{"tag":"block","protocol":"blackhole"}],
				"routing":{"rules":[{"type":"field","network":"tcp","balancerTag":"group"}],
				"balancers":[{"tag":"group","selector":["node-"],"fallbackTag":"node-first","strategy":{"type":%q}}]}
			}`, proxyJSON("node-first", first), proxyJSON("node-second", second), strategy)
			var config conf.Config
			if err := json.Unmarshal([]byte(raw), &config); err != nil {
				t.Fatal(err)
			}
			built, err := config.Build()
			if err != nil {
				t.Fatal(err)
			}
			built.App = append(built.App, serial.ToTypedMessage(&Config{Observers: []*ObserverConfig{{
				Tag: "group", ProbeUrl: probe.URL, ProbeInterval: int64(time.Minute), SubjectSelector: []string{"node-"},
			}}}))
			instance, err := xcore.New(built)
			if err != nil {
				t.Fatal(err)
			}
			defer instance.Close()
			if err := instance.Start(); err != nil {
				t.Fatal(err)
			}
			observer := instance.GetFeature(extension.ObservatoryType()).(*MultiObservatory)
			child := observer.children["group"].(*burst.Observer)
			transport := &http.Transport{DisableKeepAlives: true, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				dest, err := xnet.ParseDestination("tcp:" + address)
				if err != nil {
					return nil, err
				}
				return xcore.Dial(ctx, instance, dest)
			}}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
			request := func(wantOK bool) {
				t.Helper()
				resp, err := client.Get(data.URL)
				ok := false
				if err == nil {
					body, readErr := io.ReadAll(resp.Body)
					_ = resp.Body.Close()
					ok = readErr == nil && string(body) == "proxied"
				}
				if ok != wantOK {
					t.Fatalf("request success=%v, want %v (error %v)", ok, wantOK, err)
				}
			}
			check := func(firstAlive, secondAlive bool) {
				t.Helper()
				// Flush the native sampling window, without waiting for the periodic timer.
				for range 20 {
					child.Check([]string{"node-first", "node-second"})
					report, _ := observer.GetObservationByTag("group", context.Background())
					alive := make(map[string]bool)
					for _, s := range report.(*observatory.ObservationResult).Status {
						alive[s.OutboundTag] = s.Alive
					}
					if len(alive) == 2 && alive["node-first"] == firstAlive && alive["node-second"] == secondAlive {
						return
					}
				}
				t.Fatal("health observations did not converge")
			}

			request(true) // Cold start, before any probe can finish.
			if strategy == "leastping" || strategy == "leastload" {
				if first.requests.Load() != 1 {
					t.Fatal("cold start did not use the group's fallback")
				}
			}
			release.Do(func() { close(gate) })
			check(true, true)
			firstBefore, secondBefore := first.requests.Load(), second.requests.Load()
			for range 8 {
				request(true)
			}
			if strategy == "roundrobin" && (first.requests.Load()-firstBefore != 4 || second.requests.Load()-secondBefore != 4) {
				t.Fatal("healthy round-robin did not split eight connections equally")
			}
			probeDown.Store(true)
			check(false, false)
			before := first.requests.Load()
			for range 4 {
				request(true)
			}
			if first.requests.Load() != before+4 {
				t.Fatal("probe failure did not select the first group member")
			}

			first.down.Store(true)
			beforeSecond := second.requests.Load()
			request(false)
			if second.requests.Load() != beforeSecond {
				t.Fatal("fallback tried a member other than the first")
			}
			second.down.Store(true)
			beforeDelivered := delivered.Load()
			request(false)
			if delivered.Load() != beforeDelivered {
				t.Fatal("traffic escaped to direct while both proxies were down")
			}

			second.down.Store(false)
			probeDown.Store(false)
			check(false, true)
			before = second.requests.Load()
			for range 4 {
				request(true)
			}
			if second.requests.Load() != before+4 {
				t.Fatal("healthy selection did not recover without a restart")
			}
			if delivered.Load() != first.requests.Load()+second.requests.Load() {
				t.Fatal("a request bypassed both group members")
			}
		})
	}
}
