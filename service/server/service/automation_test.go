package service

import (
	"context"
	"errors"
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
	a.fetch = func(context.Context, string) ([]serverObj.ServerObj, string, error) {
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
	a.fetch = func(_ context.Context, address string) ([]serverObj.ServerObj, string, error) {
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
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	request := touch.Subscription{ID: 1, Address: sub.Address, Remarks: "cached client"}
	if err := ModifySubscriptionRemark(request); err != nil {
		t.Fatal(err)
	}
	got := configure.GetSubscription(0)
	if got.UpdateMode != configure.SubscriptionUpdateIntervalFailsafe || got.UpdateIntervalMinutes != 15 || got.FailureIntervalMinutes != 2 {
		t.Fatalf("cached PATCH changed policy: %+v", got)
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
	a.fetch = func(context.Context, string) ([]serverObj.ServerObj, string, error) {
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

func TestAutomaticGroupUsesWholeCatalogWithoutMutatingIt(t *testing.T) {
	resetSubscription(t)
	raw, err := ResolveURL("vless://00000000-0000-0000-0000-000000000001@example.com:443?type=raw&security=none#raw")
	if err != nil {
		t.Fatal(err)
	}
	grpc, err := ResolveURL("vless://00000000-0000-0000-0000-000000000002@example.net:443?type=grpc&security=none#grpc")
	if err != nil {
		t.Fatal(err)
	}
	sub := configure.GetSubscription(0)
	sub.Servers = []configure.ServerRaw{{ServerObj: raw}, {ServerObj: grpc}}
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	serverIndex := configure.GetLenServers()
	standalone := &configure.ServerRaw{ServerObj: testServer(t, 10009)}
	if err := configure.AppendServers([]*configure.ServerRaw{standalone}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveServers([]int{serverIndex}) })
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd, setting.ProbeInterval = true, "300s"
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	wantRaw := configure.GetSubscription(0).Servers[0].ServerObj.ExportToURL()
	wantGRPC := configure.GetSubscription(0).Servers[1].ServerObj.ExportToURL()

	a := newAutomation()
	now := time.Unix(2000, 0)
	a.now = func() time.Time { return now }
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		for _, node := range nodes {
			if v, ok := node.(*serverObj.V2Ray); ok {
				v.Net = "tcp"
				v.Path = "GunService"
			}
		}
		return make([]subscriptionProbeResult, len(nodes))
	}
	var applied []configure.NodeRef
	a.applyGroup = func(_ string, refs []configure.NodeRef) error {
		applied = append([]configure.NodeRef(nil), refs...)
		return nil
	}
	a.step(context.Background())
	if len(applied) != 3 {
		t.Fatalf("automatic group received %d members; want standalone and two subscription nodes", len(applied))
	}
	stored := configure.GetSubscription(0)
	if got := stored.Servers[0].ServerObj.ExportToURL(); got != wantRaw {
		t.Fatalf("raw probe mutated catalog: %q != %q", got, wantRaw)
	}
	if got := stored.Servers[1].ServerObj.ExportToURL(); got != wantGRPC {
		t.Fatalf("gRPC probe mutated catalog: %q != %q", got, wantGRPC)
	}
	if got := a.groups["proxy"].next.Sub(now); got != 300*time.Second {
		t.Fatalf("next group check = %s; want 300s", got)
	}
}

func TestAutomaticGroupApplyFailureUsesBackoff(t *testing.T) {
	resetSubscription(t)
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd, setting.ProbeInterval = true, "1s"
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	now := time.Unix(3000, 0)
	a.now = func() time.Time { return now }
	probes := 0
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		probes++
		return make([]subscriptionProbeResult, len(nodes))
	}
	a.applyGroup = func(string, []configure.NodeRef) error {
		now = now.Add(5 * time.Minute)
		return errors.New("injected slow apply failure")
	}
	a.step(context.Background())
	now = now.Add(29 * time.Second)
	a.step(context.Background())
	if probes != 1 {
		t.Fatalf("apply failure retried early: %d probes", probes)
	}
	now = now.Add(time.Second)
	a.step(context.Background())
	if probes != 2 {
		t.Fatalf("apply failure did not retry after backoff: %d probes", probes)
	}
}

func TestAutomaticGroupSlowProbeFailureUsesCompletionTime(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		a := newAutomation()
		now := time.Unix(5000, 0)
		a.now = func() time.Time { return now }
		ctx, cancel := context.WithCancel(context.Background())
		setting := configure.DefaultOutboundSetting()
		setting.AutoAdd, setting.ProbeInterval = true, "300s"
		a.processGroup(ctx, "proxy", setting, nil, func([]serverObj.ServerObj, string) ([]subscriptionProbeResult, error) {
			now = now.Add(10 * time.Minute)
			if cancelled {
				cancel()
				return nil, ctx.Err()
			}
			return nil, errors.New("probe failed")
		})
		cancel()
		if got := a.groups["proxy"].next.Sub(now); got != 300*time.Second {
			t.Fatalf("cancelled=%v: backoff after failure=%s, want 300s", cancelled, got)
		}
	}
}

func TestValidateOutboundSettingAcceptsSupportedStrategies(t *testing.T) {
	for _, strategy := range []configure.ObservatoryType{
		configure.LeastPing,
		configure.KeepCurrent,
		configure.RoundRobin,
		configure.Random,
	} {
		setting := configure.DefaultOutboundSetting()
		setting.Type = strategy
		if err := ValidateOutboundSetting(setting); err != nil {
			t.Fatalf("strategy %q was rejected: %v", strategy, err)
		}
	}
	setting := configure.DefaultOutboundSetting()
	setting.Type = "unsupported"
	if err := ValidateOutboundSetting(setting); err == nil {
		t.Fatal("unsupported strategy was accepted")
	}
}

func prepareManualStickyGroup(t *testing.T) []serverObj.ServerObj {
	t.Helper()
	resetSubscription(t)
	nodes := []serverObj.ServerObj{testServer(t, 12001), testServer(t, 12002)}
	sub := configure.GetSubscription(0)
	sub.Servers = []configure.ServerRaw{{ServerObj: nodes[0]}, {ServerObj: nodes[1]}}
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	for id := range nodes {
		if err := configure.AddConnect(configure.NodeRef{
			TYPE: configure.SubscriptionServerType, Sub: 0, ID: id + 1, Outbound: "proxy",
		}); err != nil {
			t.Fatal(err)
		}
	}
	setting := configure.DefaultOutboundSetting()
	setting.Type, setting.ProbeInterval = configure.KeepCurrent, "300s"
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	return nodes
}

func TestKeepCurrentChangesOnlyAfterFailureAndFailsClosed(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	healthy := map[string]bool{
		nodes[0].ExportToURL(): true,
		nodes[1].ExportToURL(): true,
	}
	now := time.Unix(6000, 0)
	a := newAutomation()
	a.now = func() time.Time { return now }
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		results := make([]subscriptionProbeResult, len(candidates))
		for i, candidate := range candidates {
			if !healthy[candidate.ExportToURL()] {
				results[i].err = errors.New("unavailable")
			}
		}
		return results
	}
	updates := 0
	a.applyState = func(outbound string, next configure.OutboundSetting, _ []configure.NodeRef, replaceMembers bool) error {
		updates++
		if outbound != "proxy" || replaceMembers {
			t.Fatalf("manual sticky update used outbound=%q replaceMembers=%v", outbound, replaceMembers)
		}
		return configure.SetOutboundSetting(outbound, next)
	}

	a.step(context.Background())
	first := configure.NodeFingerprint(nodes[0].ExportToURL())
	second := configure.NodeFingerprint(nodes[1].ExportToURL())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != first {
		t.Fatalf("initial current = %q; want first healthy node %q", got, first)
	}

	// A healthy current node is retained even when another candidate is healthy.
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != first || updates != 1 {
		t.Fatalf("healthy current changed: current=%q updates=%d", got, updates)
	}

	// Failure moves to the first healthy backup in stable group order.
	healthy[nodes[0].ExportToURL()] = false
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != second {
		t.Fatalf("failed current did not switch to backup: got %q, want %q", got, second)
	}

	// Recovery of the old node must not take ownership back from a healthy current.
	healthy[nodes[0].ExportToURL()] = true
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != second || updates != 2 {
		t.Fatalf("recovered old node reclaimed traffic: current=%q updates=%d", got, updates)
	}

	// No healthy candidate clears the internal selection. Config generation then
	// emits the group's blackhole instead of allowing direct fallback.
	healthy[nodes[0].ExportToURL()] = false
	healthy[nodes[1].ExportToURL()] = false
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != "" {
		t.Fatalf("all-failed group retained current %q", got)
	}
	if got := configure.GetConnectedServersByOutbound("proxy").Len(); got != 2 {
		t.Fatalf("manual membership changed during health checks: %d members", got)
	}
}

