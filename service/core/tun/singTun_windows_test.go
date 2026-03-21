//go:build windows
// +build windows

package tun

import (
	"testing"
)

func TestParseDNSServerHost(t *testing.T) {
	tests := []struct {
		server   string
		expected []string
	}{
		{"localhost", []string{"127.0.0.1", "::1"}},
		{"8.8.8.8", []string{"8.8.8.8"}},
		{"223.5.5.5", []string{"223.5.5.5"}},
		{"https://dns.google/dns-query", []string{"dns.google"}},
		{"tls://1.1.1.1:853", []string{"1.1.1.1"}},
		{"tcp://8.8.4.4:53", []string{"8.8.4.4"}},
		{"udp://9.9.9.9:53", []string{"9.9.9.9"}},
		{"cloudflare-dns.com", []string{"cloudflare-dns.com"}},
	}

	for _, test := range tests {
		result := parseDNSServerHost(test.server)
		if len(result) != len(test.expected) {
			t.Errorf("parseDNSServerHost(%s) = %v, expected %v", test.server, result, test.expected)
			continue
		}
		for i, host := range result {
			if host != test.expected[i] {
				t.Errorf("parseDNSServerHost(%s)[%d] = %s, expected %s", test.server, i, host, test.expected[i])
			}
		}
	}
}

func TestPlatformPreExcludeAddrs(t *testing.T) {
	// 测试函数能正常返回前缀列表
	prefixes := platformPreExcludeAddrs()

	// 验证返回的前缀都是有效的
	for _, prefix := range prefixes {
		if !prefix.IsValid() {
			t.Errorf("返回了无效的前缀: %s", prefix)
		}
	}
}
