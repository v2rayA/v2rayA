package conf

import (
	"time"
)

var (
	Version             = "debug"
	FoundNew            = false
	RemoteVersion       = ""
	TickerUpdateGFWList *time.Ticker
)

func IsDebug() bool {
	return Version == "debug"
}
