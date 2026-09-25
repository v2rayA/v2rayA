package service

import (
	"context"
	"errors"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestMonitorRecoveryCanBeCancelledWithoutBlockingSettings(t *testing.T) {
	old := resetSubscription(t)
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	old.Monitor, old.Address = true, server.URL
	if err := configure.SetSubscription(0, old); err != nil {
		t.Fatal(err)
	}
	configure.SetRunning(true)
	defer configure.SetRunning(false)
	target := activeMonitorTarget()
	target.client = server.Client()
	done := make(chan error, 1)
	go func() { done <- recoverMonitoredSubscription(context.Background(), target) }()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("recovery did not start")
	}
	if !ConfigurationMu.TryLock() {
		t.Fatal("background download blocks settings")
	}
	old.Monitor = false
	configure.SetSubscription(0, old)
	CancelSubscriptionRecovery()
	ConfigurationMu.Unlock()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("recovery did not cancel promptly")
	}
	if configure.GetSubscription(0).Monitor {
		t.Fatal("recovery overwrote disabled setting")
	}
}

func TestMonitorRequiresContinuousMinute(t *testing.T) {
	start := time.Unix(1000, 0)
	s := monitorState{}
	for _, elapsed := range []time.Duration{0, 10 * time.Second, 59 * time.Second} {
		if s.observe(start.Add(elapsed), false) {
			t.Fatal("recovery started before one minute")
		}
	}
	if !s.observe(start.Add(time.Minute), false) {
		t.Fatal("recovery did not start after one minute")
	}
	if s.observe(start.Add(65*time.Second), true) {
		t.Fatal("healthy result requested recovery")
	}
	if s.observe(start.Add(70*time.Second), false) {
		t.Fatal("intermittent failures accumulated")
	}
	if s.observe(start.Add(129*time.Second), false) {
		t.Fatal("new failure window too short")
	}
	if !s.observe(start.Add(130*time.Second), false) {
		t.Fatal("new continuous outage not detected")
	}
}

func TestMonitorRetriesImmediatelyThenBoundsLoad(t *testing.T) {
	s := monitorState{}
	for _, want := range []time.Duration{0, 5 * time.Second, 10 * time.Second, 20 * time.Second, 30 * time.Second, 30 * time.Second} {
		if got := s.retryDelay(); got != want {
			t.Fatalf("retry delay %s, want %s", got, want)
		}
	}
}

func TestMonitorTogglePersistsAndRespectsManualStop(t *testing.T) {
	old := resetSubscription(t)
	if old.Monitor {
		t.Fatal("monitor must default off")
	}
	if err := ModifySubscriptionRemark(touch.Subscription{ID: 1, Address: old.Address, Monitor: true}); err != nil {
		t.Fatal(err)
	}
	if !configure.GetSubscription(0).Monitor || !touch.GenerateTouch().Subscriptions[0].Monitor {
		t.Fatal("monitor setting not round-tripped")
	}
	if err := configure.SetRunning(false); err != nil {
		t.Fatal(err)
	}
	if activeMonitorTarget() != nil {
		t.Fatal("monitor would undo a manual stop")
	}
	if err := configure.SetRunning(true); err != nil {
		t.Fatal(err)
	}
	defer configure.SetRunning(false)
	if activeMonitorTarget() == nil {
		t.Fatal("enabled active subscription not monitored")
	}
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	if activeMonitorTarget() != nil {
		t.Fatal("unselected subscription monitored")
	}
}

func TestFirstEntryPolicyNeverFallsBack(t *testing.T) {
	old := resetSubscription(t)
	old.PreferFirst = true
	servers := []serverObj.ServerObj{testServer(t, 19001), testServer(t, 19002)}
	called := 0
	probe := func(nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		called += len(nodes)
		return []subscriptionProbeResult{{err: errors.New("first is dead")}}
	}
	if err := updateSubscriptionWithProbe(0, old, servers, "", probe); err != nil {
		t.Fatal(err)
	}
	if configure.GetSubscription(0).Servers[1].Latency != "" {
		t.Fatal("untested entry falsely marked unavailable")
	}
	if called != 1 {
		t.Fatalf("checked %d nodes in first-entry mode", called)
	}
	if w := configure.GetConnectedServersByOutbound("proxy").Get(); len(w) != 1 || w[0].ID != 1 {
		t.Fatalf("did not retain first entry: %v", w)
	}
	if err := Connect(&configure.Which{TYPE: configure.SubscriptionServerType, Sub: 0, ID: 2, Outbound: "proxy"}); err == nil {
		t.Fatal("manual selection bypassed first-entry policy")
	}
	// A reordered list follows position, not the previous endpoint's identity.
	old = configure.GetSubscription(0)
	servers[0], servers[1] = servers[1], servers[0]
	if err := updateSubscriptionWithProbe(0, old, servers, "", probe); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetSubscription(0).Servers[0].ServerObj.ExportToURL(); got != servers[0].ExportToURL() {
		t.Fatal("reorder did not replace first entry")
	}
}

func TestIncompleteProbeKeepsState(t *testing.T) {
	old := resetSubscription(t)
	servers := []serverObj.ServerObj{testServer(t, 19001), testServer(t, 19002)}
	err := updateSubscriptionWithProbe(0, old, servers, "", func([]serverObj.ServerObj, string) []subscriptionProbeResult { return nil })
	if err == nil {
		t.Fatal("accepted incomplete probe results")
	}
	if !reflect.DeepEqual(old, configure.GetSubscription(0)) {
		t.Fatal("incomplete results changed subscription")
	}
}

func TestMonitorSnapshotDetectsMetadataChange(t *testing.T) {
	old := resetSubscription(t)
	old.Monitor = true
	configure.SetSubscription(0, old)
	configure.SetRunning(true)
	defer configure.SetRunning(false)
	target := activeMonitorTarget()
	old.Remarks = "changed while probing"
	configure.SetSubscription(0, old)
	if sameMonitorTarget(target, activeMonitorTarget()) {
		t.Fatal("stale recovery could overwrite changed metadata")
	}
}
