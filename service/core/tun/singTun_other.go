//go:build !linux && !windows && !darwin
// +build !linux,!windows,!darwin

package tun

import "net/netip"

// platformPreExcludeAddrs 在 BSD 等其他平台上不预排除任何地址。
func platformPreExcludeAddrs() []netip.Prefix {
	return nil
}

// platformTunName 在 BSD 等平台上返回默认接口名称。
func platformTunName() string {
	return "v2raya-tun"
}

// platformDisableAutoRoute 在 BSD 等平台上不强制关闭 AutoRoute。
func platformDisableAutoRoute() bool {
	return false
}

// platformPostStart 在 BSD 等平台上无需额外操作。
func platformPostStart(_ []netip.Addr, _ string) {}

// platformPreClose 在 BSD 等平台上无需额外操作。
func platformPreClose(_ string) {}
