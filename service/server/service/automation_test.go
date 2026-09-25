package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/touch"
)

func testAutomation(t *testing.T, mode configure.SubscriptionUpdateMode, regular, failure int) (*automation, *int, *bool) {
	t.Helper()
	resetSubscription(t)
	sub := configure.GetSubscription(0)
	sub.UpdateMode, sub.UpdateIntervalMinutes, sub.FailureIntervalMinutes = mode, regular, failure
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	now := time.Unix(1000, 0)
	a.now = func() time.Time { return now }
	requests := 0
	healthy := true
	a.fetch = func(context.Context, string, bool) ([]serverObj.ServerObj, string, error) {
		requests++
		return []serverObj.ServerObj{testServer(t, 10001)}, "", nil
	}
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		results := make([]subscriptionProbeResult, len(nodes))
		if !healthy {
			for i := range results {
				results[i].err = errors.New("unavailable")
			}
		}
		return results
	}
	return a, &requests, &healthy
}

func TestSubscriptionUpdateModes(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		a, requests, _ := testAutomation(t, configure.SubscriptionUpdateDisabled, 0, 1)
		a.step(context.Background())
		if *requests != 0 || !a.nextDeadline().IsZero() {
			t.Fatal("disabled subscription scheduled work")
		}
	})
	t.Run("on-start", func(t *testing.T) {
		a, requests, _ := testAutomation(t, configure.SubscriptionUpdateOnStart, 0, 1)
		a.step(context.Background())
		a.step(context.Background())
		if *requests != 1 || !a.nextDeadline().IsZero() {
			t.Fatalf("on-start requests=%d, next=%v", *requests, a.nextDeadline())
		}
	})
	t.Run("interval", func(t *testing.T) {
		a, requests, _ := testAutomation(t, configure.SubscriptionUpdateAtInterval, 3, 1)
		now := time.Unix(1000, 0)
		a.now = func() time.Time { return now }
		a.step(context.Background())
		now = now.Add(2 * time.Minute)
		a.step(context.Background())
		now = now.Add(time.Minute)
		a.step(context.Background())
		if *requests != 2 {
			t.Fatalf("interval requests=%d", *requests)
		}
	})
	t.Run("interval-failsafe", func(t *testing.T) {
		a, requests, healthy := testAutomation(t, configure.SubscriptionUpdateIntervalFailsafe, 10, 1)
		now := time.Unix(1000, 0)
		a.now = func() time.Time { return now }
		*healthy = false
		a.step(context.Background())
		now = now.Add(time.Minute)
		a.step(context.Background())
		*healthy = true
		now = now.Add(time.Minute)
		a.step(context.Background())
		now = now.Add(time.Minute)
		a.step(context.Background())
		if *requests != 3 {
			t.Fatalf("failsafe requests=%d", *requests)
		}
	})
}

func TestSubscriptionScheduleUsesDatabaseIDAndPolicyFieldsOnly(t *testing.T) {
	resetSubscription(t)
	first := configure.GetSubscription(0)
	first.UpdateMode, first.UpdateIntervalMinutes = configure.SubscriptionUpdateAtInterval, 60
	if err := configure.SetSubscription(0, first); err != nil {
		t.Fatal(err)
	}
	second := &configure.SubscriptionRaw{
		Address: "https://second.invalid", UpdateMode: configure.SubscriptionUpdateAtInterval,
		UpdateIntervalMinutes: 60, FailureIntervalMinutes: 1,
		Servers: []configure.ServerRaw{{ServerObj: testServer(t, 10002)}},
	}
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{second}); err != nil {
		t.Fatal(err)
	}
	second = configure.GetSubscription(1)
	a := newAutomation()
	now := time.Unix(1000, 0)
	a.now = func() time.Time { return now }
	requests := map[int]bool{}
	a.fetch = func(_ context.Context, address string, _ bool) ([]serverObj.ServerObj, string, error) {
		if address == second.Address {
			requests[2] = true
		} else {
			requests[1] = true
		}
		return []serverObj.ServerObj{testServer(t, 10003)}, "", nil
	}
	a.step(context.Background())
	delete(requests, 2)
	second = configure.GetSubscription(1)
	second.Status = "changed"
	second.Servers[0].Latency = "12ms"
	if err := configure.SetSubscription(1, second); err != nil {
		t.Fatal(err)
	}
	if err := configure.RemoveSubscriptions([]int{0}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	a.step(context.Background())
	if requests[2] {
		t.Fatal("latency change or index shift reset the later subscription timer")
	}
}

