//go:build darwin
// +build darwin

package tun

import "net/netip"

// platformPreExcludeAddrs 在 macOS 上返回额外需要排除的地址。
// macOS 没有 fwmark，与 Windows 类似需要预排除常见 DNS 服务器以防回环，
// 但 macOS 上 sing-tun 的 Inet4RouteExcludeAddress 通常能可靠工作，
// 因此预排除列表保持最小（只防止系统 DNS 回环）。
func platformPreExcludeAddrs() []netip.Prefix {
	return nil
}

// platformTunName 在 macOS 上返回空字符串。
//
// macOS 的 TUN 接口名必须以 "utun" 开头且由内核分配，
// sing-tun 收到空字符串时会自动选择 utun0、utun1 等可用名称。
func platformTunName() string {
	return "" // 由系统自动分配（utun0 / utun1 / …）
}

// platformDisableAutoRoute 在 macOS 上返回 false：
// sing-tun 在 macOS 上的 AutoRoute 工作正常，无需手动管理。
func platformDisableAutoRoute() bool {
	return false
}

// platformPostStart 在 macOS 上通过 networksetup 配置系统 DNS。
func platformPostStart(dnsServers []netip.Addr, tunName string, autoRoute bool) {
	if len(dnsServers) > 0 {
		if err := SetupTunDNS(dnsServers, tunName); err != nil {
			// 非致命，sing-tun 会在 TUN 层面拦截 DNS 查询
		}
	}
}

// platformPreClose 在 macOS 上恢复 DNS 配置。
func platformPreClose(tunName string, autoRoute bool) {
	CleanupTunDNS(tunName)
}
