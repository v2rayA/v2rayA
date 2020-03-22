package global

import (
	"V2RayA/core/shadowsocksr"
	"os"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = false

var ServiceControlMode SystemServiceControlMode = GetServiceControlMode()

var SSRs shadowsocksr.SSRs

var V2RayPID *os.Process
