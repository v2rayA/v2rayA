package service

import (
	"sync"
	"time"
)

var ConfigurationMu sync.Mutex

type subscriptionProbeResult struct {
	latency time.Duration
	err     error
}
