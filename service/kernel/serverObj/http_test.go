package serverObj

import "testing"

func TestParseHttpURLDefaultPort(t *testing.T) {
	cases := []struct {
		link     string
		protocol string
		port     int
	}{
		{"http-proxy://user:pass@example.com#name", "http", 80},
		{"https-proxy://example.com", "https", 443},
		{"http-proxy://example.com:8080", "http", 8080},
	}
	for _, c := range cases {
		got, err := ParseHttpURL(c.link)
		if err != nil {
			t.Fatalf("%s: %v", c.link, err)
		}
		if got.Protocol != c.protocol || got.Port != c.port {
			t.Errorf("%s: got %s:%d, want %s:%d", c.link, got.Protocol, got.Port, c.protocol, c.port)
		}
	}
	for _, bad := range []string{
		"http-proxy://example.com:abc",
		"http-proxy://",
		"http-proxy://user:pass@",
		"https://example.com/sub?token=x",
		"http://example.com/sub",
	} {
		if _, err := ParseHttpURL(bad); err == nil {
			t.Errorf("%s should be rejected", bad)
		}
	}
}
