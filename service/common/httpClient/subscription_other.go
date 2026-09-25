//go:build !linux

package httpClient

import "net/http"

func DirectSubscriptionClient() *http.Client { return http.DefaultClient }
