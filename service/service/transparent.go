package service

import (
	"V2RayA/model/iptables"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
)

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSettingNotNil()
	if (!checkRunning || v2ray.IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		_ = iptables.DeleteRules()
		err = iptables.WriteRules()
	}
	return
}

func CheckAndStopTransparentProxy() (err error) {
	return iptables.DeleteRules()
}
