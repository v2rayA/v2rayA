package service

import (
	"V2RayA/global"
	"V2RayA/model/transparentProxy"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
)

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSetting()
	if (!checkRunning || v2ray.IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		if global.Iptables != nil {
			_ = transparentProxy.StopTransparentProxy(global.Iptables)
		}
		global.Iptables, err = transparentProxy.StartTransparentProxy()
	}
	return
}

func CheckAndStopTransparentProxy() (err error) {
	if global.Iptables == nil {
		return
	}
	err = transparentProxy.StopTransparentProxy(global.Iptables)
	global.Iptables = nil
	return
}
