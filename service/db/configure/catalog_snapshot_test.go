package configure

import (
	"reflect"
	"testing"

	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestCatalogSnapshotRestoresDeletionAndRenumbering(t *testing.T) {
	const group = "catalog-snapshot"
	if err := AddOutbound(group); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveOutbound(group) })
	serverIndex, subIndex := GetLenServers(), GetLenSubscriptions()
	node := ServerRaw{ServerObj: &serverObj.SOCKS{Server: "192.0.2.1", Port: 1080, Protocol: "socks5"}}
	if err := AppendServers([]*ServerRaw{&node, &node}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveServers([]int{serverIndex + 1, serverIndex}) })
	sub := &SubscriptionRaw{Address: "https://example.invalid", Servers: []ServerRaw{node}}
	if err := AppendSubscriptions([]*SubscriptionRaw{sub, sub}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveSubscriptions([]int{subIndex + 1, subIndex}) })
	refs := NewNodeRefs([]*NodeRef{
		{TYPE: ServerType, ID: serverIndex + 2, Outbound: group},
		{TYPE: SubscriptionServerType, ID: 1, Sub: subIndex + 1, Outbound: group},
	})
	if err := SetOutboundSettingAndMembers(group, DefaultOutboundSetting(), refs); err != nil {
		t.Fatal(err)
	}
	servers, subscriptions := GetServers(), GetSubscriptions()
	restore, err := db.SnapshotCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if err := RemoveNodes([]int{subIndex}, []int{serverIndex}, NewNodeRefs(nil)); err != nil {
		t.Fatal(err)
	}
	if err := db.Set("system", "snapshot-test", "keep"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Delete("system", "snapshot-test") })
	if err := restore(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(servers, GetServers()) || !reflect.DeepEqual(subscriptions, GetSubscriptions()) ||
		!reflect.DeepEqual(refs, GetConnectedServersByOutbound(group)) {
		t.Fatal("catalog IDs, ordering or references changed after restore")
	}
	var unrelated string
	if err := db.Get("system", "snapshot-test", &unrelated); err != nil || unrelated != "keep" {
		t.Fatal("catalog restore changed unrelated settings", err)
	}
}