func TestSubscriptionPatchWithoutPolicyKeepsPolicy(t *testing.T) {
	resetSubscription(t)
	sub := configure.GetSubscription(0)
	sub.UpdateMode, sub.UpdateIntervalMinutes, sub.FailureIntervalMinutes = configure.SubscriptionUpdateIntervalFailsafe, 15, 2
	sub.AllowDirectRecovery = true
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	request := touch.Subscription{ID: 1, Address: sub.Address, Remarks: "cached client", AutoSelect: sub.AutoSelect}
	if err := ModifySubscriptionRemark(request); err != nil {
		t.Fatal(err)
	}
	got := configure.GetSubscription(0)
	if got.UpdateMode != configure.SubscriptionUpdateIntervalFailsafe || got.UpdateIntervalMinutes != 15 || got.FailureIntervalMinutes != 2 || !got.AllowDirectRecovery {
		t.Fatalf("cached PATCH changed policy: %+v", got)
	}
}

func TestDirectRecoveryRequiresOptInAndOutage(t *testing.T) {
	for _, allowed := range []bool{false, true} {
		t.Run(fmt.Sprintf("allowed=%v", allowed), func(t *testing.T) {
			a, _, healthy := testAutomation(t, configure.SubscriptionUpdateIntervalFailsafe, 10, 1)
			sub := configure.GetSubscription(0)
			sub.AllowDirectRecovery = allowed
			if err := configure.SetSubscription(0, sub); err != nil {
				t.Fatal(err)
			}
			now := time.Unix(1000, 0)
			a.now = func() time.Time { return now }
			var permissions []bool
			fetch := a.fetch
			a.fetch = func(ctx context.Context, address string, allowDirect bool) ([]serverObj.ServerObj, string, error) {
				permissions = append(permissions, allowDirect)
				return fetch(ctx, address, allowDirect)
			}
			a.step(context.Background())
			*healthy = false
			now = a.subscriptions[sub.DatabaseID].nextHealth
			a.step(context.Background())
			if len(permissions) != 2 || permissions[0] || permissions[1] != allowed {
				t.Fatalf("direct download permissions = %v; want [false %v]", permissions, allowed)
			}
			if got := configure.GetSubscriptions()[0].AllowDirectRecovery; got != allowed {
				t.Fatal("refresh or list read changed direct-download consent")
			}
		})
	}
}

func TestSubscriptionPatchCanRevokeDirectRecovery(t *testing.T) {
	sub := resetSubscription(t)
	if sub.AllowDirectRecovery {
		t.Fatal("new subscription enabled direct recovery")
	}
	allow := true
	request := touch.Subscription{ID: 1, Address: sub.Address, AllowDirectRecovery: &allow}
	if err := ModifySubscriptionRemark(request); err != nil {
		t.Fatal(err)
	}
	generated := touch.GenerateTouch().Subscriptions[0]
	if generated.AllowDirectRecovery == nil || !*generated.AllowDirectRecovery {
		t.Fatal("saved opt-in was not returned to the GUI")
	}
	allow = false
	if err := ModifySubscriptionRemark(request); err != nil {
		t.Fatal(err)
	}
	if configure.GetSubscription(0).AllowDirectRecovery {
		t.Fatal("explicit false did not revoke direct-download consent")
	}
}

func TestFailsafeProbeCannotMutateStoredNode(t *testing.T) {
	resetSubscription(t)
	node, err := ResolveURL("vless://00000000-0000-0000-0000-000000000001@example.com:443?type=raw&security=none#raw")
	if err != nil {
		t.Fatal(err)
	}
	sub := configure.GetSubscription(0)
	sub.UpdateMode, sub.UpdateIntervalMinutes, sub.FailureIntervalMinutes = configure.SubscriptionUpdateIntervalFailsafe, 10, 1
	sub.Servers = []configure.ServerRaw{{ServerObj: node}}
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	want := configure.GetSubscription(0).Servers[0].ServerObj.ExportToURL()
	a := newAutomation()
	a.fetch = func(context.Context, string, bool) ([]serverObj.ServerObj, string, error) {
		return []serverObj.ServerObj{node}, "", nil
	}
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		if v, ok := nodes[0].(*serverObj.V2Ray); ok {
			v.Net = "tcp"
		}
		return make([]subscriptionProbeResult, len(nodes))
	}
	a.step(context.Background())
	if got := configure.GetSubscription(0).Servers[0].ServerObj.ExportToURL(); got != want {
		t.Fatalf("probe mutated stored node: %q != %q", got, want)
	}
}

