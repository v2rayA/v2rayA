package service

import (
	"os"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "v2raya-service-test-*")
	if err != nil {
		panic(err)
	}
	conf.GetEnvironmentConfig().Config = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestHttpLatencyDropsSubscriptions(t *testing.T) {
	input := []*configure.Which{{TYPE: configure.SubscriptionType, ID: 1}}
	got, err := TestHttpLatency(input, time.Second, 1, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("subscription remained in latency results: %+v", got)
	}
	if input[0].Latency != "" {
		t.Fatalf("subscription was probed: %q", input[0].Latency)
	}
}
