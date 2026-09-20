package v2ray

import "testing"

func TestCheckDnsUpstream(t *testing.T) {
	ok := []string{"8.8.8.8", "8.8.8.8:53", "localhost", "dns.google", "tcp://8.8.8.8", "tls://dns.google", "tls://1.1.1.1:853", "udp://[2001:4860:4860::8888]:53", "https://dns.google/dns-query", "https://1.1.1.1/dns-query"}
	for _, u := range ok {
		if err := CheckDnsUpstream(u); err != nil {
			t.Errorf("%q rejected: %v", u, err)
		}
	}
	bad := []string{"", "https://", "quic://dns.adguard.com", "ftp://x", "tls://"}
	for _, u := range bad {
		if err := CheckDnsUpstream(u); err == nil {
			t.Errorf("%q accepted", u)
		}
	}
}
