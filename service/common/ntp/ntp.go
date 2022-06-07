package ntp

import (
	"fmt"
	"github.com/beevik/ntp"
	"time"
)

const (
	DisplayFormat = "2006/01/02 15:04 MST"
)

func IsDatetimeSynced() (bool, time.Time, error) {
	t, err := ntp.Time("ntp.aliyun.com")
	if err != nil {
		return false, time.Time{}, fmt.Errorf("IsDatetimeSynced: %w", err)
	}
	if seconds := t.Sub(time.Now().UTC()).Seconds(); seconds >= 90 || seconds <= -90 {
		return false, t, nil
	}
	return true, t, nil
}
