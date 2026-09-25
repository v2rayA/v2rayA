//go:build !linux

package httpClient

import (
	"net"
	"net/http"
	"time"
)

func DirectDialer(timeout time.Duration) *net.Dialer { return &net.Dialer{Timeout: timeout} }

func DirectSubscriptionClient() *http.Client { return http.DefaultClient }
