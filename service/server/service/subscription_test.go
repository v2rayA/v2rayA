package service

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestMain(m *testing.M) {
	flag.Parse()
	dir, err := os.MkdirTemp("", "v2raya-subscription-test-")
	if err != nil {
		panic(err)
	}
	args := os.Args
	os.Args = os.Args[:1]
	p := *conf.GetEnvironmentConfig()
	os.Args = args
	p.Config = dir
	p.Lite = true
	p.CoreStartupTimeout = 2
	conf.SetConfig(p)
	code := m.Run()
	_ = db.DB().Close()
	os.RemoveAll(dir)
	os.Exit(code)
}

func testServer(t *testing.T, port int) serverObj.ServerObj {
	t.Helper()
	s, err := ResolveURL(fmt.Sprintf("http-proxy://127.0.0.1:%d#server-%d", port, port))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func resetSubscription(t *testing.T) *configure.SubscriptionRaw {
	t.Helper()
	for _, out := range configure.GetOutbounds() {
		if err := configure.ClearConnects(out); err != nil {
			t.Fatal(err)
		}
	}
	cfg := configure.New()
	sub := &configure.SubscriptionRaw{Address: "http://subscription.invalid", AutoSelect: true,
		Servers: []configure.ServerRaw{{ServerObj: testServer(t, 10001)}}}
	cfg.Subscriptions = []*configure.SubscriptionRaw{sub}
	if err := configure.SetConfigure(cfg); err != nil {
		t.Fatal(err)
	}
	if err := configure.AddConnect(configure.Which{TYPE: configure.SubscriptionServerType, Sub: 0, ID: 1, Outbound: "proxy"}); err != nil {
		t.Fatal(err)
	}
	return sub
}

func TestSubscriptionWaitsAndChoosesFastestHealthy(t *testing.T) {
	old := resetSubscription(t)
	before := configure.GetConnectedServers()
	servers := []serverObj.ServerObj{testServer(t, 10002), testServer(t, 10003), testServer(t, 10004)}
	probe := func(got []serverObj.ServerObj, url string) []subscriptionProbeResult {
		if len(got) != 3 || url != HttpTestURL {
			t.Fatalf("unexpected probe input: %d %s", len(got), url)
		}
		if !reflect.DeepEqual(before, configure.GetConnectedServers()) {
			t.Fatal("selection changed before probe finished")
		}
		if !reflect.DeepEqual(old, configure.GetSubscription(0)) {
			t.Fatal("subscription changed before probe finished")
		}
		return []subscriptionProbeResult{{err: errors.New("closed")}, {latency: 40 * time.Millisecond}, {latency: 10 * time.Millisecond}}
	}
	if err := updateSubscriptionWithProbe(0, old, servers, "updated", probe); err != nil {
		t.Fatal(err)
	}
	w := configure.GetConnectedServersByOutbound("proxy").Get()
	if len(w) != 1 || w[0].ID != 3 || w[0].Sub != 0 {
		t.Fatalf("wrong selection: %+v", w)
	}
	sub := configure.GetSubscription(0)
	if sub.Info != "updated" || sub.Servers[0].Latency != "UNAVAILABLE" || sub.Servers[2].Latency != "10ms" {
		t.Fatalf("wrong result: %+v", sub)
	}
}

func TestSubscriptionAllDeadPreservesState(t *testing.T) {
	old := resetSubscription(t)
	before := configure.GetConnectedServers()
	err := updateSubscriptionWithProbe(0, old, []serverObj.ServerObj{testServer(t, 10002)}, "changed", func([]serverObj.ServerObj, string) []subscriptionProbeResult {
		return []subscriptionProbeResult{{err: errors.New("timeout")}}
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	if !reflect.DeepEqual(old, configure.GetSubscription(0)) || !reflect.DeepEqual(before, configure.GetConnectedServers()) {
		t.Fatal("failed probe changed state")
	}
}

func TestSubscriptionPreservesOtherOutboundsAndRemaps(t *testing.T) {
	old := resetSubscription(t)
	if err := configure.AddOutbound("other"); err != nil {
		t.Fatal(err)
	}
	if err := configure.AddConnect(configure.Which{TYPE: configure.SubscriptionServerType, Sub: 0, ID: 1, Outbound: "other"}); err != nil {
		t.Fatal(err)
	}
	servers := []serverObj.ServerObj{testServer(t, 10002)}
	if err := updateSubscriptionWithProbe(0, old, servers, "", func([]serverObj.ServerObj, string) []subscriptionProbeResult {
		return []subscriptionProbeResult{{latency: time.Millisecond}}
	}); err != nil {
		t.Fatal(err)
	}
	w := configure.GetConnectedServersByOutbound("other").Get()
	if len(w) != 1 || w[0].ID != 2 {
		t.Fatalf("other outbound lost: %+v", w)
	}
	raw, err := w[0].LocateServerRaw()
	if err != nil || raw.ServerObj.ExportToURL() != old.Servers[0].ServerObj.ExportToURL() {
		t.Fatal("other outbound changed server")
	}
}

func TestSubscriptionDoesNotStealOtherSubscription(t *testing.T) {
	resetSubscription(t)
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{AutoSelect: true}}); err != nil {
		t.Fatal(err)
	}
	if !subscriptionOwnsProxy(0) || subscriptionOwnsProxy(1) {
		t.Fatal("active subscription ownership lost")
	}
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	if !subscriptionOwnsProxy(0) || subscriptionOwnsProxy(1) {
		t.Fatal("unconnected selection is not deterministic")
	}
	if err := configure.AddConnect(configure.Which{TYPE: configure.ServerType, ID: 1, Outbound: "proxy"}); err != nil {
		t.Fatal(err)
	}
	if subscriptionOwnsProxy(0) || subscriptionOwnsProxy(1) {
		t.Fatal("manual server would be replaced")
	}
}

func TestSubscriptionDownloadFailuresPreserveState(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"http-error", base64.StdEncoding.EncodeToString([]byte("http-proxy://127.0.0.1:9999")), 503},
		{"empty", "", 200}, {"invalid", "not a subscription", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := resetSubscription(t)
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); fmt.Fprint(w, tc.body) }))
			defer s.Close()
			old.Address = s.URL
			if err := configure.SetSubscription(0, old); err != nil {
				t.Fatal(err)
			}
			before := configure.GetConnectedServers()
			if err := UpdateSubscription(0, false); err == nil {
				t.Fatal("expected failure")
			}
			if !reflect.DeepEqual(old, configure.GetSubscription(0)) || !reflect.DeepEqual(before, configure.GetConnectedServers()) {
				t.Fatal("download failure changed state")
			}
		})
	}
}

