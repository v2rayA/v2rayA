//go:build !linux && !darwin && !windows

package httpClient

import (
	"net"
	"time"
)

func DirectDialer(timeout time.Duration, _ bool) *net.Dialer {
	return &net.Dialer{Timeout: timeout}
}
