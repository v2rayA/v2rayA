package service

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestProbeCoreSlotSerializesAndCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if err := acquireProbeCore(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := acquireProbeCore(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting cancellation: %v", err)
	}
	releaseProbeCore()
	var active, peak atomic.Int32
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := acquireProbeCore(context.Background()); err != nil {
				t.Error(err)
				return
			}
			n := active.Add(1)
			if n > peak.Load() {
				peak.Store(n)
			}
			time.Sleep(time.Millisecond)
			active.Add(-1)
			releaseProbeCore()
		}()
	}
	wg.Wait()
	if peak.Load() != 1 {
		t.Fatalf("concurrent probe owners: %d", peak.Load())
	}
}

func TestSelectionOrdersPingThenStopsAtFirstFastNode(t *testing.T) {
	for _, strategy := range []configure.ObservatoryType{configure.LeastPing, configure.KeepCurrent, configure.Random, configure.RoundRobin} {
		t.Run(string(strategy), func(t *testing.T) {
			a := newTestAutomation()
			nodes := []serverObj.ServerObj{testServer(t, 14001), testServer(t, 14002), testServer(t, 14003)}
			candidates := []groupCandidate{{node: nodes[0]}, {node: nodes[1]}, {node: nodes[2]}}
			a.ping = func(context.Context, []serverObj.ServerObj) []subscriptionProbeResult {
				return []subscriptionProbeResult{{latency: 30 * time.Millisecond}, {latency: 10 * time.Millisecond}, {latency: 20 * time.Millisecond}}
			}
			a.choose = func(n int) int { return n - 1 }
			setting := configure.DefaultOutboundSetting()
			setting.Type = strategy
			var order []int
			probe := func(nodes []serverObj.ServerObj, _ string) ([]subscriptionProbeResult, error) {
				if len(nodes) != 1 {
					t.Fatal("more than one candidate per core check")
				}
				port := nodes[0].GetPort()
				order = append(order, port)
				r := healthyTestResults(1)
				if port == 14002 {
					r[0].throughput = 1
				}
				return r, nil
			}
			results, err := a.selectGroup(context.Background(), setting, candidates, probe)
			if err != nil {
				t.Fatal(err)
			}
			want := []int{14002, 14003}
			if strategy == configure.RoundRobin {
				want = append(want, 14001)
			}
			if !reflect.DeepEqual(order, want) {
				t.Fatalf("order %v, want %v", order, want)
			}
			if !eligibleProbeResults(results)[2] || eligibleProbeResults(results)[1] {
				t.Fatal("speed gating failed")
			}
		})
	}
}

func TestKeepCurrentRetainsReachableSlowCurrent(t *testing.T) {
	a := newTestAutomation()
	nodes := []serverObj.ServerObj{testServer(t, 14101), testServer(t, 14102)}
	candidates := []groupCandidate{{node: nodes[0]}, {node: nodes[1]}}
	setting := configure.DefaultOutboundSetting()
	setting.Type = configure.KeepCurrent
	setting.StickyCurrent = configure.NodeFingerprint(nodes[1].ExportToURL())
	pings := 0
	a.ping = func(context.Context, []serverObj.ServerObj) []subscriptionProbeResult {
		pings++
		return healthyTestResults(2)
	}
	slow := false
	var order []int
	probe := func(nodes []serverObj.ServerObj, _ string) ([]subscriptionProbeResult, error) {
		order = append(order, nodes[0].GetPort())
		r := healthyTestResults(1)
		if slow && nodes[0].GetPort() == 14102 {
			r[0].throughput = 1
		}
		return r, nil
	}
	if _, err := a.selectGroup(context.Background(), setting, candidates, probe); err != nil {
		t.Fatal(err)
	}
	if pings != 0 || !reflect.DeepEqual(order, []int{14102}) {
		t.Fatalf("healthy current triggered extra work: %v %d", order, pings)
	}
	slow = true
	order = nil
	if _, err := a.selectGroup(context.Background(), setting, candidates, probe); err != nil {
		t.Fatal(err)
	}
	if pings != 0 || !reflect.DeepEqual(order, []int{14102}) {
		t.Fatalf("slow failover: %v %d", order, pings)
	}
}

func TestMembershipIncludesDeadNodesWithoutAnyProbes(t *testing.T) {
	prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	a := newTestAutomation()
	a.ping = func(context.Context, []serverObj.ServerObj) []subscriptionProbeResult {
		t.Fatal("membership pinged")
		return nil
	}
	a.probe = func(context.Context, []serverObj.ServerObj, string) []subscriptionProbeResult {
		t.Fatal("membership launched core")
		return nil
	}
	if err := a.forceGroups(context.Background(), "proxy"); err != nil {
		t.Fatal(err)
	}
	if n := configure.GetConnectedServersByOutbound("proxy").Len(); n != 2 {
		t.Fatalf("membership=%d", n)
	}
}

func TestUnknownOrLowSpeedNeverBecomesEligible(t *testing.T) {
	r := []subscriptionProbeResult{{}, {speedMeasured: true, throughput: 1}, {speedMeasured: true, throughput: subscriptionMinSpeed}}
	if got := eligibleProbeResults(r); !reflect.DeepEqual(got, []bool{false, false, true}) {
		t.Fatal(got)
	}
}

func TestSubscriptionRecoveryStopsAfterFirstReachableNode(t *testing.T) {
	nodes := prepareManualStickyGroup(t)
	sub := configure.GetSubscription(0)
	a := newTestAutomation()
	calls := 0
	a.probe = func(_ context.Context, candidates []serverObj.ServerObj, _ string) []subscriptionProbeResult {
		calls++
		if len(candidates) != 1 || candidates[0].ExportToURL() != nodes[0].ExportToURL() {
			t.Fatal("recovery did not probe sequentially from the first candidate")
		}
		return healthyTestResults(1)
	}
	unavailable, err := a.subscriptionUnavailable(context.Background(), sub, HttpTestURL, &subscriptionSchedule{})
	if err != nil || unavailable || calls != 1 {
		t.Fatalf("unavailable=%v calls=%d err=%v", unavailable, calls, err)
	}
}

func TestLegacyFirstAvailableNormalizesToLeastPing(t *testing.T) {
	prepareManualStickyGroup(t)
	setting := configure.GetOutboundSetting("proxy")
	setting.Type = "firstavailable"
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetOutboundSetting("proxy"); got.Type != configure.LeastPing {
		t.Fatalf("legacy strategy: %v", got.Type)
	}
	if err := ValidateOutboundSetting(setting); err == nil {
		t.Fatal("removed strategy accepted by new settings")
	}
}
