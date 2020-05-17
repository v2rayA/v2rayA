package global

import (
	"os"
	"v2rayA/plugins"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode SystemServiceControlMode

var Plugins plugins.Plugins

var V2RayPID *os.Process

func IsDebug() bool {
	return Version == "debug"
}
