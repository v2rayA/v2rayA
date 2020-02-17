package v2ray

import (
	"V2RayA/model/iptables"
	"V2RayA/persistence/configure"
)

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSettingNotNil()
	if (!checkRunning || IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		iptables.DeleteRules()
		err = iptables.WriteRules()
	}
	return
}

func CheckAndStopTransparentProxy() {
	iptables.DeleteRules()
}
