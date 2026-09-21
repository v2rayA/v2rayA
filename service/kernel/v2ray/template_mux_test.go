package v2ray

import (
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestMuxAppliesToVmessOnly(t *testing.T) {
	tmpl := &Template{Setting: configure.NewSetting()}
	tmpl.Setting.MuxOn = configure.Yes
	tmpl.Setting.Mux = 8
	tmpl.Outbounds = []coreObj.OutboundObject{
		{Tag: "a", Protocol: "vmess"},
		{Tag: "b", Protocol: "vless"},
		{Tag: "direct", Protocol: "freedom"},
	}
	tmpl.SetOutboundSockopt()
	if m := tmpl.Outbounds[0].Mux; m == nil || !m.Enabled || m.Concurrency != 8 {
		t.Fatalf("vmess mux = %+v", m)
	}
	if tmpl.Outbounds[1].Mux != nil || tmpl.Outbounds[2].Mux != nil {
		t.Fatal("mux set on a non-vmess outbound")
	}
}
