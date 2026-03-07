package service

import (
	"fmt"

	"github.com/v2rayA/v2rayA/db/configure"
)

var (
	getWorkflowSetting          = GetSetting
	updateWorkflowSetting       = UpdateSetting
	updateWorkflowSubscription  = UpdateSubscription
	getWorkflowSubscription     = configure.GetSubscription
	getWorkflowSubscriptionsLen = configure.GetLenSubscriptions
	getWorkflowConnectedServers = configure.GetConnectedServers
	disconnectWorkflowServer    = Disconnect
	connectWorkflowServer       = Connect
	isSupportedWorkflowServer   = IsSupported
)

func RefreshSubscriptionAndReselect(index int) error {
	if index < 0 || index >= getWorkflowSubscriptionsLen() || getWorkflowSubscription(index) == nil {
		return fmt.Errorf("bad request: ID exceed range")
	}

	setting := getWorkflowSetting()
	nextSetting := *setting
	nextSetting.Transparent = configure.TransparentClose
	if err := updateWorkflowSetting(&nextSetting); err != nil {
		return err
	}
	if err := updateWorkflowSubscription(index, false); err != nil {
		return err
	}
	if err := disconnectSubscriptionServers(index); err != nil {
		return err
	}
	if err := connectFirstAvailableSubscriptionServer(index); err != nil {
		return err
	}
	return nil
}

func disconnectSubscriptionServers(index int) error {
	for _, which := range getWorkflowConnectedServers().Get() {
		if which.TYPE != configure.SubscriptionServerType || which.Sub != index {
			continue
		}
		if err := disconnectWorkflowServer(*which, false); err != nil {
			return err
		}
	}
	return nil
}

func connectFirstAvailableSubscriptionServer(index int) error {
	subscription := getWorkflowSubscription(index)
	if subscription == nil {
		return fmt.Errorf("bad request: ID exceed range")
	}

	var lastErr error
	which := configure.Which{
		TYPE:     configure.SubscriptionServerType,
		Sub:      index,
		Outbound: "proxy",
	}
	for i := range subscription.Servers {
		which.ID = i + 1
		supported, err := isSupportedWorkflowServer(which)
		if err != nil {
			lastErr = err
			continue
		}
		if !supported {
			continue
		}
		if err := connectWorkflowServer(&which); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	if lastErr != nil {
		return fmt.Errorf("no connectable server found in subscription: %w", lastErr)
	}
	return fmt.Errorf("no connectable server found in subscription")
}
