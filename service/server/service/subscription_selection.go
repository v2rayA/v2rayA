package service

import (
	"sync"
	"time"
)

var ConfigurationMu sync.Mutex

type subscriptionProbeResult struct {
	latency        time.Duration
	throughput     int64
	speedMeasured  bool
	currentHealthy bool
	err            error
}
