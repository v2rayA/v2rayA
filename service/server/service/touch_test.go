package service

import (
	"testing"

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
