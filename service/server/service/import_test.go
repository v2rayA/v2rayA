package service

import "testing"

func TestIsBareHttpProxyLink(t *testing.T) {
	yes := []string{
		"http://user:pass@proxy.example:8080",
		"https://user:pass@proxy.example:8443",
		"http://user@proxy.example:8080#work",
		"http://user:pass@proxy.example:8080/",
	}
	no := []string{
		"http://sub.example.com/link/abc?mu=1",
		"https://sub.example.com/api/v1/client/subscribe?token=x",
		"http://proxy.example:8080",
		"https://example.com",
		"ss://YWVzLTEyOC1nY206dGVzdA@127.0.0.1:8388",
	}
	for _, u := range yes {
		if !isBareHttpProxyLink(u) {
			t.Errorf("%s should be taken for an http proxy link", u)
		}
	}
	for _, u := range no {
		if isBareHttpProxyLink(u) {
			t.Errorf("%s should not be taken for an http proxy link", u)
		}
	}
}
