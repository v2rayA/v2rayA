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
	setting.Type = configure.RoundRobin
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
	setting.Type = configure.RoundRobin
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

func TestForcedAutomaticGroupRefreshIgnoresFutureSchedule(t *testing.T) {
	resetSubscription(t)
	setting := configure.DefaultOutboundSetting()
	setting.Type = configure.RoundRobin
	setting.AutoAdd, setting.ProbeInterval = true, "300s"
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	a.now = func() time.Time { return time.Unix(4000, 0) }
	probes := 0
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		probes++
		return make([]subscriptionProbeResult, len(nodes))
	}
	a.applyGroup = func(string, []configure.NodeRef) error { return nil }

	a.step(context.Background())
	if probes != 1 {
		t.Fatalf("initial probes = %d; want 1", probes)
	}
	if err := a.forceGroups(context.Background(), "proxy"); err != nil {
		t.Fatal(err)
	}
	if probes != 2 {
		t.Fatalf("forced refresh probes = %d; want 2", probes)
	}
}

func TestForcedGroupRefreshDoesNotWaitForSubscriptionDownload(t *testing.T) {
	resetSubscription(t)
	sub := configure.GetSubscription(0)
	sub.UpdateMode = configure.SubscriptionUpdateAtInterval
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	a.fetch = func(context.Context, string, bool) ([]serverObj.ServerObj, string, error) {
		t.Fatal("manual group refresh downloaded a subscription")
		return nil, "", nil
	}
	probes := 0
	a.probe = func(_ context.Context, nodes []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		probes++
		return make([]subscriptionProbeResult, len(nodes))
	}
	a.applyGroup = func(string, []configure.NodeRef) error { return nil }
	if err := a.forceGroups(context.Background(), "proxy"); err != nil {
		t.Fatal(err)
	}
	if probes != 1 {
		t.Fatalf("manual group refresh made %d probe passes, want one", probes)
	}
}

func TestSubscriptionCatalogChangeWakesAutomaticGroupsWithoutExistingMembers(t *testing.T) {
	resetSubscription(t)
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	for {
		select {
		case <-automationWake:
		default:
			goto drained
		}
	}
drained:
	old := configure.GetSubscription(0)
	if err := storeSubscriptionUpdate(0, old, []serverObj.ServerObj{testServer(t, 12403)}, "", false); err != nil {
		t.Fatal(err)
	}
	select {
	case <-automationWake:
	default:
		t.Fatal("new catalog server did not wake automatic group checks")
	}
}

func TestAutomaticGroupSlowProbeFailureUsesCompletionTime(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		a := newAutomation()
		now := time.Unix(5000, 0)
		a.now = func() time.Time { return now }
		ctx, cancel := context.WithCancel(context.Background())
		setting := configure.DefaultOutboundSetting()
		setting.Type = configure.RoundRobin
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
		configure.FirstAvailable,
	} {
		setting := configure.DefaultOutboundSetting()
		setting.Type = strategy
		if err := ValidateOutboundSetting(setting); err != nil {
			t.Fatalf("strategy %q was rejected: %v", strategy, err)
		}
	}
	fixed := configure.DefaultOutboundSetting()
	fixed.Type = configure.Fixed
	fixed.Selected = "socks5://server.example:1080"
	if err := ValidateOutboundSetting(fixed); err != nil {
		t.Fatalf("fixed strategy was rejected: %v", err)
	}
	fixed.Selected = ""
	if err := ValidateOutboundSetting(fixed); err == nil {
		t.Fatal("fixed strategy without a selected server was accepted")
	}
	setting := configure.DefaultOutboundSetting()
	setting.Type = "unsupported"
	if err := ValidateOutboundSetting(setting); err == nil {
		t.Fatal("unsupported strategy was accepted")
	}
}

