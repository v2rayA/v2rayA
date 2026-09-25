package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func testServer(t *testing.T, port int) serverObj.ServerObj {
	t.Helper()
	s, err := ResolveURL(fmt.Sprintf("http-proxy://127.0.0.1:%d#server-%d", port, port))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func namedTestServer(t *testing.T, port int, name string) serverObj.ServerObj {
	t.Helper()
	s, err := ResolveURL(fmt.Sprintf("http-proxy://127.0.0.1:%d#%s", port, name))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func resetSubscription(t *testing.T) *configure.SubscriptionRaw {
	t.Helper()
	t.Cleanup(func() {
		for _, out := range configure.GetOutbounds() {
			_ = configure.ClearConnects(out)
		}
		ids := make([]int, configure.GetLenSubscriptions())
		for i := range ids {
			ids[i] = i
		}
		if len(ids) > 0 {
			_ = configure.RemoveSubscriptions(ids)
		}
		_ = configure.SetRunning(false)
		_ = configure.SetLastKernelExitStatus(configure.LastKernelExitStopped)
	})
	for _, out := range configure.GetOutbounds() {
		if err := configure.ClearConnects(out); err != nil {
			t.Fatal(err)
		}
	}
	indexes := make([]int, configure.GetLenSubscriptions())
	for i := range indexes {
		indexes[i] = i
	}
	if len(indexes) > 0 {
		if err := configure.RemoveSubscriptions(indexes); err != nil {
			t.Fatal(err)
		}
	}
	_ = configure.SetOutboundSetting("proxy", configure.DefaultOutboundSetting())
	cfg := configure.New()
	sub := &configure.SubscriptionRaw{Address: "http://subscription.invalid", FailureIntervalMinutes: 1,
		Servers: []configure.ServerRaw{{ServerObj: testServer(t, 10001)}}}
	cfg.Subscriptions = []*configure.SubscriptionRaw{sub}
	if err := configure.SetConfigure(cfg); err != nil {
		t.Fatal(err)
	}
	if err := configure.AddConnect(configure.NodeRef{TYPE: configure.SubscriptionServerType, Sub: 0, ID: 1, Outbound: "proxy"}); err != nil {
		t.Fatal(err)
	}
	return configure.GetSubscription(0)
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

func TestSubscriptionNameFallbackDoesNotCollapseTwoSelections(t *testing.T) {
	old := resetSubscription(t)
	old.Servers = []configure.ServerRaw{
		{ServerObj: namedTestServer(t, 10001, "same")},
		{ServerObj: namedTestServer(t, 10002, "same")},
	}
	if err := configure.SetSubscription(0, old); err != nil {
		t.Fatal(err)
	}
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	for id := 1; id <= 2; id++ {
		if err := configure.AddConnect(configure.NodeRef{TYPE: configure.SubscriptionServerType, Sub: 0, ID: id, Outbound: "proxy"}); err != nil {
			t.Fatal(err)
		}
	}
	old = configure.GetSubscription(0)
	if err := storeSubscriptionUpdate(0, old, []serverObj.ServerObj{namedTestServer(t, 10003, "same")}, "", false); err != nil {
		t.Fatal(err)
	}
	refs := configure.GetConnectedServersByOutbound("proxy").Get()
	if len(refs) != 2 || refs[0].ID == refs[1].ID {
		t.Fatalf("ambiguous name fallback collapsed selections: %+v", refs)
	}
}

func TestSubscriptionRemapsOneNodeInMultipleGroups(t *testing.T) {
	for _, rotated := range []bool{false, true} {
		t.Run(fmt.Sprintf("rotated=%v", rotated), func(t *testing.T) {
			old := resetSubscription(t)
			const secondGroup = "second"
			if err := configure.AddOutbound(secondGroup); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = configure.RemoveOutbound(secondGroup) })
			if err := configure.AddConnect(configure.NodeRef{TYPE: configure.SubscriptionServerType, Sub: 0, ID: 1, Outbound: secondGroup}); err != nil {
				t.Fatal(err)
			}
			name := "unrelated"
			if rotated {
				name = old.Servers[0].ServerObj.GetName()
			}
			replacement := namedTestServer(t, 10002, name)
			if err := storeSubscriptionUpdate(0, old, []serverObj.ServerObj{replacement}, "", false); err != nil {
				t.Fatal(err)
			}
			wantID, wantNodes := 2, 2
			if rotated {
				wantID, wantNodes = 1, 1
			}
			if got := len(configure.GetSubscription(0).Servers); got != wantNodes {
				t.Fatalf("stored %d nodes, want %d", got, wantNodes)
			}
			for _, group := range []string{"proxy", secondGroup} {
				refs := configure.GetConnectedServersByOutbound(group).Get()
				if len(refs) != 1 || refs[0].ID != wantID {
					t.Fatalf("%s references = %+v, want ID %d", group, refs, wantID)
				}
			}
		})
	}
}

type subscriptionRoundTripper func(*http.Request) (*http.Response, error)

func (f subscriptionRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestSubscriptionDownloadsRespectProxyMode(t *testing.T) {
	for _, mode := range []configure.ProxyMode{configure.ProxyModeProxy, configure.ProxyModePac} {
		for _, transparent := range []configure.TransparentMode{configure.TransparentClose, configure.TransparentProxy} {
			t.Run(fmt.Sprintf("%s/transparent=%s", mode, transparent), func(t *testing.T) {
				old := resetSubscription(t)
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				port := listener.Addr().(*net.TCPAddr).Port
				listener.Close()
				if err := configure.SetPorts(&configure.Ports{Socks5: port, HttpWithPac: port}); err != nil {
					t.Fatal(err)
				}
				setting := configure.GetSettingNotNil()
				setting.ProxyModeWhenSubscribe, setting.Transparent = mode, transparent
				if err := configure.SetSetting(setting); err != nil {
					t.Fatal(err)
				}
				var directRequests atomic.Int32
				provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					directRequests.Add(1)
					fmt.Fprint(w, base64.StdEncoding.EncodeToString([]byte("http-proxy://127.0.0.1:1234#test")))
				}))
				defer provider.Close()
				old.Address = provider.URL
				old.AllowDirectRecovery = true
				if err := configure.SetSubscription(0, old); err != nil {
					t.Fatal(err)
				}
				if err := UpdateSubscription(0, false); err == nil {
					t.Fatal("manual update bypassed the failed proxy")
				}
				if err := ImportSubscription(provider.URL); err == nil {
					t.Fatal("import bypassed the failed proxy")
				}
				if _, _, err := fetchSubscriptionForAutomation(context.Background(), provider.URL, false); err == nil {
					t.Fatal("scheduled update bypassed the failed proxy")
				}
				if got := directRequests.Load(); got != 0 {
					t.Fatalf("non-recovery downloads made %d direct requests", got)
				}
			})
		}
	}
}

