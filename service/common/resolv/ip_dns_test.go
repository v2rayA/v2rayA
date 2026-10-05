package resolv

import "testing"

func TestParseIPDNS(t *testing.T) {
	for _, tc := range []struct{ input, want, protocol string }{
		{"1.1.1.1", "udp://1.1.1.1:53", "udp"},
		{"1.1.1.1:5353", "udp://1.1.1.1:5353", "udp"},
		{"2001:db8::1", "udp://[2001:db8::1]:53", "udp"},
		{"[2001:0db8::1]:5353", "udp://[2001:db8::1]:5353", "udp"},
		{"tcp://[2001:db8::1]", "tcp://[2001:db8::1]:53", "tcp"},
		{"tls://1.1.1.1", "tls://1.1.1.1:853", "tcp-tls"},
		{"https://[2001:db8::1]/dns-query?a=1", "https://[2001:db8::1]:443/dns-query?a=1", "https"},
		{"udp://127.0.0.53", "udp://127.0.0.53:53", "udp"},
		{"udp://[::ffff:1.1.1.1]:0053", "udp://1.1.1.1:53", "udp"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			e, err := ParseIPDNS(tc.input)
			if err != nil || e.URL != tc.want || e.Protocol() != tc.protocol {
				t.Fatalf("got %+v, %v; want %s %s", e, err, tc.want, tc.protocol)
			}
		})
	}
	for _, input := range []string{"", "dns.google", "tls://dns.google", "https://dns.google/dns-query", "udp://0.0.0.0", "udp://[::]", "udp://224.0.0.1", "udp://255.255.255.255", "udp://[ff02::1]", "udp://1.1.1.1:0", "udp://1.1.1.1:65536", "udp://1.1.1.1:bad", "udp://1.1.1.1:", "quic://1.1.1.1", "udp://1.1.1.1/path", "tcp://1.1.1.1?a=1", "https://user@1.1.1.1/path", "https://1.1.1.1/path#fragment"} {
		if e, err := ParseIPDNS(input); err == nil {
			t.Errorf("accepted %q as %+v", input, e)
		}
	}
}

func TestFallbackDNSURLsIsReadOnly(t *testing.T) {
	urls := FallbackDNSURLs()
	urls[0] = "changed"
	if FallbackDNSURLs()[0] == "changed" {
		t.Fatal("caller mutated built-in DNS list")
	}
	if FallbackDNSURLs()[1] != "tcp://119.29.29.29:53" {
		t.Fatal("fallback protocol lost")
	}
}
