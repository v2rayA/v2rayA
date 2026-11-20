package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/server/service"
)

// The unrelated invalid connection makes any attempted reload observable as an
// error, without replacing the production reload path with a mock.
func TestCatalogDeletionReloadsOnlyAffectedGroups(t *testing.T) {
	for _, typ := range []configure.TouchType{configure.ServerType, configure.SubscriptionType} {
		for _, mode := range []string{"unrelated", "connected", "automatic"} {
			t.Run(string(typ)+"/"+mode, func(t *testing.T) {
				runningCoreWithInvalidConnection(t)
				const group = "catalog-delete-test"
				if err := configure.AddOutbound(group); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = configure.RemoveOutbound(group) })
				node := configure.ServerRaw{ServerObj: &serverObj.SOCKS{Server: "192.0.2.1", Port: 1080, Protocol: "socks5"}}
				ref := configure.NodeRef{TYPE: configure.ServerType, ID: configure.GetLenServers() + 1, Outbound: group}
				touch := configure.Which{NodeRef: ref}
				if typ == configure.ServerType {
					if err := configure.AppendServers([]*configure.ServerRaw{&node}); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = configure.RemoveServers([]int{touch.ID - 1}) })
				} else {
					touch.TYPE, touch.ID = typ, configure.GetLenSubscriptions()+1
					ref.TYPE, ref.ID, ref.Sub = configure.SubscriptionServerType, 1, touch.ID-1
					if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{Address: "http://example.invalid", Servers: []configure.ServerRaw{node}}}); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = configure.RemoveSubscriptions([]int{touch.ID - 1}) })
				}
				if mode == "connected" {
					if err := configure.AddConnect(ref); err != nil {
						t.Fatal(err)
					}
				} else if mode == "automatic" {
					setting := configure.GetOutboundSetting(group)
					setting.AutoAdd = true
					if err := configure.SetOutboundSetting(group, setting); err != nil {
						t.Fatal(err)
					}
				}
				process := v2ray.ProcessManager.Process()
				servers, subscriptions := configure.GetServers(), configure.GetSubscriptions()
				members := configure.GetConnectedServers()
				err := service.DeleteWhich([]*configure.Which{&touch})
				if (err != nil) != (mode != "unrelated") {
					t.Fatalf("delete error = %v, mode = %s", err, mode)
				}
				if mode == "unrelated" && (v2ray.ProcessManager.Process() != process || !v2ray.ProcessManager.Running()) {
					t.Fatal("unrelated deletion interrupted the running core")
				}
				if err != nil && (!reflect.DeepEqual(servers, configure.GetServers()) ||
					!reflect.DeepEqual(subscriptions, configure.GetSubscriptions()) ||
					!reflect.DeepEqual(members, configure.GetConnectedServers())) {
					t.Fatal("failed reload did not restore the catalog and group references")
				}
			})
		}
	}
}

func TestDeleteOutboundReloadsOnlyRemovedActiveGroup(t *testing.T) {
	for _, mode := range []string{"empty", "connected", "automatic-empty"} {
		t.Run(mode, func(t *testing.T) {
			runningCoreWithInvalidConnection(t)
			const group = "outbound-delete-test"
			if err := configure.AddOutbound(group); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = configure.RemoveOutbound(group) })
			if mode == "connected" {
				if err := configure.AddConnect(configure.NodeRef{TYPE: configure.ServerType, ID: 999, Outbound: group}); err != nil {
					t.Fatal(err)
				}
			} else if mode == "automatic-empty" {
				setting := configure.GetOutboundSetting(group)
				setting.AutoAdd = true
				if err := configure.SetOutboundSetting(group, setting); err != nil {
					t.Fatal(err)
				}
			}
			setting, members := configure.GetOutboundSetting(group), configure.GetConnectedServersByOutbound(group)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodDelete, "/outbound", strings.NewReader(`{"outbound":"outbound-delete-test"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			DeleteOutbound(ctx)
			var response struct {
				Code common.Code `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if (response.Code == common.FAIL) != (mode != "empty") {
				t.Fatalf("mode %s: %s", mode, recorder.Body.String())
			}
			if mode != "empty" && (!slices.Contains(configure.GetOutbounds(), group) ||
				!reflect.DeepEqual(setting, configure.GetOutboundSetting(group)) ||
				!reflect.DeepEqual(members, configure.GetConnectedServersByOutbound(group))) {
				t.Fatal("failed reload did not restore the outbound and its members")
			}
		})
	}
}

func TestAutomaticSubscriptionReloadsOnlyChangedCatalog(t *testing.T) {
	const first = "socks5://192.0.2.1:1080#first"
	const second = "socks5://192.0.2.2:1080#second"
	for _, tc := range []struct {
		name, body string
		changed    bool
	}{
		{"identical", first + "\n" + second, false},
		{"reordered", second + "\n" + first, true},
		{"added", first + "\n" + second + "\nsocks5://192.0.2.3:1080#third", true},
		{"removed", first, true},
		{"endpoint", "socks5://192.0.2.3:1080#first\n" + second, true},
		{"credentials", "socks5://user:password@192.0.2.1:1080#first\n" + second, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runningCoreWithInvalidConnection(t)
			const group = "subscription-reload-test"
			if err := configure.AddOutbound(group); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = configure.RemoveOutbound(group) })
			setting := configure.GetOutboundSetting(group)
			setting.AutoAdd = true
			if err := configure.SetOutboundSetting(group, setting); err != nil {
				t.Fatal(err)
			}
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Subscription-Userinfo", "download=2000; total=10000")
				fmt.Fprint(w, tc.body)
			}))
			defer source.Close()
			index := configure.GetLenSubscriptions()
			nodes, _, err := service.ResolveByLines(first + "\n" + second)
			if err != nil {
				t.Fatal(err)
			}
			old := &configure.SubscriptionRaw{Address: source.URL, Info: "old usage", Servers: []configure.ServerRaw{{ServerObj: nodes[0], Latency: "10ms"}, {ServerObj: nodes[1]}}}
			if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{old}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = configure.RemoveSubscriptions([]int{index}) })
			process := v2ray.ProcessManager.Process()
			err = service.UpdateSubscriptionWithOptions(index, false, service.SubscriptionFetchOptions{BypassProxy: true})
			if (err != nil) != tc.changed {
				t.Fatalf("update error = %v, catalog changed = %v", err, tc.changed)
			}
			stored := configure.GetSubscription(index)
			if tc.changed {
				if len(stored.Servers) != 2 || stored.Servers[0].ServerObj.ExportToURL() != nodes[0].ExportToURL() || stored.Info != old.Info {
					t.Fatal("failed reload did not restore the old subscription")
				}
			} else {
				if stored.Info == old.Info || stored.Status == "" {
					t.Fatal("metadata was not refreshed")
				}
				if v2ray.ProcessManager.Process() != process || !v2ray.ProcessManager.Running() {
					t.Fatal("unchanged catalog interrupted the core")
				}
			}
		})
	}
}
