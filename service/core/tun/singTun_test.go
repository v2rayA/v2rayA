package tun

import (
	"net/netip"
	"testing"
)

func TestGetConnectedProxyServerPrefixes(t *testing.T) {
	// 测试函数是否能正常运行
	prefixes := getConnectedProxyServerPrefixes()

	// 验证返回值类型正确
	if prefixes == nil {
		t.Error("getConnectedProxyServerPrefixes should not return nil")
	}

	// 验证返回的前缀列表格式正确
	for _, prefix := range prefixes {
		if !prefix.IsValid() {
			t.Errorf("Invalid prefix in result: %s", prefix)
		}
		if prefix.Bits() != 32 && prefix.Bits() != 128 {
			t.Errorf("Expected /32 or /128 prefix, got %s", prefix)
		}
	}
}

func TestResolveDnsHost(t *testing.T) {
	// 测试DNS解析功能
	ips := resolveDnsHost("example.com")

	// 验证返回值类型正确
	if ips == nil {
		t.Error("resolveDnsHost should not return nil")
	}

	// 验证IP地址格式正确
	for _, ip := range ips {
		if !ip.IsValid() {
			t.Errorf("Invalid IP in result: %s", ip)
		}
	}
}

func TestResolveDnsServersToExcludes(t *testing.T) {
	// 测试DNS服务器排除功能
	dnsHosts := []string{"8.8.8.8", "1.1.1.1"}
	excludes := ResolveDnsServersToExcludes(dnsHosts)

	// 验证返回值类型正确
	if excludes == nil {
		t.Error("ResolveDnsServersToExcludes should not return nil")
	}

	// 验证返回的前缀列表格式正确
	for _, prefix := range excludes {
		if !prefix.IsValid() {
			t.Errorf("Invalid prefix in result: %s", prefix)
		}
		if prefix.Bits() != 32 && prefix.Bits() != 128 {
			t.Errorf("Expected /32 or /128 prefix, got %s", prefix)
		}
	}
}

func TestIsReservedAddress(t *testing.T) {
	// 测试保留地址检测功能
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"127.0.0.1", true},             // Loopback
		{"192.168.1.1", true},           // Private network
		{"10.0.0.1", true},              // Private network
		{"172.16.0.1", true},            // Private network
		{"169.254.1.1", true},           // Link-local
		{"224.0.0.1", true},             // Multicast
		{"240.0.0.1", true},             // Reserved
		{"0.0.0.1", true},               // Current network
		{"8.8.8.8", false},              // Public IP
		{"1.1.1.1", false},              // Public IP
		{"::1", true},                   // IPv6 Loopback
		{"fe80::1", true},               // IPv6 Link-local
		{"fc00::1", true},               // IPv6 ULA
		{"ff00::1", true},               // IPv6 Multicast
		{"::", true},                    // IPv6 Unspecified
		{"2001:4860:4860::8888", false}, // IPv6 Public
	}

	for _, tc := range testCases {
		ip, err := netip.ParseAddr(tc.ip)
		if err != nil {
			t.Errorf("Failed to parse IP %s: %v", tc.ip, err)
			continue
		}

		result := isReservedAddress(ip)
		if result != tc.expected {
			t.Errorf("isReservedAddress(%s) = %v, expected %v", tc.ip, result, tc.expected)
		}
	}
}
