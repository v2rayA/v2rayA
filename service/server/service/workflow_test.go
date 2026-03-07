package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

func TestRefreshSubscriptionAndReselectSequence(t *testing.T) {
	originalGetSetting := getWorkflowSetting
	originalUpdateSetting := updateWorkflowSetting
	originalUpdateSubscription := updateWorkflowSubscription
	originalGetSubscription := getWorkflowSubscription
	originalGetSubscriptionsLen := getWorkflowSubscriptionsLen
	originalGetConnectedServers := getWorkflowConnectedServers
	originalDisconnectServer := disconnectWorkflowServer
	originalConnectServer := connectWorkflowServer
	originalIsSupported := isSupportedWorkflowServer
	t.Cleanup(func() {
		getWorkflowSetting = originalGetSetting
		updateWorkflowSetting = originalUpdateSetting
		updateWorkflowSubscription = originalUpdateSubscription
		getWorkflowSubscription = originalGetSubscription
		getWorkflowSubscriptionsLen = originalGetSubscriptionsLen
		getWorkflowConnectedServers = originalGetConnectedServers
		disconnectWorkflowServer = originalDisconnectServer
		connectWorkflowServer = originalConnectServer
		isSupportedWorkflowServer = originalIsSupported
	})

	var calls []string
	getWorkflowSubscriptionsLen = func() int { return 1 }
	getWorkflowSubscription = func(index int) *configure.SubscriptionRaw {
		return &configure.SubscriptionRaw{
			Servers: []configure.ServerRaw{{}, {}},
		}
	}
	getWorkflowSetting = func() *configure.Setting {
		return &configure.Setting{Transparent: configure.TransparentProxy}
	}
	updateWorkflowSetting = func(setting *configure.Setting) error {
		if setting.Transparent != configure.TransparentClose {
			t.Fatalf("expected transparent mode close, got %q", setting.Transparent)
		}
		calls = append(calls, "setting")
		return nil
	}
	updateWorkflowSubscription = func(index int, disconnectIfNecessary bool) error {
		if index != 0 {
			t.Fatalf("unexpected subscription index: %d", index)
		}
		if disconnectIfNecessary {
			t.Fatalf("unexpected disconnectIfNecessary=true")
		}
		calls = append(calls, "subscription")
		return nil
	}
	getWorkflowConnectedServers = func() *configure.Whiches {
		return configure.NewWhiches([]*configure.Which{
			{TYPE: configure.SubscriptionServerType, Sub: 0, ID: 2, Outbound: "proxy"},
			{TYPE: configure.ServerType, ID: 1, Outbound: "proxy"},
		})
	}
	disconnectWorkflowServer = func(which configure.Which, clearOutbound bool) error {
		if which.TYPE != configure.SubscriptionServerType || which.Sub != 0 || which.ID != 2 {
			t.Fatalf("unexpected disconnect target: %#v", which)
		}
		if clearOutbound {
			t.Fatalf("unexpected clearOutbound=true")
		}
		calls = append(calls, "disconnect")
		return nil
	}
	isSupportedWorkflowServer = func(which configure.Which) (bool, error) {
		return true, nil
	}
	connectWorkflowServer = func(which *configure.Which) error {
		if which.ID != 1 {
			t.Fatalf("expected first server to be selected, got %d", which.ID)
		}
		calls = append(calls, "connect")
		return nil
	}

	if err := RefreshSubscriptionAndReselect(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"setting", "subscription", "disconnect", "connect"}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("unexpected call sequence: %#v", calls)
	}
}

func TestRefreshSubscriptionAndReselectNoConnectableServer(t *testing.T) {
	originalGetSetting := getWorkflowSetting
	originalUpdateSetting := updateWorkflowSetting
	originalUpdateSubscription := updateWorkflowSubscription
	originalGetSubscription := getWorkflowSubscription
	originalGetSubscriptionsLen := getWorkflowSubscriptionsLen
	originalGetConnectedServers := getWorkflowConnectedServers
	originalDisconnectServer := disconnectWorkflowServer
	originalConnectServer := connectWorkflowServer
	originalIsSupported := isSupportedWorkflowServer
	t.Cleanup(func() {
		getWorkflowSetting = originalGetSetting
		updateWorkflowSetting = originalUpdateSetting
		updateWorkflowSubscription = originalUpdateSubscription
		getWorkflowSubscription = originalGetSubscription
		getWorkflowSubscriptionsLen = originalGetSubscriptionsLen
		getWorkflowConnectedServers = originalGetConnectedServers
		disconnectWorkflowServer = originalDisconnectServer
		connectWorkflowServer = originalConnectServer
		isSupportedWorkflowServer = originalIsSupported
	})

	getWorkflowSubscriptionsLen = func() int { return 1 }
	getWorkflowSubscription = func(index int) *configure.SubscriptionRaw {
		return &configure.SubscriptionRaw{
			Servers: []configure.ServerRaw{{}},
		}
	}
	getWorkflowSetting = func() *configure.Setting { return &configure.Setting{} }
	updateWorkflowSetting = func(setting *configure.Setting) error { return nil }
	updateWorkflowSubscription = func(index int, disconnectIfNecessary bool) error { return nil }
	getWorkflowConnectedServers = func() *configure.Whiches { return nil }
	isSupportedWorkflowServer = func(which configure.Which) (bool, error) { return true, nil }
	connectWorkflowServer = func(which *configure.Which) error { return errors.New("dial failed") }
	disconnectWorkflowServer = func(which configure.Which, clearOutbound bool) error { return nil }

	if err := RefreshSubscriptionAndReselect(0); err == nil {
		t.Fatal("expected error")
	}
}
