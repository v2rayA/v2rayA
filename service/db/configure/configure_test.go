package configure

import (
	"os"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "v2raya-configure-test-*")
	if err != nil {
		panic(err)
	}
	conf.GetEnvironmentConfig().Config = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestGetLenSubscriptionServersPanicsOnMissingSubscription(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("missing subscription must report a recoverable panic")
		}
	}()
	GetLenSubscriptionServers(-1)
}

func TestSetSubscriptionAndConnectsRollsBackConnections(t *testing.T) {
	const outbound = "transaction-test"
	if err := AddOutbound(outbound); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveOutbound(outbound) })

	index := GetLenSubscriptions()
	subscription := &SubscriptionRaw{Address: "https://example.test/subscription"}
	if err := AppendSubscriptions([]*SubscriptionRaw{subscription}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveSubscriptions([]int{index}) })

	initial := &NodeRefs{Touches: []*NodeRef{{
		TYPE:     SubscriptionServerType,
		ID:       1,
		Sub:      index,
		Outbound: outbound,
	}}}
	if err := OverwriteConnects(initial); err != nil {
		t.Fatal(err)
	}
	replacement := &NodeRefs{Touches: []*NodeRef{{
		TYPE:     SubscriptionServerType,
		ID:       2,
		Sub:      index,
		Outbound: outbound,
	}}}
	if err := SetSubscriptionAndConnects(index+1, subscription, replacement); err == nil {
		t.Fatal("updating a missing subscription succeeded")
	}

	got := GetConnectedServersByOutbound(outbound)
	if got == nil || got.Len() != 1 || *got.Get()[0] != *initial.Get()[0] {
		t.Fatalf("connections changed after rollback: %+v", got)
	}

	if err := SetSubscriptionAndConnects(index, subscription, replacement); err != nil {
		t.Fatalf("replace subscription and connections: %v", err)
	}
	got = GetConnectedServersByOutbound(outbound)
	if got == nil || got.Len() != 1 || *got.Get()[0] != *replacement.Get()[0] {
		t.Fatalf("connections were not committed with subscription: %+v", got)
	}
}