func TestFailsafeCancellationKeepsDeadline(t *testing.T) {
	for _, phase := range []string{"health-probe", "recovery-fetch", "post-update-probe", "regular-fetch"} {
		t.Run(phase, func(t *testing.T) {
			a, requests, healthy := testAutomation(t, configure.SubscriptionUpdateIntervalFailsafe, 10, 2)
			now := time.Unix(1000, 0)
			a.now = func() time.Time { return now }
			a.step(context.Background())
			sub := configure.GetSubscription(0)
			state := a.subscriptions[sub.DatabaseID]
			regularDeadline := state.nextUpdate
			now = state.nextHealth
			if phase == "regular-fetch" {
				now = state.nextUpdate
			}
			*healthy = false
			fetch, probe := a.fetch, a.probe
			probeCalls := 0
			a.fetch = func(ctx context.Context, address string, recovery bool) ([]serverObj.ServerObj, string, error) {
				if phase == "recovery-fetch" || phase == "regular-fetch" {
					// Cancellation after a slow request must back off from completion.
					now = now.Add(3 * time.Minute)
					CancelAutomation()
					return nil, "", ctx.Err()
				}
				return fetch(ctx, address, recovery)
			}
			a.probe = func(ctx context.Context, nodes []serverObj.ServerObj, url string) []subscriptionProbeResult {
				probeCalls++
				if phase == "health-probe" || (phase == "post-update-probe" && probeCalls == 2) {
					CancelAutomation()
				}
				return probe(ctx, nodes, url)
			}
			next := a.step(context.Background())
			if next.IsZero() || !next.After(now) || next.After(now.Add(2*time.Minute)) {
				t.Fatalf("cancelled %s lost the worker deadline: now=%v, next=%v, state=%+v", phase, now, next, state)
			}
			if phase != "regular-fetch" && state.nextUpdate != regularDeadline {
				t.Fatal("cancellation postponed the regular deadline")
			}
			a.fetch, a.probe = fetch, probe
			now = next
			before := *requests
			a.step(context.Background())
			if *requests != before+1 {
				t.Fatal("recovery did not resume at the retained deadline")
			}
		})
	}
}

func TestFailsafeDoesNotTreatUnprobeableNodesAsAnOutage(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		t.Run(fmt.Sprintf("mixed=%v", mixed), func(t *testing.T) {
			a, _, _ := testAutomation(t, configure.SubscriptionUpdateIntervalFailsafe, 10, 2)
			now := time.Unix(1000, 0)
			a.now = func() time.Time { return now }
			nodes := []serverObj.ServerObj{&serverObj.Plugin{Protocol: serverObj.PluginManagerScheme, Link: "plugin://fixture"}}
			if mixed {
				nodes = append(nodes, testServer(t, 10002))
			}
			requests := 0
			a.fetch = func(_ context.Context, _ string, recovery bool) ([]serverObj.ServerObj, string, error) {
				requests++
				if recovery {
					t.Fatal("unprobeable candidates triggered recovery")
				}
				return nodes, "", nil
			}
			a.probe = func(context.Context, []serverObj.ServerObj, string) []subscriptionProbeResult {
				t.Fatal("unprobeable subscription started candidate cores")
				return nil
			}
			a.step(context.Background())
			for range 5 {
				now = now.Add(2 * time.Minute)
				a.step(context.Background())
			}
			if requests != 2 {
				t.Fatalf("requests=%d, want startup and regular update only", requests)
			}
		})
	}
}

func TestSwitchingToOnStartUpdatesOnceImmediately(t *testing.T) {
	a, requests, _ := testAutomation(t, configure.SubscriptionUpdateAtInterval, 10, 1)
	a.step(context.Background())
	sub := configure.GetSubscription(0)
	sub.UpdateMode = configure.SubscriptionUpdateOnStart
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	a.step(context.Background())
	a.step(context.Background())
	if *requests != 2 || !a.nextDeadline().IsZero() {
		t.Fatalf("on-start after policy edit: requests=%d, next=%v", *requests, a.nextDeadline())
	}
}
