package service

import (
	"V2RayA/persistence/configure"
	"errors"
)

func GetSharingAddress(w *configure.Which) (addr string, err error) {
	if w == nil {
		return "", errors.New("which不能为nil")
	}
	subscriptions := configure.GetSubscriptions()
	if w.TYPE == configure.SubscriptionType {
		ind := w.ID - 1
		if ind < 0 || ind >= len(subscriptions) {
			return "", errors.New("id超出范围")
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
			return "", errors.New("生成地址时发生错误")
		}
	}
	return
}
