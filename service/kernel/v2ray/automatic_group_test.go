package v2ray

import (
	"errors"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestEmptyManagedProxyBlocksOnlyItsOwnTraffic(t *testing.T) {
	for _, tc := range []struct {
		name    string
		setting configure.OutboundSetting
	}{
		{name: "automatic membership", setting: func() configure.OutboundSetting {
			setting := configure.DefaultOutboundSetting()
			setting.AutoAdd = true
			return setting
		}()},
		{name: "keep current", setting: func() configure.OutboundSetting {
			setting := configure.DefaultOutboundSetting()
			setting.Type = configure.KeepCurrent
			return setting
		}()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := configure.GetOutboundSetting(configure.DefaultOutboundName)
			t.Cleanup(func() { _ = configure.SetOutboundSetting(configure.DefaultOutboundName, old) })
			if err := configure.SetOutboundSetting(configure.DefaultOutboundName, tc.setting); err != nil {
				t.Fatal(err)
			}
			tmpl := NewEmptyTemplate(configure.GetSettingNotNil())
			data := NewServerData([]serverInfo{{Info: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "other"}, OutboundName: "other"}})
			if _, _, err := tmpl.resolveOutbounds(data); err != nil {
				t.Fatal(err)
			}
			if tmpl.Outbounds[0].Tag != configure.DefaultOutboundName || tmpl.Outbounds[0].Protocol != "blackhole" {
				t.Fatalf("empty managed PROXY falls through: %+v", tmpl.Outbounds)
			}
			for _, tag := range []string{"other", "direct"} {
				found := false
				for _, outbound := range tmpl.Outbounds {
					if outbound.Tag == tag && outbound.Protocol != "blackhole" {
						found = true
					}
				}
				if !found {
					t.Fatalf("unrelated outbound %s was blocked", tag)
				}
			}
		})
	}
}

func TestNativeGroupStrategiesFailClosed(t *testing.T) {
	oldSetting := configure.GetOutboundSetting(configure.DefaultOutboundName)
	oldPorts := *configure.GetPortsNotNil()
	t.Cleanup(func() {
		_ = configure.SetOutboundSetting(configure.DefaultOutboundName, oldSetting)
		_ = configure.SetPorts(&oldPorts)
	})
	ports := oldPorts
	ports.Api.Port = 19790
	if err := configure.SetPorts(&ports); err != nil {
		t.Fatal(err)
	}
	for _, strategy := range []configure.ObservatoryType{
		configure.LeastPing,
		configure.RoundRobin,
		configure.Random,
	} {
		t.Run(strategy.String(), func(t *testing.T) {
			setting := configure.DefaultOutboundSetting()
			setting.Type = strategy
			if err := configure.SetOutboundSetting(configure.DefaultOutboundName, setting); err != nil {
				t.Fatal(err)
			}
			data := NewServerData([]serverInfo{
				socksInfo(configure.DefaultOutboundName, "192.0.2.1"),
				socksInfo(configure.DefaultOutboundName, "192.0.2.2"),
			})
			tmpl := NewEmptyTemplate(configure.GetSettingNotNil())
			if _, _, err := tmpl.resolveOutbounds(data); err != nil {
				t.Fatal(err)
			}
			if _, err := tmpl.SetAPI(data); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = tmpl.Close() })
			if len(tmpl.Routing.Balancers) != 1 {
				t.Fatalf("balancers = %+v", tmpl.Routing.Balancers)
			}
			balancer := tmpl.Routing.Balancers[0]
			if balancer.Strategy.Type != strategy.String() || balancer.FallbackTag != "block" {
				t.Fatalf("strategy %q generated %+v", strategy, balancer)
			}
			if tmpl.MultiObservatory == nil || len(tmpl.MultiObservatory.Observers) != 1 {
				t.Fatalf("strategy %q has no matching observer: %+v", strategy, tmpl.MultiObservatory)
			}
		})
	}
}

func TestPreservedReloadNewProcessFailureTearsInterceptionDownOnce(t *testing.T) {
	setting := *configure.NewSetting()
	setting.Transparent = configure.TransparentProxy
	setting.TransparentType = configure.TransparentTproxy
	wantErr := errors.New("injected process creation failure")
	teardowns := 0
	manager := CoreProcessManager{
		retainedSetting: &setting,
		newProcess: func(*Template, func() error, func() error, func(*Process)) (*Process, error) {
			return nil, wantErr
		},
	}
	manager.retainedTeardown = func(*configure.Setting) {
		teardowns++
		manager.transparentOn.Store(false)
	}
	manager.transparentOn.Store(true)
	env := conf.GetEnvironmentConfig()
	oldEnv := *env
	env.Lite, env.TransparentHook, env.CoreHook = false, "", ""
	t.Cleanup(func() { *env = oldEnv })

	err := manager.start(&Template{Setting: &setting}, true)
	if !errors.Is(err, wantErr) {
		t.Fatalf("start error = %v; want %v", err, wantErr)
	}
	if teardowns != 1 {
		t.Fatalf("transparent teardown count = %d; want 1", teardowns)
	}
	if manager.retainedSetting != nil {
		t.Fatal("retained setting survived failed reload")
	}
	if manager.transparentOn.Load() {
		t.Fatal("transparent interception remained marked active")
	}
}

func TestPausedReloadDoesNotRetainMissingInterception(t *testing.T) {
	setting := *configure.NewSetting()
	setting.Transparent, setting.TransparentType = configure.TransparentProxy, configure.TransparentTproxy
	wantErr := errors.New("stop before starting a core")
	manager := CoreProcessManager{networkPaused: true, retainedSetting: &setting}
	manager.transparentOn.Store(true)
	manager.newProcess = func(*Template, func() error, func() error, func(*Process)) (*Process, error) {
		if manager.retainingInterception.Load() {
			t.Error("paused reload retained interception that the connectivity monitor removed")
		}
		return nil, wantErr
	}
	env := conf.GetEnvironmentConfig()
	previous := *env
	env.Lite, env.TransparentHook, env.CoreHook = false, "", ""
	t.Cleanup(func() { *env = previous })
	if err := manager.start(&Template{Setting: &setting}, true); !errors.Is(err, wantErr) {
		t.Fatalf("start error = %v; want %v", err, wantErr)
	}
}