func TestKeepCurrentManualPinOverridesWorkerChoice(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.Selected = nodes[0].ExportToURL()
	setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		results := make([]subscriptionProbeResult, len(candidates))
		results[1].err = errors.New("unavailable")
		return results
	}
	a.applyState = func(string, configure.OutboundSetting, []configure.NodeRef, bool) error {
		t.Fatal("worker changed sticky state while a manual pin was active")
		return nil
	}
	a.step(context.Background())
	if got := configure.GetOutboundSetting("proxy"); got.Selected != setting.Selected || got.StickyCurrent != setting.StickyCurrent {
		t.Fatalf("manual pin or internal state changed: %+v", got)
	}
}

func TestAutomaticKeepCurrentAppliesMembershipAndChoiceTogether(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		results := make([]subscriptionProbeResult, len(candidates))
		for i, candidate := range candidates {
			if candidate.ExportToURL() == nodes[0].ExportToURL() {
				results[i].err = errors.New("unavailable")
			}
		}
		return results
	}
	calls := 0
	a.applyState = func(outbound string, next configure.OutboundSetting, members []configure.NodeRef, replaceMembers bool) error {
		calls++
		if outbound != "proxy" || !replaceMembers {
			t.Fatalf("automatic sticky update used outbound=%q replaceMembers=%v", outbound, replaceMembers)
		}
		if len(members) != 1 || members[0].ID != 2 {
			t.Fatalf("healthy membership = %+v; want only second node", members)
		}
		if want := configure.NodeFingerprint(nodes[1].ExportToURL()); next.StickyCurrent != want {
			t.Fatalf("sticky current = %q; want %q", next.StickyCurrent, want)
		}
		return nil
	}
	a.applyGroup = func(string, []configure.NodeRef) error {
		t.Fatal("membership and sticky choice were applied in separate reloads")
		return nil
	}
	a.step(context.Background())
	if calls != 1 {
		t.Fatalf("combined state updates = %d; want 1", calls)
	}
}
