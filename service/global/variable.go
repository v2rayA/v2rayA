package global

import (
	"V2RayA/core/shadowsocksr"
	"os"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode = GetServiceControlMode()

var SSRs shadowsocksr.SSRs

var V2RayPID *os.Process