func TestRandomStrategyUsesLowestNonEmptyLatencyBucket(t *testing.T) {
	nodes := []serverObj.ServerObj{
		testServer(t, 13001),
		testServer(t, 13002),
		testServer(t, 13003),
		testServer(t, 13004),
	}
	candidates := make([]groupCandidate, len(nodes))
	for i, node := range nodes {
		candidates[i] = groupCandidate{node: node}
	}
	results := []subscriptionProbeResult{
		{latency: 250 * time.Millisecond},
		{latency: 100 * time.Millisecond},
		{latency: 200 * time.Millisecond},
		{latency: 800 * time.Millisecond},
	}
	got := selectedGroupCandidate(configure.Random, "", candidates, results, func(size int) int {
		if size != 2 {
			t.Fatalf("eligible random bucket size = %d; want 2", size)
		}
		return 1
	})
	if want := configure.NodeFingerprint(nodes[2].ExportToURL()); got != want {
		t.Fatalf("random choice = %q; want second member of <250 ms bucket %q", got, want)
	}

	results = []subscriptionProbeResult{
		{latency: 300 * time.Millisecond},
		{latency: 450 * time.Millisecond},
		{latency: 600 * time.Millisecond},
		{err: errors.New("timeout")},
	}
	got = selectedGroupCandidate(configure.Random, "", candidates, results, func(size int) int {
		if size != 2 {
			t.Fatalf("eligible fallback bucket size = %d; want 2", size)
		}
		return 0
	})
	if want := configure.NodeFingerprint(nodes[0].ExportToURL()); got != want {
		t.Fatalf("random fallback choice = %q; want member of <500 ms bucket %q", got, want)
	}
}

func TestRandomStrategyKeepsHealthySelectionBetweenChecks(t *testing.T) {
	nodes := []serverObj.ServerObj{testServer(t, 13011), testServer(t, 13012)}
	candidates := []groupCandidate{{node: nodes[0]}, {node: nodes[1]}}
	current := configure.NodeFingerprint(nodes[1].ExportToURL())
	results := []subscriptionProbeResult{
		{latency: 10 * time.Millisecond, throughput: 200 * 1024, speedMeasured: true},
		{latency: 80 * time.Millisecond, throughput: 200 * 1024, speedMeasured: true},
	}
	got := selectedGroupCandidate(configure.Random, current, candidates, results, func(int) int {
		t.Fatal("healthy random selection was redrawn")
		return 0
	})
	if got != current {
		t.Fatalf("healthy random selection changed from %q to %q", current, got)
	}
	results[1].err = errors.New("URL check failed")
	got = selectedGroupCandidate(configure.Random, current, candidates, results, func(size int) int {
		if size != 1 {
			t.Fatalf("failover chooser saw %d candidates, want one", size)
		}
		return 0
	})
	if want := configure.NodeFingerprint(nodes[0].ExportToURL()); got != want {
		t.Fatalf("failed random selection = %q, want %q", got, want)
	}
}

func TestLeastPingIgnoresSlowLowLatencyCandidate(t *testing.T) {
	nodes := []serverObj.ServerObj{testServer(t, 13201), testServer(t, 13202), testServer(t, 13203)}
	candidates := make([]groupCandidate, len(nodes))
	for i, node := range nodes {
		candidates[i] = groupCandidate{node: node}
	}
	results := []subscriptionProbeResult{
		{latency: 10 * time.Millisecond, throughput: 50 * 1024, speedMeasured: true},
		{latency: 80 * time.Millisecond, throughput: 140 * 1024, speedMeasured: true},
		{latency: 120 * time.Millisecond, throughput: 180 * 1024, speedMeasured: true},
	}
	got := selectedGroupCandidate(configure.LeastPing, "", candidates, results, func(int) int {
		t.Fatal("least-ping called the random chooser")
		return 0
	})
	if want := configure.NodeFingerprint(nodes[1].ExportToURL()); got != want {
		t.Fatalf("least-ping choice = %q; want lowest latency above 100 KiB/s %q", got, want)
	}
}

