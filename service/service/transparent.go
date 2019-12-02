package service

import (
	"V2RayA/model/iptables"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
)

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSettingNotNil()
	if !v2ray.IsTransparentSupported(){
		return //TODO: 当前版本v2ray不支持透明代理，需要返回error吗？
	}
	if (!checkRunning || v2ray.IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		_ = iptables.DeleteRules()
		err = iptables.WriteRules()
	}
	return
}

func CheckAndStopTransparentProxy() (err error) {
	if !v2ray.IsTransparentSupported(){
		return //TODO: 当前版本v2ray不支持透明代理，需要返回error吗？
	}
	return iptables.DeleteRules()
}
