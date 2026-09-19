package service

import (
	"fmt"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func GetSharingAddress(w *configure.NodeRef) (addr string, err error) {
	if w == nil {
		return "", fmt.Errorf("no server was given to share")
	}
	subscriptions := configure.GetSubscriptions()
	if w.TYPE == configure.SubscriptionType {
		ind := w.ID - 1
		if ind < 0 || ind >= len(subscriptions) {
			return "", common.Coded("SUBSCRIPTION_NOT_FOUND", fmt.Errorf("subscription #%d does not exist; reload the page", w.ID), map[string]interface{}{"id": w.ID})
		}
		addr = subscriptions[ind].Address
	} else {
		var tsr *configure.ServerRaw
		tsr, err = w.LocateServerRaw()
		if err != nil {
			return
		}
		addr = tsr.ServerObj.ExportToURL()
		if addr == "" {
			return "", fmt.Errorf("server %q has no shareable link; re-import it", tsr.ServerObj.GetName())
		}
	}
	return
}
