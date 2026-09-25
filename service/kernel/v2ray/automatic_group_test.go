package v2ray

import (
	"errors"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestEmptyAutomaticProxyBlocksOnlyItsOwnTraffic(t *testing.T) {
	old := configure.GetOutboundSetting(configure.DefaultOutboundName)
	t.Cleanup(func() { _ = configure.SetOutboundSetting(configure.DefaultOutboundName, old) })
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting(configure.DefaultOutboundName, setting); err != nil {
		t.Fatal(err)
	}
	tmpl := NewEmptyTemplate(configure.GetSettingNotNil())
	data := NewServerData([]serverInfo{{Info: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "other"}, OutboundName: "other"}})
	if _, _, err := tmpl.resolveOutbounds(data); err != nil {
		t.Fatal(err)
	}
	if tmpl.Outbounds[0].Tag != configure.DefaultOutboundName || tmpl.Outbounds[0].Protocol != "blackhole" {
		t.Fatalf("empty automatic PROXY falls through: %+v", tmpl.Outbounds)
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
