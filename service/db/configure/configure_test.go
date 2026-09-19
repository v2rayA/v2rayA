package configure

import (
	"os"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "v2raya-configure-test-*")
	if err != nil {
		panic(err)
	}
	conf.GetEnvironmentConfig().Config = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestGetLenSubscriptionServersPanicsOnMissingSubscription(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("missing subscription must report a recoverable panic")
		}
	}()
	GetLenSubscriptionServers(-1)
}