func TestSubscriptionInvalidIndex(t *testing.T) {
	resetSubscription(t)
	for _, index := range []int{-1, 99} {
		if err := UpdateSubscription(index, false); err == nil {
			t.Fatal("expected invalid index error")
		}
	}
}

type subscriptionRoundTripper func(*http.Request) (*http.Response, error)

func (f subscriptionRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestSubscriptionDownloadFallbackKeepsTimeout(t *testing.T) {
	original := http.DefaultClient
	defer func() { http.DefaultClient = original }()
	checkRequest := func(r *http.Request) {
		t.Helper()
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 30*time.Second {
			t.Fatal("download attempt has no bounded timeout")
		}
		if !strings.Contains(r.UserAgent(), "WebRequestHelper") {
			t.Fatal("missing subscription user agent")
		}
	}
	firstAttempts, fallbackAttempts := 0, 0
	client := &http.Client{Transport: subscriptionRoundTripper(func(r *http.Request) (*http.Response, error) {
		checkRequest(r)
		firstAttempts++
		return nil, errors.New("selected proxy unavailable")
	})}
	http.DefaultClient = &http.Client{Transport: subscriptionRoundTripper(func(r *http.Request) (*http.Response, error) {
		checkRequest(r)
		fallbackAttempts++
		body := base64.StdEncoding.EncodeToString([]byte("http-proxy://127.0.0.1:1234#fallback"))
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	servers, _, err := ResolveSubscriptionWithClient("http://subscription.invalid", client)
	if err != nil || len(servers) != 1 || firstAttempts != 1 || fallbackAttempts != 1 {
		t.Fatalf("fallback failed: %v, %d servers, attempts %d/%d", err, len(servers), firstAttempts, fallbackAttempts)
	}
	if client.Timeout != 0 || http.DefaultClient.Timeout != 0 {
		t.Fatal("download modified a shared HTTP client")
	}
}

func TestProbeWithRealXray(t *testing.T) {
	if os.Getenv("V2RAYA_V2RAY_BIN") == "" {
		t.Skip("set V2RAYA_V2RAY_BIN to run real-core tests")
	}
	resetSubscription(t)
	for _, status := range []int{204, 302, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Hostname() != "probe.invalid" {
					t.Errorf("request did not traverse proxy: %s", r.URL)
				}
				conn, rw, err := w.(http.Hijacker).Hijack()
				if err != nil {
					return
				}
				defer conn.Close()
				fmt.Fprint(rw, "HTTP/1.1 200 Connection established\r\n\r\n")
				rw.Flush()
				if _, err = http.ReadRequest(rw.Reader); err != nil {
					return
				}
				fmt.Fprintf(rw, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", status, http.StatusText(status))
				rw.Flush()
			}))
			defer proxy.Close()
			port := proxy.Listener.Addr().(*net.TCPAddr).Port
			_, err := probeSubscriptionServer(testServer(t, port), "http://probe.invalid/check", time.Second)
			if (err == nil) != (status == 204) {
				t.Fatalf("status %d: %v", status, err)
			}
		})
	}
	t.Run("closed", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()
		if _, err := probeSubscriptionServer(testServer(t, port), "http://probe.invalid/check", time.Second); err == nil {
			t.Fatal("closed port reported healthy")
		}
	})
	t.Run("blackhole", func(t *testing.T) {
		proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
		defer proxy.Close()
		start := time.Now()
		_, err := probeSubscriptionServer(testServer(t, proxy.Listener.Addr().(*net.TCPAddr).Port), "http://probe.invalid/check", 200*time.Millisecond)
		if err == nil || time.Since(start) > 3*time.Second {
			t.Fatalf("unbounded blackhole: %v", err)
		}
	})
}