func TestAllSlowCandidatesFallBackToFastest(t *testing.T) {
	nodes := []serverObj.ServerObj{testServer(t, 13301), testServer(t, 13302), testServer(t, 13303)}
	candidates := make([]groupCandidate, len(nodes))
	for i, node := range nodes {
		candidates[i] = groupCandidate{node: node}
	}
	results := []subscriptionProbeResult{
		{latency: 10 * time.Millisecond, throughput: 20 * 1024, speedMeasured: true},
		{latency: 100 * time.Millisecond, throughput: 90 * 1024, speedMeasured: true},
		{latency: 50 * time.Millisecond, throughput: 60 * 1024, speedMeasured: true},
	}
	current := configure.NodeFingerprint(nodes[0].ExportToURL())
	got := selectedGroupCandidate(configure.KeepCurrent, current, candidates, results, func(int) int { return 0 })
	if want := configure.NodeFingerprint(nodes[1].ExportToURL()); got != want {
		t.Fatalf("all-slow fallback = %q; want fastest candidate %q", got, want)
	}
	eligible := eligibleProbeResults(results)
	if eligible[0] || !eligible[1] || eligible[2] {
		t.Fatalf("all-slow eligible set = %+v; want only fastest", eligible)
	}
}

func TestFirstAvailableUsesStableGroupOrder(t *testing.T) {
	nodes := []serverObj.ServerObj{testServer(t, 13101), testServer(t, 13102), testServer(t, 13103)}
	candidates := make([]groupCandidate, len(nodes))
	for i, node := range nodes {
		candidates[i] = groupCandidate{node: node}
	}
	results := []subscriptionProbeResult{
		{err: errors.New("unavailable")},
		{latency: 900 * time.Millisecond},
		{latency: 50 * time.Millisecond},
	}
	got := selectedGroupCandidate(configure.FirstAvailable, "", candidates, results, func(int) int {
		t.Fatal("first-available called the random chooser")
		return 0
	})
	if want := configure.NodeFingerprint(nodes[1].ExportToURL()); got != want {
		t.Fatalf("first available = %q; want stable second member %q", got, want)
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

func TestKeepCurrentConfirmsLowSpeedBeforeFailover(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.AutoAdd = true
	setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(7000, 0)
	a := newAutomation()
	a.now = func() time.Time { return now }
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		if len(candidates) != 2 {
			t.Fatalf("got %d candidates, want two", len(candidates))
		}
		return []subscriptionProbeResult{
			{throughput: 200 * 1024, speedMeasured: true},
			{throughput: 50 * 1024, speedMeasured: true},
		}
	}
	updates := 0
	a.applyState = func(_ string, next configure.OutboundSetting, members []configure.NodeRef, _ bool) error {
		updates++
		if len(members) != 1 || next.StickyCurrent != configure.NodeFingerprint(nodes[0].ExportToURL()) {
			t.Fatalf("confirmed slow-node failover = %+v, members=%+v", next, members)
		}
		return configure.SetOutboundSetting("proxy", next)
	}
	a.applyMembers = func(string, []configure.NodeRef) error {
		t.Fatal("one slow sample changed group membership")
		return nil
	}
	a.step(context.Background())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != setting.StickyCurrent || updates != 0 {
		t.Fatalf("single slow sample switched current: %q, updates=%d", got, updates)
	}
	now = now.Add(300 * time.Second)
	a.step(context.Background())
	if updates != 1 {
		t.Fatalf("confirmed low speed caused %d failovers, want one", updates)
	}
}

func TestKeepCurrentMembershipChangeDoesNotReloadSelectedNode(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.AutoAdd = true
	setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	a.probe = func(_ context.Context, _ []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		return []subscriptionProbeResult{{err: errors.New("backup unavailable")}, {throughput: 200 * 1024, speedMeasured: true}}
	}
	a.applyState = func(string, configure.OutboundSetting, []configure.NodeRef, bool) error {
		t.Fatal("unchanged current node restarted the core")
		return nil
	}
	a.applyGroup = func(string, []configure.NodeRef) error {
		t.Fatal("unchanged current node restarted the group")
		return nil
	}
	changes := 0
	a.applyMembers = func(_ string, members []configure.NodeRef) error {
		changes++
		if len(members) != 1 || members[0].ID != 2 {
			t.Fatalf("membership = %+v, want only the current second node", members)
		}
		return nil
	}
	a.step(context.Background())
	if changes != 1 {
		t.Fatalf("membership-only changes = %d, want one", changes)
	}
}

func TestAutomaticStrategiesKeepTheRunningNodeWhenOnlyMembershipChanges(t *testing.T) {
	for _, strategy := range []configure.ObservatoryType{
		configure.LeastPing, configure.KeepCurrent, configure.Random, configure.FirstAvailable,
	} {
		t.Run(string(strategy), func(t *testing.T) {
			nodes := prepareManualStickyGroup(t)
			setting := configure.GetOutboundSetting("proxy")
			setting.Type = strategy
			setting.AutoAdd = true
			setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
			if err := configure.SetOutboundSetting("proxy", setting); err != nil {
				t.Fatal(err)
			}
			now := time.Unix(8000, 0)
			a := newAutomation()
			a.now = func() time.Time { return now }
			a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
				if len(candidates) != 2 {
					t.Fatalf("got %d catalog candidates, want two", len(candidates))
				}
				return []subscriptionProbeResult{
					{err: errors.New("backup unavailable")},
					{latency: 50 * time.Millisecond, throughput: 200 * 1024, speedMeasured: true},
				}
			}
			a.applyState = func(string, configure.OutboundSetting, []configure.NodeRef, bool) error {
				t.Fatal("membership-only refresh restarted the running core")
				return nil
			}
			a.applyGroup = func(string, []configure.NodeRef) error {
				t.Fatal("membership-only refresh reloaded the group")
				return nil
			}
			changes := 0
			a.applyMembers = func(outbound string, members []configure.NodeRef) error {
				changes++
				return writeGroupMembers(outbound, members)
			}
			a.step(context.Background())
			now = now.Add(300 * time.Second)
			a.step(context.Background())
			if changes != 1 || configure.GetOutboundSetting("proxy").StickyCurrent != setting.StickyCurrent {
				t.Fatalf("membership changes=%d, current=%q; want one write and retained node", changes, configure.GetOutboundSetting("proxy").StickyCurrent)
			}
		})
	}
}

