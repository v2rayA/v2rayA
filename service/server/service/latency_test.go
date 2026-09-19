package service

import (
	"errors"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

type closeTrackingBody struct {
	io.Reader
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

func TestHTTPLatencyClosesErrorResponseBody(t *testing.T) {
	body := &closeTrackingBody{Reader: strings.NewReader("bad gateway")}
	which := new(configure.Which)
	setHTTPLatencyResult(which, &http.Response{StatusCode: http.StatusBadGateway, Body: body}, nil, time.Now())
	if !body.closed {
		t.Fatal("response body was not closed")
	}
	if which.Latency != "BAD RESPONSE" {
		t.Fatalf("latency = %q", which.Latency)
	}
}

func TestHttpLatencyRejectsInvalidTestURLBeforeStartingCore(t *testing.T) {
	wasRunning := v2ray.ProcessManager.Running()
	_, err := TestHttpLatency(nil, time.Second, 1, false, "%")
	var coded *common.CodedError
	if !errors.As(err, &coded) || coded.Code != "INVALID_TEST_URL" {
		t.Fatalf("error = %v, want INVALID_TEST_URL", err)
	}
	if got := v2ray.ProcessManager.Running(); got != wasRunning {
		t.Fatalf("core running state changed from %v to %v", wasRunning, got)
	}
}

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "v2raya-service-test-*")
	if err != nil {
		panic(err)
	}
	conf.GetEnvironmentConfig().Config = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestHttpLatencyDropsSubscriptions(t *testing.T) {
	input := []*configure.Which{{NodeRef: configure.NodeRef{TYPE: configure.SubscriptionType, ID: 1}}}
	got, err := TestHttpLatency(input, time.Second, 1, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("subscription remained in latency results: %+v", got)
	}
	if input[0].Latency != "" {
		t.Fatalf("subscription was probed: %q", input[0].Latency)
	}
}

func TestAutoSelectMembersKeepsGroupAndSkipsUnsupported(t *testing.T) {
	existing := []*configure.NodeRef{{TYPE: configure.ServerType, ID: 1, Outbound: "proxy"}}
	sub := &configure.SubscriptionRaw{Servers: []configure.ServerRaw{
		{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "first"}},
		{},
		{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1081, Protocol: "socks5", Name: "third"}},
	}}
	got := autoSelectMembers(4, sub, existing)
	want := []configure.NodeRef{
		{TYPE: configure.ServerType, ID: 1, Outbound: "proxy"},
		{TYPE: configure.SubscriptionServerType, ID: 1, Sub: 4, Outbound: "proxy"},
		{TYPE: configure.SubscriptionServerType, ID: 3, Sub: 4, Outbound: "proxy"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("member %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSupportCheckLeavesNoProducerBehind(t *testing.T) {
	obj := &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "probe"}
	if _, err := isSupportedObj(obj); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		if _, err := isSupportedObj(obj); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(50 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before {
		t.Fatalf("goroutines grew from %d to %d over 20 support checks", before, after)
	}
}

func TestReplaceOutboundConnectionsWritesTheGroupOnce(t *testing.T) {
	const outbound = "replace-test"
	if err := configure.AddOutbound(outbound); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveOutbound(outbound) })
	members := []configure.NodeRef{
		{TYPE: configure.SubscriptionServerType, ID: 2, Sub: 0},
		{TYPE: configure.ServerType, ID: 1, Sub: 7},
		{TYPE: configure.SubscriptionServerType, ID: 2, Sub: 0},
	}
	if err := ReplaceOutboundConnections(outbound, members); err != nil {
		t.Fatal(err)
	}
	got := configure.GetConnectedServersByOutbound(outbound)
	if got == nil || got.Len() != 2 || got.Get()[0].ID != 2 || got.Get()[1].TYPE != configure.ServerType || got.Get()[1].Sub != 0 {
		t.Fatalf("stored group %+v", got)
	}
	if err := ReplaceOutboundConnections(outbound, nil); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetConnectedServersByOutbound(outbound); got != nil && got.Len() != 0 {
		t.Fatalf("group not cleared: %+v", got.Get())
	}
}
