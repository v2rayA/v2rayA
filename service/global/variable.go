package global

import (
	"v2rayA/plugins"
	"os"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode = GetServiceControlMode()

var Plugins plugins.Plugins

var V2RayPID *os.Process
