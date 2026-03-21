//go:build windows
// +build windows

package tun

import "net/netip"

// platformPreExcludeAddrs 在 Windows 上预排除常用公网 DNS 服务器地址。
//
// Windows 没有 fwmark 机制，如果 v2ray/xray 核心向这些 DNS 发送直连请求，
// 而 TUN 恰好将这些目标地址的流量也劫持，就会导致路由回环。
// 预排除后，这些地址的流量会走物理网卡直接出去。
func platformPreExcludeAddrs() []netip.Prefix {
	var prefixes []netip.Prefix
	wellKnownDNS := []string{
		// Cloudflare
		"1.1.1.1/32", "1.0.0.1/32",
		// Google
		"8.8.8.8/32", "8.8.4.4/32",
		// Quad9
		"9.9.9.9/32", "149.112.112.112/32",
		// OpenDNS
		"208.67.222.222/32", "208.67.220.220/32",
		// 国内常用
		"114.114.114.114/32",
		"223.5.5.5/32", "223.6.6.6/32",
		// IPv6 Google
		"2001:4860:4860::8888/128", "2001:4860:4860::8844/128",
		// IPv6 Cloudflare
		"2606:4700:4700::1111/128", "2606:4700:4700::1001/128",
	}
	for _, cidr := range wellKnownDNS {
		if p, err := netip.ParsePrefix(cidr); err == nil {
			prefixes = append(prefixes, p)
		}
	}
	return prefixes
}

// platformTunName 在 Windows 上返回自定义 TUN 接口名称。
func platformTunName() string {
	return "v2raya-tun"
}

// platformDisableAutoRoute 在 Windows 上返回 true。
//
// sing-tun 的 AutoRoute 在 Windows 上可靠性不足，
// 改由 SetupTunRouteRules 手动添加默认路由（metric=1）。
func platformDisableAutoRoute() bool {
	return true
}

// platformPostStart 在 TUN 启动后为接口配置 DNS 服务器。
// sing-tun 在 Windows 上不会自动将 DNSServers 写入系统接口配置。
func platformPostStart(dnsServers []netip.Addr, tunName string) {
	if len(dnsServers) > 0 {
		if err := SetupTunDNS(dnsServers, tunName); err != nil {
			// 非致命错误：DNS 设置失败不影响流量转发
		}
	}
}

// platformPreClose 在 TUN 关闭前清理 Windows 特有资源。
func platformPreClose(tunName string) {
	if tunName != "" {
		CleanupTunDNS(tunName)
	}
}
