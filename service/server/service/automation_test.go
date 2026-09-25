package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/touch"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

func TestAutomaticGroupsUseEntireCatalogAndRespectOff(t *testing.T) {
	sub := resetSubscription(t)
	sub.Servers = append(sub.Servers, configure.ServerRaw{ServerObj: testServer(t, 10002)})
	configure.SetSubscription(0, sub)
	configure.AppendSubscriptions([]*configure.SubscriptionRaw{{Address: "https://second.invalid", Servers: []configure.ServerRaw{{ServerObj: testServer(t, 10003)}}}})
	configure.AppendServers([]*configure.ServerRaw{{ServerObj: testServer(t, 10004)}})
	setting := configure.DefaultOutboundSetting()
	if setting.ProbeInterval != "300s" || setting.AutoAdd {
		t.Fatal("unsafe defaults")
	}
	setting.AutoAdd = true
	configure.SetOutboundSetting("proxy", setting)
	a := newAutomation()
	now := time.Unix(1000, 0)
	a.now = func() time.Time { return now }
	calls := 0
	dead := map[int]bool{10001: true}
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		calls++
		results := make([]subscriptionProbeResult, len(nodes))
		for i, n := range nodes {
			if dead[n.GetPort()] {
				results[i].err = errors.New("unavailable")
			}
		}
		return results
	}
	a.step(context.Background())
	if n := configure.GetConnectedServersByOutbound("proxy").Len(); n != 3 {
		t.Fatalf("got %d members, want all 3 reachable nodes", n)
	}
	if v2ray.ProcessManager.Running() {
		t.Fatal("automation started manually stopped core")
	}
	a.step(context.Background())
	if calls != 1 {
		t.Fatal("unchanged catalog rechecked before interval")
	}
	dead[10002], dead[10003], dead[10004] = true, true, true
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if configure.GetConnectedServersByOutbound("proxy").Len() != 0 {
		t.Fatal("dead members retained")
	}
	if len(configure.GetSubscription(0).Servers) != 2 {
		t.Fatal("health checks removed catalog nodes")
	}
	delete(dead, 10003)
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if configure.GetConnectedServersByOutbound("proxy").Len() != 1 {
		t.Fatal("recovered node not re-added")
	}
	setting.AutoAdd = false
	configure.SetOutboundSetting("proxy", setting)
	before := configure.GetConnectedServers()
	now = now.Add(time.Hour)
	a.step(context.Background())
	if !reflect.DeepEqual(before, configure.GetConnectedServers()) || calls != 3 {
		t.Fatal("disabled group changed or probed")
	}
}

func TestSubscriptionRetryAndRegularTimersAreIndependent(t *testing.T) {
	sub := resetSubscription(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		fmt.Fprint(w, base64.StdEncoding.EncodeToString([]byte("http-proxy://127.0.0.1:10001#server-10001")))
	}))
	defer server.Close()
	sub.Address, sub.AutoUpdate, sub.UpdateIntervalMinutes, sub.FailureIntervalMinutes = server.URL, true, 3, 1
	configure.SetSubscription(0, sub)
	a := newAutomation()
	now := time.Unix(1000, 0)
	a.now = func() time.Time { return now }
	healthy := false
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		results := make([]subscriptionProbeResult, len(nodes))
		for i := range results {
			if !healthy {
				results[i].err = errors.New("dead")
			}
		}
		return results
	}
	a.step(context.Background())
	if requests != 0 {
		t.Fatal("downloaded before either timer")
	}
	now = now.Add(time.Minute)
	a.step(context.Background())
	now = now.Add(time.Minute)
	healthy = true
	a.step(context.Background())
	if requests != 2 {
		t.Fatalf("failure timer downloads=%d", requests)
	}
	now = now.Add(time.Minute)
	a.step(context.Background())
	if requests != 3 {
		t.Fatal("failure retries postponed unconditional update")
	}
	now = now.Add(time.Minute)
	a.step(context.Background())
	if requests != 3 {
		t.Fatal("failure retries continued after recovery")
	}
	sub = configure.GetSubscription(0)
	sub.UpdateIntervalMinutes = 0
	configure.SetSubscription(0, sub)
	a.step(context.Background())
	now = now.Add(time.Hour)
	a.step(context.Background())
	if requests != 3 {
		t.Fatal("regular=0 still downloads healthy subscription")
	}
	if v2ray.ProcessManager.Running() {
		t.Fatal("subscription refresh started core")
	}
}

