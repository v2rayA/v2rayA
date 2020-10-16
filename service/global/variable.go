package global

import (
	"os"
	"github.com/v2rayA/v2rayA/plugin"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode SystemServiceControlMode

var Plugins plugin.Plugins

var V2RayPID *os.Process

func IsDebug() bool {
	return Version == "debug"
}
