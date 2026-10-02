package service

import (
	"reflect"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestDeleteWhichMixedSubscriptionsDisconnects(t *testing.T) {
	server := &configure.ServerRaw{ServerObj: &serverObj.SOCKS{Server: "192.0.2.1", Port: 1080, Protocol: "socks5"}}
	if err := configure.AppendServers([]*configure.ServerRaw{server}); err != nil {
		t.Fatal(err)
	}
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{
		{Address: "https://example.com/one", Servers: []configure.ServerRaw{*server}},
		{Address: "https://example.com/two", Servers: []configure.ServerRaw{*server}},
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		configure.ClearConnects("proxy")
		configure.RemoveServers([]int{0})
		configure.RemoveSubscriptions([]int{0, 1})
	})
	if err := configure.AddConnect(configure.NodeRef{TYPE: configure.SubscriptionServerType, ID: 1, Sub: 1}); err != nil {
		t.Fatal(err)
	}
	err := DeleteWhich([]*configure.Which{
		{NodeRef: configure.NodeRef{TYPE: configure.ServerType, ID: 1}},
		{NodeRef: configure.NodeRef{TYPE: configure.SubscriptionType, ID: 1}},
		{NodeRef: configure.NodeRef{TYPE: configure.SubscriptionType, ID: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := configure.GetConnectedServers(); got.Len() != 0 {
		t.Fatalf("deleted subscription retained %d connections: %+v", got.Len(), got.Get()[0])
	}
	if n := configure.GetLenSubscriptions(); n != 0 {
		t.Fatalf("retained %d subscriptions", n)
	}
}

func TestDeleteWhichRenumbersSurvivingMembersAcrossGroups(t *testing.T) {
	resetSubscription(t)
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{Address: "https://second.invalid", Servers: []configure.ServerRaw{{ServerObj: testServer(t, 10002)}}}}); err != nil {
		t.Fatal(err)
	}
	if err := configure.AddOutbound("deletion-test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveOutbound("deletion-test") })
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	for _, out := range []string{"proxy", "deletion-test"} {
		if err := configure.AddConnect(configure.NodeRef{TYPE: configure.SubscriptionServerType, ID: 1, Sub: 1, Outbound: out}); err != nil {
			t.Fatal(err)
		}
	}
	if err := DeleteWhich([]*configure.Which{{NodeRef: configure.NodeRef{TYPE: configure.SubscriptionType, ID: 1}}}); err != nil {
		t.Fatal(err)
	}
	for _, out := range []string{"proxy", "deletion-test"} {
		got := configure.GetConnectedServersByOutbound(out).Get()
		if len(got) != 1 || got[0].Sub != 0 || got[0].ID != 1 {
			t.Fatalf("%s references after deletion: %+v", out, got)
		}
		server, err := got[0].LocateServerRaw()
		if err != nil || server.ServerObj.ExportToURL() != testServer(t, 10002).ExportToURL() {
			t.Fatalf("%s reference resolved to the wrong remaining server, err=%v", out, err)
		}
	}
}

func TestDeleteWhichFailureKeepsCatalogAndMembersTogether(t *testing.T) {
	resetSubscription(t)
	before := configure.GetConnectedServers()
	if _, err := db.GetDB().Exec(`CREATE TRIGGER reject_subscription_delete BEFORE DELETE ON subscriptions BEGIN SELECT RAISE(ABORT, 'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.GetDB().Exec("DROP TRIGGER reject_subscription_delete") })
	if err := DeleteWhich([]*configure.Which{{NodeRef: configure.NodeRef{TYPE: configure.SubscriptionType, ID: 1}}}); err == nil {
		t.Fatal("expected a deletion failure")
	}
	if configure.GetLenSubscriptions() != 1 || !reflect.DeepEqual(before, configure.GetConnectedServers()) {
		t.Fatal("failed deletion partially changed the catalog or references")
	}
}

func TestDeleteWhichRemovesAutomaticMembers(t *testing.T) {
	for _, kind := range []configure.TouchType{configure.SubscriptionType, configure.ServerType} {
		t.Run(string(kind), func(t *testing.T) {
			resetSubscription(t)
			setting := configure.DefaultOutboundSetting()
			setting.AutoAdd = true
			if err := configure.SetOutboundSetting("proxy", setting); err != nil {
				t.Fatal(err)
			}
			if kind == configure.ServerType {
				if err := configure.AppendServers([]*configure.ServerRaw{{ServerObj: testServer(t, 10002)}}); err != nil {
					t.Fatal(err)
				}
				if err := configure.ClearConnects("proxy"); err != nil {
					t.Fatal(err)
				}
				if err := configure.AddConnect(configure.NodeRef{TYPE: configure.ServerType, ID: 1, Outbound: "proxy"}); err != nil {
					t.Fatal(err)
				}
			}
			if err := DeleteWhich([]*configure.Which{{NodeRef: configure.NodeRef{TYPE: kind, ID: 1}}}); err != nil {
				t.Fatal(err)
			}
			if got := configure.GetConnectedServersByOutbound("proxy"); got.Len() != 0 {
				t.Fatalf("deleted node retained automatic references: %+v", got.Get())
			}
			if !configure.GetOutboundSetting("proxy").AutoAdd {
				t.Fatal("deletion disabled automatic membership")
			}
			if kind == configure.SubscriptionType && configure.GetLenSubscriptions() != 0 {
				t.Fatal("subscription was not deleted")
			}
			if kind == configure.ServerType && configure.GetLenServers() != 0 {
				t.Fatal("server was not deleted")
			}
		})
	}
}

func TestAutomaticProxyRejectsManualEditsWithEmptyOutbound(t *testing.T) {
	resetSubscription(t)
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	before := configure.GetConnectedServers()
	ref := configure.NodeRef{TYPE: configure.SubscriptionServerType, ID: 1, Sub: 0}
	for _, err := range []error{Connect(&ref), Disconnect(ref, false), Disconnect(configure.NodeRef{}, true)} {
		if err == nil || !strings.Contains(err.Error(), "manages its membership automatically") {
			t.Fatalf("manual edit returned %v, want ownership error", err)
		}
	}
	if !reflect.DeepEqual(before, configure.GetConnectedServers()) {
		t.Fatal("rejected edit changed group membership")
	}
}