func TestAutomationCancellationDiscardsStaleProbes(t *testing.T) {
	resetSubscription(t)
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	configure.SetOutboundSetting("proxy", setting)
	a := newAutomation()
	entered := make(chan struct{})
	a.probe = func(ctx context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		close(entered)
		<-ctx.Done()
		return make([]subscriptionProbeResult, len(nodes))
	}
	done := make(chan struct{})
	go func() { a.step(context.Background()); close(done) }()
	<-entered
	if !ConfigurationMu.TryLock() {
		t.Fatal("probe blocks user settings")
	}
	CancelAutomation()
	setting.AutoAdd = false
	configure.SetOutboundSetting("proxy", setting)
	configure.ClearConnects("proxy")
	ConfigurationMu.Unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancellation blocked")
	}
	if configure.GetConnectedServersByOutbound("proxy").Len() != 0 {
		t.Fatal("stale results restored members")
	}
}

func TestSubscriptionIntervalsRequiredAndRoundTrip(t *testing.T) {
	sub := resetSubscription(t)
	request := touch.Subscription{ID: 1, Address: sub.Address, AutoUpdate: true}
	if ModifySubscriptionRemark(request) == nil {
		t.Fatal("missing required intervals accepted")
	}
	zero, one := 0, 1
	request.UpdateIntervalMinutes, request.FailureIntervalMinutes = &zero, &zero
	if ModifySubscriptionRemark(request) == nil {
		t.Fatal("zero failure interval accepted")
	}
	request.FailureIntervalMinutes = &one
	if err := ModifySubscriptionRemark(request); err != nil {
		t.Fatal(err)
	}
	got := touch.GenerateTouch().Subscriptions[0]
	if !got.AutoUpdate || *got.UpdateIntervalMinutes != 0 || *got.FailureIntervalMinutes != 1 {
		t.Fatalf("settings not persisted: %+v", got)
	}
}

func TestSubscriptionRefreshDoesNotPopulateManualGroup(t *testing.T) {
	sub := resetSubscription(t)
	configure.ClearConnects("proxy")
	if err := storeSubscriptionUpdate(0, sub, []serverObj.ServerObj{testServer(t, 10002)}, "", false); err != nil {
		t.Fatal(err)
	}
	if configure.GetConnectedServersByOutbound("proxy").Len() != 0 {
		t.Fatal("subscription owns group membership again")
	}
}

func TestSubscriptionDownloadCancellationDoesNotBlockSettings(t *testing.T) {
	sub := resetSubscription(t)
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	sub.Address, sub.AutoUpdate, sub.UpdateIntervalMinutes = server.URL, true, 1
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	now := time.Unix(1000, 0)
	a.now = func() time.Time { return now }
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		return make([]subscriptionProbeResult, len(nodes))
	}
	a.step(context.Background())
	now = now.Add(time.Minute)
	done := make(chan struct{})
	go func() { a.step(context.Background()); close(done) }()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("download did not start")
	}
	if !ConfigurationMu.TryLock() {
		t.Fatal("download holds configuration lock")
	}
	CancelAutomation()
	sub.AutoUpdate = false
	configure.SetSubscription(0, sub)
	ConfigurationMu.Unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("download ignored cancellation")
	}
	if configure.GetSubscription(0).AutoUpdate {
		t.Fatal("late download re-enabled subscription")
	}
}
