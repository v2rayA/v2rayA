package service

import (
	"sync"
	"time"
)

// ConfigurationMu serializes scheduled updates with API operations that change the core or database.
var ConfigurationMu sync.Mutex

type subscriptionProbeResult struct {
	latency time.Duration
	err     error
}
