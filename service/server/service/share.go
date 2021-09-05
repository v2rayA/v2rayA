package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/db/configure"
)

func GetSharingAddress(w *configure.Which) (addr string, err error) {
	if w == nil {
		return "", fmt.Errorf("which can not be nil")
	}
	subscriptions := configure.GetSubscriptionsV2()
	if w.TYPE == configure.SubscriptionType {
		ind := w.ID - 1
		if ind < 0 || ind >= len(subscriptions) {
			return "", fmt.Errorf("id exceed range")
		}
		addr = subscriptions[ind].Address
	} else {
		var tsr *configure.ServerRawV2
		tsr, err = w.LocateServerRaw()
		if err != nil {
			return
		}
		addr = tsr.ServerObj.ExportToURL()
		if addr == "" {
			return "", fmt.Errorf("an error occurred while generating the address")
		}
	}
	return
}
