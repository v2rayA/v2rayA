package global

import (
	"V2RayA/model/transparentProxy"
)

var Version = "0.0"

var ServiceControlMode SystemServiceControlMode = GetServiceControlMode()

var Iptables *transparentProxy.IpTablesMangle
