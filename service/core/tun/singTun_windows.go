//go:build windows
// +build windows

package tun

import (
	"net"
	"net/netip"
	"strings"

	"github.com/v2rayA/v2rayA/db/configure"
)

// parseDNSServerHost 解析 DNS 服务器地址字符串，返回主机名列表
func parseDNSServerHost(server string) []string {
	var hosts []string

	// 处理 localhost
	if server == "localhost" {
		hosts = append(hosts, "127.0.0.1", "::1")
		return hosts
	}

	// 处理带协议的地址
	if strings.Contains(server, "://") {
		// 移除协议前缀
		server = strings.TrimPrefix(server, "https://")
		server = strings.TrimPrefix(server, "tls://")
		server = strings.TrimPrefix(server, "tcp://")
		server = strings.TrimPrefix(server, "udp://")

		// 如果包含路径，只取主机部分
		if strings.Contains(server, "/") {
			parts := strings.Split(server, "/")
			server = parts[0]
		}
	}

	// 处理带端口的地址
	if strings.Contains(server, ":") {
		// 分离主机和端口
		host, _, err := net.SplitHostPort(server)
		if err == nil && host != "" {
			hosts = append(hosts, host)
		}
	} else {
		// 纯 IP 或域名
		hosts = append(hosts, server)
	}

	return hosts
}

// platformPreExcludeAddrs 在 Windows 上根据用户 DNS 配置返回需要排除的地址列表
//
// Windows 没有 fwmark 机制，如果 v2ray/xray 核心向这些 DNS 发送直连请求，
// 而 TUN 恰好将这些目标地址的流量也劫持，就会导致路由回环。
// 只排除用户配置中设置为 "direct" 的 DNS 服务器地址。
func platformPreExcludeAddrs() []netip.Prefix {
	var prefixes []netip.Prefix

	// 获取用户配置的 DNS 规则
	dnsRules := configure.GetDnsRulesNotNil()

	// 只排除设置为 "direct" 的 DNS 服务器
	for _, rule := range dnsRules {
		if rule.Outbound == "direct" {
			dnsHosts := parseDNSServerHost(rule.Server)
			for _, host := range dnsHosts {
				// 解析主机名为 IP 地址
				ips := resolveDnsHost(host)
				for _, ip := range ips {
					prefix := netip.PrefixFrom(ip, ip.BitLen())
					prefixes = append(prefixes, prefix)
				}
			}
		}
	}

	return prefixes
}

// platformTunName 在 Windows 上返回自定义 TUN 接口名称。
func platformTunName() string {
	return "v2raya-tun"
}

// platformDisableAutoRoute 在 Windows 上返回 false。
//
// sing-tun v0.8.0+ 在 Windows 上通过 winipcfg LUID API 和 FWPM (StrictRoute) 实现
// 可靠的路由管理，无需手动管理。
func platformDisableAutoRoute() bool {
	return false
}

// platformPostStart 在 TUN 启动后为接口配置 DNS 服务器。
//
// 当 AutoRoute 启用时，sing-tun 已通过 luid.SetDNS() 处理 DNS，无需手动配置。
// 当 AutoRoute 关闭时（用户显式禁用），手动调用 SetupTunDNS。
func platformPostStart(dnsServers []netip.Addr, tunName string, autoRoute bool) {
	if autoRoute {
		// sing-tun 通过 winipcfg luid.SetDNS() 自动设置 DNS
		return
	}
	if len(dnsServers) > 0 {
		if err := SetupTunDNS(dnsServers, tunName); err != nil {
			// 非致命错误：DNS 设置失败不影响流量转发
		}
	}
}

// platformPreClose 在 TUN 关闭前清理 Windows 特有资源。
// AutoRoute 启用时 sing-tun 自行清理，无需手动操作。
func platformPreClose(tunName string, autoRoute bool) {
	if autoRoute {
		return
	}
	if tunName != "" {
		CleanupTunDNS(tunName)
	}
}
