package v2ray

import (
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"testing"
)

func TestEmptyAutomaticProxyBlocksInsteadOfUsingAnotherGroup(t *testing.T) {
	old := configure.GetOutboundSetting("proxy")
	t.Cleanup(func() { configure.SetOutboundSetting("proxy", old) })
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	configure.SetOutboundSetting("proxy", setting)
	tmpl := NewEmptyTemplate(configure.GetSettingNotNil())
	data := NewServerData([]serverInfo{{Info: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "other"}, OutboundName: "other"}})
	if _, _, err := tmpl.resolveOutbounds(data); err != nil {
		t.Fatal(err)
	}
	if tmpl.Outbounds[0].Tag != "proxy" || tmpl.Outbounds[0].Protocol != "blackhole" {
		t.Fatalf("empty PROXY falls through: %+v", tmpl.Outbounds)
	}
	if name, _ := tmpl.FirstProxyOutboundName(nil); name != "proxy" {
		t.Fatalf("default switched to %s", name)
	}
	for _, tag := range []string{"other", "direct"} {
		found := false
		for _, out := range tmpl.Outbounds {
			if out.Tag == tag && out.Protocol != "blackhole" {
				found = true
			}
		}
		if !found {
			t.Fatalf("unrelated outbound %s blocked", tag)
		}
	}
}

func TestRetainInterceptionOnlyForUnchangedPacketProxySettings(t *testing.T) {
	oldLite := conf.GetEnvironmentConfig().Lite
	conf.GetEnvironmentConfig().Lite = false
	t.Cleanup(func() { conf.GetEnvironmentConfig().Lite = oldLite })
	before := *configure.NewSetting()
	before.Transparent = configure.TransparentProxy
	before.TransparentType = configure.TransparentTproxy
	after := before
	if !canRetainInterception(&before, &after) {
		t.Fatal("unchanged TPROXY should keep firewall across group update")
	}
	after.Transparent = configure.TransparentClose
	if canRetainInterception(&before, &after) {
		t.Fatal("manual disable retained firewall")
	}
	after = before
	after.TransparentType = configure.TransparentTun
	before = after
	if canRetainInterception(&before, &after) {
		t.Fatal("core-owned TUN cannot survive core restart")
	}
}
