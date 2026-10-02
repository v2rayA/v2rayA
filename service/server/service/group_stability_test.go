package service

import (
	"context"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestCatalogPermutationDoesNotInvalidateSelectionButMembershipChangeDoes(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	a := newTestAutomation()
	now := time.Unix(8000, 0)
	a.now = func() time.Time { return now }
	checks := 0
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		checks += len(candidates)
		return healthyTestResults(len(candidates))
	}
	setting := configure.GetOutboundSetting("proxy")
	setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
	_ = configure.SetOutboundSetting("proxy", setting)
	a.step(context.Background())
	before := configure.GetOutboundSetting("proxy")
	if before.StickyCurrent != setting.StickyCurrent {
		t.Fatal("initial current changed")
	}
	old := configure.GetSubscription(0)
	renamed, err := ResolveURL(nodes[1].ExportToURL())
	if err != nil {
		t.Fatal(err)
	}
	renamed.SetName("renamed")
	if err := storeSubscriptionUpdate(0, old, []serverObj.ServerObj{renamed, nodes[0]}, "", false); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	a.step(context.Background())
	if checks != 1 || configure.GetOutboundSetting("proxy").StickyCurrent != before.StickyCurrent {
		t.Fatal("reorder/rename restarted the selection schedule")
	}
	// Remove the non-active member. A true membership change must re-evaluate
	// even when the previous active member is still present and reachable.
	old = configure.GetSubscription(0)
	if err := storeSubscriptionUpdate(0, old, []serverObj.ServerObj{renamed}, "", true); err != nil {
		t.Fatal(err)
	}
	pings := 0
	a.ping = func(_ context.Context, n []serverObj.ServerObj) []subscriptionProbeResult {
		pings++
		return healthyTestResults(len(n))
	}
	a.step(context.Background())
	if pings != 1 || checks != 2 {
		t.Fatalf("membership change did not re-evaluate: pings=%d checks=%d", pings, checks)
	}
	if configure.GetOutboundSetting("proxy").Type != configure.KeepCurrent {
		t.Fatal("worker changed policy")
	}
}

func TestReachableCurrentWithUnknownSpeedStaysEligible(t *testing.T) {
	a := newTestAutomation()
	node := testServer(t, 14111)
	setting := configure.DefaultOutboundSetting()
	setting.Type, setting.StickyCurrent = configure.KeepCurrent, configure.NodeFingerprint(node.ExportToURL())
	result, err := a.selectGroup(context.Background(), setting, []groupCandidate{{node: node}}, func([]serverObj.ServerObj, string) ([]subscriptionProbeResult, error) {
		return []subscriptionProbeResult{{latency: time.Millisecond}}, nil
	})
	if err != nil || !eligibleProbeResults(result)[0] {
		t.Fatalf("unknown speed evicted reachable current: %v", err)
	}
}

func TestCatalogChangeBackBeforeWorkerStillInvalidatesCurrent(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	a := newTestAutomation()
	setting := configure.GetOutboundSetting("proxy")
	setting.AutoAdd = true
	setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
	setting.CatalogRevision = groupCatalogRevision(connectedGroupCandidates("proxy"))
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	extra := testServer(t, 14119)
	for _, next := range [][]serverObj.ServerObj{{nodes[0], nodes[1], extra}, nodes} {
		if err := storeSubscriptionUpdate(0, configure.GetSubscription(0), next, "", true); err != nil {
			t.Fatal(err)
		}
	}
	if !configure.GetOutboundSetting("proxy").SelectionInvalidated {
		t.Fatal("two catalog changes lost invalidation")
	}
	pings := 0
	a.ping = func(_ context.Context, n []serverObj.ServerObj) []subscriptionProbeResult {
		pings++
		return healthyTestResults(len(n))
	}
	a.probe = func(_ context.Context, n []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		return healthyTestResults(len(n))
	}
	a.step(context.Background())
	got := configure.GetOutboundSetting("proxy")
	if pings != 1 || got.SelectionInvalidated || got.Type != configure.KeepCurrent || got.StickyCurrent != configure.NodeFingerprint(nodes[0].ExportToURL()) {
		t.Fatalf("catalog re-evaluation failed: pings=%d setting=%+v", pings, got)
	}
}
