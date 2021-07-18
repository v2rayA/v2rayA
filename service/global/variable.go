package global

import (
	"os"
	"time"
)

var (
	Version                  = "debug"
	FoundNew                 = false
	RemoteVersion            = ""
	V2RayPID                 *os.Process
	TickerUpdateGFWList      *time.Ticker
	TickerUpdateSubscription *time.Ticker
)

func IsDebug() bool {
	return Version == "debug"
}
