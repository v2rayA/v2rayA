package ntp

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

const (
	DisplayFormat = "2006/01/02 15:04 MST"
)

var (
	ntpSyncCache struct {
		value    bool
		lastReq  time.Time
		realTime time.Time
		mu       sync.Mutex
	}
)

func IsDatetimeSynced() (ok bool, t time.Time, err error) {
	ntpSyncCache.mu.Lock()
	defer ntpSyncCache.mu.Unlock()
	if time.Since(ntpSyncCache.lastReq) < 5*time.Second {
		return ntpSyncCache.value, ntpSyncCache.realTime, nil
	}
	defer func() {
		// Do not care about the success.
		ntpSyncCache.value = ok
		ntpSyncCache.lastReq = time.Now()
		ntpSyncCache.realTime = t
	}()
	t, err = ntp.Time("ntp.aliyun.com")
	if err != nil {
		return false, time.Time{}, fmt.Errorf("IsDatetimeSynced: %w", err)
	}
	if seconds := t.Sub(time.Now().UTC()).Seconds(); seconds >= 90 || seconds <= -90 {
		return false, t, nil
	}
	return true, t, nil
}
