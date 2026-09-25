package httpClient

import (
	"net/http"
	"time"
)

func DirectSubscriptionClient(mark bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = DirectDialer(10*time.Second, mark).DialContext
	transport.DisableKeepAlives = true
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}
}