func TestLegacyManualPinBecomesFixedAndStopsWorker(t *testing.T) {
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
	if got := configure.GetOutboundSetting("proxy"); got.Type != configure.Fixed || got.Selected != setting.Selected || got.StickyCurrent != "" {
		t.Fatalf("legacy manual pin was not normalized to fixed: %+v", got)
	}
}

func TestRoundRobinWorkerFiltersSlowMembers(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.Type, setting.StickyCurrent = configure.RoundRobin, ""
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	a := newAutomation()
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		return []subscriptionProbeResult{
			{latency: 10 * time.Millisecond, throughput: 40 * 1024, speedMeasured: true},
			{latency: 80 * time.Millisecond, throughput: 160 * 1024, speedMeasured: true},
		}
	}
	calls := 0
	a.applyState = func(outbound string, next configure.OutboundSetting, _ []configure.NodeRef, replaceMembers bool) error {
		calls++
		want := configure.NodeFingerprint(nodes[1].ExportToURL())
		if outbound != "proxy" || replaceMembers || next.EligibleMembers != want {
			t.Fatalf("round-robin filter update = outbound %q replace=%v eligible=%v", outbound, replaceMembers, next.EligibleMembers)
		}
		return configure.SetOutboundSetting(outbound, next)
	}
	a.step(context.Background())
	if calls != 1 {
		t.Fatalf("round-robin filter updates = %d; want 1", calls)
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
