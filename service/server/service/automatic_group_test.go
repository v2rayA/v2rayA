package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

func TestAutomaticGroupTracksImportsUpdatesAndDeletionWhileStopped(t *testing.T) {
	previous := configure.GetOutboundSetting("proxy")
	setting := previous
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.SetOutboundSetting("proxy", previous); _ = configure.ClearConnects("proxy") })
	const link = "socks5://192.0.2.1:1#unreachable"
	serverIndex, subIndex := configure.GetLenServers(), configure.GetLenSubscriptions()
	t.Cleanup(func() {
		_ = configure.RemoveServers([]int{serverIndex})
		_ = configure.RemoveSubscriptions([]int{subIndex})
	})
	if err := ImportServer(link, nil); err != nil {
		t.Fatal(err)
	}
	body := link
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
	defer source.Close()
	if err := ImportSubscription(source.URL); err != nil {
		t.Fatal(err)
	}
	before := configure.GetConnectedServersByOutbound("proxy").Len()
	for _, outbound := range []string{"", "proxy"} {
		ref := configure.NodeRef{TYPE: configure.ServerType, ID: serverIndex + 1, Outbound: outbound}
		if Connect(&ref) == nil || Disconnect(ref, false) == nil || ReplaceOutboundConnections(outbound, nil) == nil {
			t.Fatal("automatic ownership can be bypassed")
		}
	}
	if err := SelectServersFromSubscription(subIndex, true); err != nil {
		t.Fatal(err)
	}
	if configure.GetConnectedServersByOutbound("proxy").Len() != before {
		t.Fatal("legacy autoSelect changed automatic membership")
	}
	body = "socks5://192.0.2.2:2#replacement\nsocks5://192.0.2.3:3#another"
	if err := UpdateSubscription(subIndex, false); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetSubscription(subIndex); len(got.Servers) != 2 {
		t.Fatalf("old automatic member was retained: %+v", got.Servers)
	}
	if got := configure.GetConnectedServersByOutbound("proxy").Len(); got != before+1 {
		t.Fatalf("updated catalog not reflected: %d", got)
	}
	if err := DeleteWhich([]*configure.Which{
		{NodeRef: configure.NodeRef{TYPE: configure.ServerType, ID: serverIndex + 1}},
		{NodeRef: configure.NodeRef{TYPE: configure.SubscriptionType, ID: subIndex + 1}},
	}); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetConnectedServersByOutbound("proxy").Len(); got != before-2 {
		t.Fatalf("deleted nodes remain: %d", got)
	}
}
