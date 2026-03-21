//go:build linux
// +build linux

package tun

import (
	"net/netip"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	// TUN_ROUTE_TABLE 是 TUN 接口专用路由表
	TUN_ROUTE_TABLE = 2026
	// FWMARK 是 v2ray/xray 出站及插件流量所使用的标记
	FWMARK = 0x80
)

// SetupTunRouteRules 配置策略路由规则，使标记流量绑过 TUN 接口。
// 这样可防止 v2ray/xray 核心及插件的流量被 TUN 捕获，从而避免路由回环。
func SetupTunRouteRules() error {
	commands := []string{
		// IPv4：将 fwmark 0x80 的流量优先走 main 路由表
		"ip rule add fwmark 0x80 table main pref 100 2>/dev/null || true",
		// IPv6：同上
		"ip -6 rule add fwmark 0x80 table main pref 100 2>/dev/null || true",
	}
	for _, cmd := range commands {
		if err := cmds.ExecCommands(cmd, false); err != nil {
			log.Warn("[TUN] SetupTunRouteRules: 执行命令失败 '%s': %v", cmd, err)
		}
	}
	log.Info("[TUN] Linux 策略路由规则（fwmark 0x80 走 main 表）已设置")
	return nil
}

// CleanupTunRouteRules 删除 SetupTunRouteRules 添加的策略路由规则。
func CleanupTunRouteRules() error {
	commands := []string{
		"ip rule del fwmark 0x80 table main pref 100 2>/dev/null || true",
		"ip -6 rule del fwmark 0x80 table main pref 100 2>/dev/null || true",
	}
	for _, cmd := range commands {
		if err := cmds.ExecCommands(cmd, false); err != nil {
			log.Warn("[TUN] CleanupTunRouteRules: 执行命令失败 '%s': %v", cmd, err)
		}
	}
	log.Info("[TUN] Linux 策略路由规则已清除")
	return nil
}

// SetupExcludeRoutes 在 Linux 上为空操作。
// Linux 通过 fwmark 策略路由实现排除，无需静态路由。
func SetupExcludeRoutes(_ []netip.Prefix) error {
	return nil
}

// CleanupExcludeRoutes 在 Linux 上为空操作。
func CleanupExcludeRoutes() error {
	return nil
}

// SetupTunDNS 在 Linux 上为空操作。
// sing-tun 已通过 SystemdResolved 或 /etc/resolv.conf 处理 DNS。
func SetupTunDNS(_ []netip.Addr, _ string) error {
	return nil
}

// CleanupTunDNS 在 Linux 上为空操作。
func CleanupTunDNS(_ string) error {
	return nil
}
