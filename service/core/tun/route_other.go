//go:build !linux && !windows && !darwin
// +build !linux,!windows,!darwin

package tun

import (
"fmt"
"net/netip"
"os/exec"
"strings"

"github.com/v2rayA/v2rayA/pkg/util/log"
)

var excludedRoutes []netip.Prefix

// SetupTunRouteRules 在 FreeBSD/OpenBSD 上为空操作。
// 这些平台暂不支持策略路由规则配置。
func SetupTunRouteRules() error {
return nil
}

// CleanupTunRouteRules 在 FreeBSD/OpenBSD 上为空操作。
func CleanupTunRouteRules() error {
return nil
}

// SetupExcludeRoutes 在 FreeBSD/OpenBSD 上为代理服务端添加绕过 TUN 的静态路由。
// BSD 系统的 route 命令语法与 macOS 基本一致。
func SetupExcludeRoutes(addrs []netip.Prefix) error {
if len(addrs) == 0 {
return nil
}
excludedRoutes = addrs

gw, err := getDefaultGateway()
if err != nil {
log.Warn("[TUN][BSD] 获取默认网关失败: %v", err)
return err
}

for _, prefix := range addrs {
addr := prefix.Addr()
var out []byte
if addr.Is4() {
out, err = exec.Command("route", "add", "-host", addr.String(), gw).CombinedOutput()
} else {
out, err = exec.Command("route", "add", "-inet6", prefix.String(), gw).CombinedOutput()
}
if err != nil {
s := string(out)
if !strings.Contains(s, "File exists") && !strings.Contains(s, "already exists") {
log.Warn("[TUN][BSD] 添加排除路由 %s 失败: %v, output: %s", addr, err, s)
}
} else {
log.Info("[TUN][BSD] 添加排除路由 %s → %s", addr, gw)
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
log.Warn("[TUN][BSD] 删除排除路由 %s 失败: %v, output: %s", addr, err, string(out))
}
}
excludedRoutes = nil
return nil
}

// SetupTunDNS 在 FreeBSD/OpenBSD 上暂不支持自动 DNS 配置，为空操作。
func SetupTunDNS(_ []netip.Addr, _ string) error {
return nil
}

// CleanupTunDNS 在 FreeBSD/OpenBSD 上为空操作。
func CleanupTunDNS(_ string) error {
return nil
}

// getDefaultGateway 通过 netstat -rn 获取 BSD 系统的默认 IPv4 网关。
func getDefaultGateway() (string, error) {
// FreeBSD/OpenBSD: netstat -rn 中 Destination=default 那行
out, err := exec.Command("sh", "-c", "netstat -rn 2>/dev/null | awk '/^default/{print $2; exit}'").CombinedOutput()
if err != nil {
return "", fmt.Errorf("获取默认网关失败: %w, output: %s", err, string(out))
}
gw := strings.TrimSpace(string(out))
if gw == "" {
return "", fmt.Errorf("默认网关为空（可能无网络连接）")
}
return gw, nil
}
