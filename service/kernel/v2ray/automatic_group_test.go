package v2ray

import (
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestEmptyAutomaticGroupDoesNotFallThrough(t *testing.T) {
	previous := configure.GetOutboundSetting("proxy")
	t.Cleanup(func() { _ = configure.SetOutboundSetting("proxy", previous) })
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	tmpl := NewEmptyTemplate(configure.GetSettingNotNil())
	defer tmpl.Close()
	data := NewServerData([]serverInfo{{Info: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "other"}, OutboundName: "other"}})
	if _, _, err := tmpl.resolveOutbounds(data); err != nil {
		t.Fatal(err)
	}
	if tmpl.Outbounds[0].Tag != "proxy" || tmpl.Outbounds[0].Protocol != "blackhole" {
		t.Fatalf("empty automatic group falls through: %+v", tmpl.Outbounds)
	}
	for _, tag := range []string{"other", "direct"} {
		found := false
		for _, out := range tmpl.Outbounds {
			if out.Tag == tag && out.Protocol != "blackhole" {
				found = true
			}
		}
		if !found {
			t.Fatalf("unrelated outbound %s was blocked", tag)
		}
	}
}
