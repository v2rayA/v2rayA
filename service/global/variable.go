package global

import (
	"V2RayA/model/transparentProxy"
)

var ServiceControlMode SystemServiceControlMode = GetServiceControlMode()

var Iptables *transparentProxy.IpTablesMangle
