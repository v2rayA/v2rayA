package ntp

import (
	"time"
)

const (
	DisplayFormat = "2006/01/02 15:04 MST"
)

func IsDatetimeSynced() (ok bool, t time.Time, err error) {
	return true, time.Now(), nil
}
