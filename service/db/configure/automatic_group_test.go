package configure

import (
	"reflect"
	"testing"

	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestAutomaticGroupUsesCatalogAndFreezesOnDisable(t *testing.T) {
	const group = "all-catalog-test"
	if err := AddOutbound(group); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveOutbound(group) })
	serverIndex, subIndex := GetLenServers(), GetLenSubscriptions()
	node := ServerRaw{ServerObj: &serverObj.SOCKS{Server: "192.0.2.1", Port: 1, Protocol: "socks5"}, Latency: "TIMEOUT"}
	if err := AppendServers([]*ServerRaw{&node}); err != nil {
		t.Fatal(err)
	}
	if err := AppendSubscriptions([]*SubscriptionRaw{{Address: "https://example.invalid/sub", Servers: []ServerRaw{node, node}}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveServers([]int{serverIndex}); _ = RemoveSubscriptions([]int{subIndex}) })
	setting := DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := SetOutboundSetting(group, setting); err != nil {
		t.Fatal(err)
	}
	refs := GetConnectedServersByOutbound(group)
	var found int
	for _, ref := range refs.Get() {
		if ref.Outbound != group {
			t.Fatalf("wrong group: %+v", ref)
		}
		if ref.TYPE == ServerType && ref.ID == serverIndex+1 || ref.TYPE == SubscriptionServerType && ref.Sub == subIndex {
			found++
		}
	}
	if found != 3 {
		t.Fatalf("unreachable or duplicate catalog entries were filtered: %d", found)
	}
	sub := GetSubscription(subIndex)
	sub.Servers = sub.Servers[:1]
	if err := SetSubscription(subIndex, sub); err != nil {
		t.Fatal(err)
	}
	if got := GetConnectedServersByOutbound(group).Len(); got != refs.Len()-1 {
		t.Fatalf("catalog removal not reflected: %d", got)
	}
	refs = GetConnectedServersByOutbound(group)
	setting.AutoAdd = false
	if err := SetOutboundSettingAndMembers(group, setting, refs); err != nil {
		t.Fatal(err)
	}
	sub.Servers = nil
	if err := SetSubscription(subIndex, sub); err != nil {
		t.Fatal(err)
	}
	if got := GetConnectedServersByOutbound(group); !reflect.DeepEqual(got, refs) {
		t.Fatalf("disabled group still follows catalog: %+v", got)
	}
}

func TestGroupSettingAndMembersCommitTogether(t *testing.T) {
	const group = "atomic-setting-test"
	if err := AddOutbound(group); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveOutbound(group) })
	previous := GetOutboundSetting(group)
	members := NewNodeRefs([]*NodeRef{{TYPE: ServerType, ID: 42, Outbound: group}})
	if err := OverwriteConnects(members); err != nil {
		t.Fatal(err)
	}
	_, err := db.GetDB().Exec(`CREATE TRIGGER reject_group_setting BEFORE INSERT ON system_config WHEN NEW.key = 'outbound.atomic-setting-test:setting' BEGIN SELECT RAISE(ABORT, 'injected failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.GetDB().Exec("DROP TRIGGER reject_group_setting") })
	next := previous
	next.AutoAdd = true
	if err := SetOutboundSettingAndMembers(group, next, NewNodeRefs(nil)); err == nil {
		t.Fatal("injected write failure was ignored")
	}
	if GetOutboundSetting(group) != previous || !reflect.DeepEqual(GetConnectedServersByOutbound(group), members) {
		t.Fatal("partial setting/member commit")
	}
}
