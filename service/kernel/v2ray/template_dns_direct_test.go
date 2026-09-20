package v2ray

import "testing"

func TestDirectDnsServersFollowTheRules(t *testing.T) {
	// the default rules: localhost direct (skipped: loopback), 223.5.5.5 direct, 1.0.0.1 through the proxy
	got := directDnsServers()
	if len(got) != 1 || got[0] != "223.5.5.5:53" {
		t.Fatalf("directDnsServers() = %v", got)
	}
	if fallbackResolver() != "223.5.5.5" {
		t.Fatalf("fallbackResolver() = %q", fallbackResolver())
	}
}
