package service

import (
	"V2RayA/model/transparentProxy"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
)

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSetting()
	if (!checkRunning || v2ray.IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		_ = transparentProxy.DeleteRules()
		err = transparentProxy.WriteRules()
	}
	return
}

func CheckAndStopTransparentProxy() (err error) {
	return transparentProxy.DeleteRules()
}
