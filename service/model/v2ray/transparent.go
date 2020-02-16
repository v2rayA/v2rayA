package v2ray

import (
	"V2RayA/model/iptables"
	"V2RayA/persistence/configure"
	"log"
)

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSettingNotNil()
	if (!checkRunning || IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		iptables.DeleteRules()
		err = iptables.WriteRules()
		log.Println("CheckAndSetupTransparentProxy: set iptables rules")
	}
	return
}

func CheckAndStopTransparentProxy() {
	iptables.DeleteRules()
}
