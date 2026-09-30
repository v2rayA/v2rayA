package v2ray

import (
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestGroupStrategiesUseNativeBalancer(t *testing.T) {
	name := configure.DefaultOutboundName
	previous := configure.GetOutboundSetting(name)
	t.Cleanup(func() { _ = configure.SetOutboundSetting(name, previous) })
	for _, strategy := range []configure.ObservatoryType{
		configure.LeastPing, configure.LeastLoad, configure.RoundRobin, configure.Random,
	} {
		t.Run(strategy.String(), func(t *testing.T) {
			setting := configure.DefaultOutboundSetting()
			setting.Type = strategy
			if err := configure.SetOutboundSetting(name, setting); err != nil {
				t.Fatal(err)
			}
			tmpl := NewEmptyTemplate(configure.GetSettingNotNil())
			t.Cleanup(func() { _ = tmpl.Close() })
			data := NewServerData([]serverInfo{
				{Info: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "first"}, OutboundName: name},
				{Info: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1081, Protocol: "socks5", Name: "second"}, OutboundName: name},
			})
			if _, _, err := tmpl.resolveOutbounds(data); err != nil {
				t.Fatal(err)
			}
			if _, err := tmpl.SetAPI(data); err != nil {
				t.Fatal(err)
			}
			if len(tmpl.Routing.Balancers) != 1 {
				t.Fatalf("balancers=%d, want 1", len(tmpl.Routing.Balancers))
			}
			balancer := tmpl.Routing.Balancers[0]
			if balancer.Tag != name || balancer.Strategy.Type != strategy.String() || balancer.FallbackTag != "block" {
				t.Fatalf("unexpected balancer: %+v", balancer)
			}
			if tmpl.MultiObservatory == nil || len(tmpl.MultiObservatory.Observers) != 1 {
				t.Fatal("strategy with block fallback has no Xray observatory")
			}
		})
	}
}
