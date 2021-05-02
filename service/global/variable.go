package global

import (
	"os"
	"time"
)

var (
	Version                  = "debug"
	FoundNew                 = false
	RemoteVersion            = ""
	SupportTproxy            = true
	ServiceControlMode       SystemServiceControlMode
	V2RayPID                 *os.Process
	TickerUpdateGFWList      *time.Ticker
	TickerUpdateSubscription *time.Ticker
)

func IsDebug() bool {
	return Version == "debug"
}
