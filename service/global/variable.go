package global

import (
	"os"
)

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode SystemServiceControlMode


var V2RayPID *os.Process

func IsDebug() bool {
	return Version == "debug"
}
