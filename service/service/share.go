package service

import (
	"v2rayA/db/configure"
)

func GetSharingAddress(w *configure.Which) (addr string, err error) {
	if w == nil {
		return "", newError("which can not be nil")
	}
	subscriptions := configure.GetSubscriptions()
	if w.TYPE == configure.SubscriptionType {
		ind := w.ID - 1
		if ind < 0 || ind >= len(subscriptions) {
			return "", newError("id exceed range")
		}
		addr = subscriptions[ind].Address
	} else {
		var tsr *configure.ServerRaw
		tsr, err = w.LocateServer()
		if err != nil {
			return
		}
		addr = tsr.VmessInfo.ExportToURL()
		if addr == "" {
			return "", newError("an error occurred while generating the address")
		}
	}
	return
}
