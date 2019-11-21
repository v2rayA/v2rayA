package global

import (
	"V2RayA/model/transparentProxy"
)

var Version = "debug"

var ServiceControlMode SystemServiceControlMode = GetServiceControlMode()

var Iptables *transparentProxy.IpTablesMangle
