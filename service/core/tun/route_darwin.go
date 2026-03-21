//go:build darwin
// +build darwin

package tun

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var excludedRoutes []netip.Prefix

// SetupTunRouteRules 在 macOS 上为空操作。
//
// macOS 不支持 fwmark 策略路由，sing-tun 的 AutoRoute
// 会自动通过路由表优先级将流量导向 TUN 接口。
func SetupTunRouteRules() error {
	return nil
}

// CleanupTunRouteRules 在 macOS 上为空操作。
func CleanupTunRouteRules() error {
	return nil
}

// SetupExcludeRoutes 在 macOS 上为代理服务端地址添加"绕过 TUN"的静态主机路由。
//
// macOS 没有 fwmark，需要为每个服务端 IP 显式添加经物理网关的路由，
// 以避免代理流量被 TUN 再次捕获。
func SetupExcludeRoutes(addrs []netip.Prefix) error {
	if len(addrs) == 0 {
		return nil
	}
	excludedRoutes = addrs

	gw, err := getDefaultGateway()
	if err != nil {
		log.Warn("[TUN][macOS] 获取默认网关失败: %v", err)
		return err
	}

	for _, prefix := range addrs {
		addr := prefix.Addr()
		var out []byte
		if addr.Is4() {
			// macOS route add -host <ip> <gw>  （主机路由，精确匹配单 IP）
			out, err = exec.Command("route", "add", "-host", addr.String(), gw).CombinedOutput()
		} else {
			// macOS route add -inet6 <prefix> <gw>
			out, err = exec.Command("route", "add", "-inet6", prefix.String(), gw).CombinedOutput()
		}
		if err != nil {
			s := string(out)
			// 忽略"路由已存在"的错误
			if !strings.Contains(s, "File exists") && !strings.Contains(s, "already exists") {
				log.Warn("[TUN][macOS] 添加排除路由 %s 失败: %v, output: %s", addr, err, s)
			} else {
				log.Info("[TUN][macOS] 排除路由 %s 已存在", addr)
			}
		} else {
			log.Info("[TUN][macOS] 添加排除路由 %s → %s", addr, gw)
		}
	}
	return nil
}

// CleanupExcludeRoutes 删除 SetupExcludeRoutes 添加的所有静态路由。
func CleanupExcludeRoutes() error {
	for _, prefix := range excludedRoutes {
		addr := prefix.Addr()
		var out []byte
		var err error
		if addr.Is4() {
			out, err = exec.Command("route", "delete", "-host", addr.String()).CombinedOutput()
		} else {
			out, err = exec.Command("route", "delete", "-inet6", prefix.String()).CombinedOutput()
		}
		if err != nil {
			log.Warn("[TUN][macOS] 删除排除路由 %s 失败: %v, output: %s", addr, err, string(out))
		}
	}
	excludedRoutes = nil
	return nil
}

// SetupTunDNS 在 macOS 上通过 networksetup 为主网络服务设置 DNS 服务器。
//
// 如果找不到主网络服务，则跳过（sing-tun 会通过 DNSServers 选项
// 在 TUN 接口级别完成 DNS 截获，不一定需要系统级配置）。
func SetupTunDNS(dnsServers []netip.Addr, _ string) error {
	if len(dnsServers) == 0 {
		return nil
	}

	svc, err := getPrimaryNetworkService()
	if err != nil {
		log.Warn("[TUN][macOS] SetupTunDNS: 获取主网络服务失败 (%v)，跳过 DNS 设置", err)
		return nil
	}

	var addrs []string
	for _, dns := range dnsServers {
		if dns.Is4() { // 优先使用 IPv4 DNS
			addrs = append(addrs, dns.String())
		}
	}
	if len(addrs) == 0 {
		return nil
	}

	args := append([]string{"-setdnsservers", svc}, addrs...)
	out, err := exec.Command("networksetup", args...).CombinedOutput()
	if err != nil {
		log.Warn("[TUN][macOS] 设置 DNS 失败: %v, output: %s", err, string(out))
		return err
	}
	log.Info("[TUN][macOS] 网络服务 '%s' DNS 已设置: %s", svc, strings.Join(addrs, ", "))
	return nil
}

// CleanupTunDNS 将主网络服务的 DNS 恢复为自动获取（DHCP/Empty）。
func CleanupTunDNS(_ string) error {
	svc, err := getPrimaryNetworkService()
	if err != nil {
		return nil
	}
	out, err := exec.Command("networksetup", "-setdnsservers", svc, "Empty").CombinedOutput()
	if err != nil {
		log.Warn("[TUN][macOS] 重置 DNS 失败: %v, output: %s", err, string(out))
	} else {
		log.Info("[TUN][macOS] 网络服务 '%s' DNS 已重置为自动", svc)
	}
	return nil
}

// getDefaultGateway 在 macOS 上通过 route -n get default 获取默认 IPv4 网关。
func getDefaultGateway() (string, error) {
	out, err := exec.Command("sh", "-c", "route -n get default 2>/dev/null | awk '/gateway/{print $2}'").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("获取默认网关失败: %w, output: %s", err, string(out))
	}
	gw := strings.TrimSpace(string(out))
	if gw == "" {
		return "", fmt.Errorf("默认网关为空（可能无网络连接）")
	}
	return gw, nil
}

// getPrimaryNetworkService 通过当前默认路由的网络接口推断主网络服务名称。
func getPrimaryNetworkService() (string, error) {
	// 1. 获取当前默认路由使用的接口名（如 en0）
	ifOut, err := exec.Command("sh", "-c", "route -n get default 2>/dev/null | awk '/interface/{print $2}'").CombinedOutput()
	if err != nil || strings.TrimSpace(string(ifOut)) == "" {
		return "", fmt.Errorf("未能获取默认接口名")
	}
	ifName := strings.TrimSpace(string(ifOut))

	// 2. 遍历 networksetup 的服务列表，找到对应该接口的服务
	svcOut, err := exec.Command("networksetup", "-listallhardwareports").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("networksetup -listallhardwareports 失败: %w", err)
	}

	// 输出格式：
	// Hardware Port: Wi-Fi
	// Device: en0
	// Ethernet Address: ...
	lines := strings.Split(string(svcOut), "\n")
	var lastSvc string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Hardware Port:") {
			lastSvc = strings.TrimSpace(strings.TrimPrefix(line, "Hardware Port:"))
		}
		if strings.HasPrefix(line, "Device:") {
			dev := strings.TrimSpace(strings.TrimPrefix(line, "Device:"))
			if dev == ifName {
				return lastSvc, nil
			}
		}
	}
	return "", fmt.Errorf("未找到接口 %s 对应的网络服务", ifName)
}
