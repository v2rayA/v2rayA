package ntp

import (
	"V2RayA/common"
	"github.com/beevik/ntp"
	"time"
)

func IsDatetimeSynced() (bool, error) {
	t, err := ntp.Time("ntp1.aliyun.com")
	if err != nil {
		return false, newError(err)
	}
	if common.Abs(t.UTC().Second()-time.Now().UTC().Second()) >= 120 {
		return false, nil
	}
	return true, nil
}