func TestSubscriptionRecoveryFallbackKeepsTimeout(t *testing.T) {
	resetSubscription(t)
	setting := configure.GetSettingNotNil()
	setting.ProxyModeWhenSubscribe = configure.ProxyModeProxy
	if err := configure.SetSetting(setting); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	if err := configure.SetPorts(&configure.Ports{Socks5: port}); err != nil {
		t.Fatal(err)
	}
	originalDirect := directSubscriptionClient
	t.Cleanup(func() { directSubscriptionClient = originalDirect })
	fallbackAttempts := 0
	client := &http.Client{Transport: subscriptionRoundTripper(func(r *http.Request) (*http.Response, error) {
		fallbackAttempts++
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 30*time.Second {
			t.Fatal("download attempt has no bounded timeout")
		}
		if !strings.Contains(r.UserAgent(), "WebRequestHelper") {
			t.Fatal("missing subscription user agent")
		}
		body := base64.StdEncoding.EncodeToString([]byte("http-proxy://127.0.0.1:1234#fallback"))
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	directSubscriptionClient = func() *http.Client { return client }
	servers, _, err := fetchSubscriptionForAutomation(context.Background(), "http://subscription.invalid", true)
	if err != nil || len(servers) != 1 || fallbackAttempts != 1 {
		t.Fatalf("fallback failed: %v, %d servers, attempts %d", err, len(servers), fallbackAttempts)
	}
	if client.Timeout != 0 {
		t.Fatal("download modified a shared HTTP client")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := fetchSubscriptionForAutomation(ctx, "http://subscription.invalid", true); !errors.Is(err, context.Canceled) || fallbackAttempts != 1 {
		t.Fatalf("cancelled fetch retried: %v, attempts %d", err, fallbackAttempts)
	}
}

func TestProbeWithCurrentCore(t *testing.T) {
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
