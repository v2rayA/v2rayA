package global

import (
	"V2RayA/model/shadowsocksr"
	"os"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode SystemServiceControlMode = GetServiceControlMode()

var SSRs shadowsocksr.SSRs

var V2RayPID *os.Process
