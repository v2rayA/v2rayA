package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

func TestNodeProbeValidation(t *testing.T) {
	for _, target := range []string{"", "https://example.com/generate_204", "http://127.0.0.1:8080/check?q=1"} {
		if err := ValidateNodeProbe("proxy", target); err != nil {
			t.Fatal(err)
		}
	}
	for _, target := range []string{"%", "file:///tmp/test", "https:///path", "https://user:pass@example.com", "//example.com"} {
		if err := ValidateNodeProbe("proxy", target); err == nil {
			t.Errorf("accepted %q", target)
		}
	}
	if err := ValidateNodeProbe("unknown", ""); err == nil {
		t.Fatal("accepted unknown mode")
	}
}

func TestNodeProbeOnlyAccepts2xx(t *testing.T) {
	for _, status := range []int{200, 204, 299, 301, 302, 404, 503} {
		which := new(configure.Which)
		setHTTPLatencyResult(which, &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}, nil, time.Now())
		if strings.HasSuffix(which.Latency, "ms") != (status < 300) {
			t.Fatalf("status %d: %s", status, which.Latency)
		}
	}
}

func TestNodeProbeRealCore(t *testing.T) {
	bin := os.Getenv("V2RAYA_TEST_CORE")
	if bin == "" {
		t.Skip("set V2RAYA_TEST_CORE to exercise the real core")
	}
	oldBin := conf.GetEnvironmentConfig().V2rayBin
	conf.GetEnvironmentConfig().V2rayBin = bin
	t.Cleanup(func() { conf.GetEnvironmentConfig().V2rayBin = oldBin })
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	var tmpl v2ray.Template
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"inbounds":[{"listen":"127.0.0.1","port":%d,"protocol":"socks","settings":{"auth":"password","accounts":[{"user":"probe","pass":"correct"}]}}],"outbounds":[{"protocol":"freedom","tag":"direct"}]}`, port)), &tmpl); err != nil {
		t.Fatal(err)
	}
	stop, err := startProbeCore(context.Background(), &tmpl, []string{strconv.Itoa(port)})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	var requests atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/custom" {
			t.Errorf("wrong target: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer target.Close()
	base := configure.GetLenServers()
	for _, password := range []string{"correct", "wrong"} {
		if err := configure.AppendServers([]*configure.ServerRaw{{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: port, Protocol: "socks5", Username: "probe", Password: password}}}); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { _ = configure.RemoveServers([]int{base, base + 1}) })
	nodes := []*configure.Which{{NodeRef: configure.NodeRef{TYPE: configure.ServerType, ID: base + 1}}, {NodeRef: configure.NodeRef{TYPE: configure.ServerType, ID: base + 2}}}
	ping, err := Ping(nodes, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range ping {
		if !strings.HasSuffix(node.Latency, "ms") {
			t.Fatalf("open port failed: %s", node.Latency)
		}
	}
	result, err := TestHttpLatency(context.Background(), nodes, time.Second, 2, false, target.URL+"/custom")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(result[0].Latency, "ms") || strings.HasSuffix(result[1].Latency, "ms") || requests.Load() != 1 {
		t.Fatalf("results: %s, %s; target requests: %d", result[0].Latency, result[1].Latency, requests.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := TestHttpLatency(ctx, nodes, time.Second, 2, false, target.URL); err != context.Canceled {
		t.Fatalf("cancel: %v", err)
	}
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), time.Second)
	if err != nil {
		t.Fatalf("existing core stopped: %v", err)
	}
	conn.Close()
}
