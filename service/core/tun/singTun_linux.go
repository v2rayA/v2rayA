//go:build linux
// +build linux

package tun

import "net/netip"

// platformPreExcludeAddrs 在 Linux 上返回额外需要排除的地址列表。
// Linux 通过 fwmark 策略路由保证 v2ray/xray 流量不经 TUN，无需预排除公网 DNS。
func platformPreExcludeAddrs() []netip.Prefix {
	return nil
}

// platformTunName 在 Linux 上返回自定义 TUN 接口名称。
func platformTunName() string {
	return "v2raya-tun"
}

// platformDisableAutoRoute 在 Linux 上始终返回 false：沿用调用方的 autoRoute 设置。
func platformDisableAutoRoute() bool {
	return false
}

// platformPostStart 在 TUN 启动完成后执行平台特定操作。
// Linux 无需额外处理。
func platformPostStart(_ []netip.Addr, _ string) {}

// platformPreClose 在 TUN 关闭前执行平台特定清理。
// Linux 无需额外处理。
func platformPreClose(_ string) {}
