//go:build windows
// +build windows

package tun

import (
"fmt"
"net/netip"
"os/exec"
"strings"

"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Windows TUN 接口地址及网关（sing-tun 固定分配）
const (
tunAddr4    = "172.19.0.1"
tunGateway4 = "172.19.0.2"
)

var (
excludedRoutes       []netip.Prefix
tunDefaultRouteAdded bool
loopbackRouteAdded   bool
// 记录各阶段使用的接口别名，供清理时精确匹配
savedTunIfAlias      string
savedLoopbackIfAlias string
savedPhysIfAlias     string
)

// SetupTunRouteRules 在 Windows 上通过 netsh 手动建立 TUN 默认路由。
//
// sing-tun 的 AutoRoute 在 Windows 上可靠性不足，因此关闭 AutoRoute，
// 改由本函数添加以下路由（均使用 store=active，进程退出后自动失效）：
//
//   - 127.0.0.0/8 → 127.0.0.1  metric 0  （保证回环流量不经 TUN）
//   - 0.0.0.0/0   → 172.19.0.2 metric 1  （全局流量走 TUN）
func SetupTunRouteRules() error {
// --- 1. 回环路由：找到回环接口别名 ---
loopbackIf, err := getInterfaceAliasByIP("127.0.0.1")
if err != nil {
log.Warn("[TUN][Windows] 获取回环接口别名失败: %v，跳过回环路由", err)
} else {
savedLoopbackIfAlias = loopbackIf
out, err := exec.Command("netsh", "interface", "ipv4", "add", "route",
"127.0.0.0/8", loopbackIf,
"nexthop=127.0.0.1", "metric=0", "store=active").CombinedOutput()
if err != nil {
s := string(out)
if isAlreadyExists(s) {
log.Info("[TUN][Windows] 回环路由 127.0.0.0/8 已存在")
} else {
log.Warn("[TUN][Windows] 添加回环路由失败: %v, output: %s", err, s)
}
} else {
loopbackRouteAdded = true
log.Info("[TUN][Windows] 添加回环路由 127.0.0.0/8 接口=%s metric=0", loopbackIf)
}
}

// --- 2. 默认路由：找到 TUN 接口别名 ---
tunIf, err := getInterfaceAliasByIP(tunAddr4)
if err != nil {
log.Warn("[TUN][Windows] 获取 TUN 接口别名失败: %v", err)
return err
}
savedTunIfAlias = tunIf

out, err := exec.Command("netsh", "interface", "ipv4", "add", "route",
"0.0.0.0/0", tunIf,
"nexthop="+tunGateway4, "metric=1", "store=active").CombinedOutput()
if err != nil {
s := string(out)
if isAlreadyExists(s) {
tunDefaultRouteAdded = true
log.Info("[TUN][Windows] 默认路由 0.0.0.0/0 接口=%s 已存在", tunIf)
} else {
log.Warn("[TUN][Windows] 添加默认路由失败: %v, output: %s", err, s)
return err
}
} else {
tunDefaultRouteAdded = true
log.Info("[TUN][Windows] 添加默认路由 0.0.0.0/0 接口=%s nexthop=%s metric=1", tunIf, tunGateway4)
}
return nil
}

// CleanupTunRouteRules 删除 SetupTunRouteRules 添加的路由条目。
func CleanupTunRouteRules() error {
if loopbackRouteAdded && savedLoopbackIfAlias != "" {
out, err := exec.Command("netsh", "interface", "ipv4", "delete", "route",
"127.0.0.0/8", savedLoopbackIfAlias).CombinedOutput()
if err != nil {
log.Warn("[TUN][Windows] 删除回环路由失败: %v, output: %s", err, string(out))
} else {
log.Info("[TUN][Windows] 已删除回环路由 127.0.0.0/8 接口=%s", savedLoopbackIfAlias)
}
loopbackRouteAdded = false
}

if tunDefaultRouteAdded && savedTunIfAlias != "" {
out, err := exec.Command("netsh", "interface", "ipv4", "delete", "route",
"0.0.0.0/0", savedTunIfAlias).CombinedOutput()
if err != nil {
log.Warn("[TUN][Windows] 删除 TUN 默认路由失败: %v, output: %s", err, string(out))
} else {
log.Info("[TUN][Windows] 已删除默认路由 0.0.0.0/0 接口=%s", savedTunIfAlias)
}
tunDefaultRouteAdded = false
}
savedTunIfAlias = ""
savedLoopbackIfAlias = ""
return nil
}

// SetupExcludeRoutes 为代理服务端 IP 添加"绕过 TUN"的静态路由。
//
// Windows 没有 fwmark 机制，需为每个服务端 IP 显式添加经物理网关的路由，
// 避免代理流量被 TUN 再次捕获（回环）。
func SetupExcludeRoutes(addrs []netip.Prefix) error {
if len(addrs) == 0 {
return nil
}
excludedRoutes = addrs

gw4, err := getDefaultGateway()
if err != nil {
log.Warn("[TUN][Windows] 获取默认 IPv4 网关失败: %v", err)
return err
}

physIf, err := getPhysicalInterfaceAlias()
if err != nil {
log.Warn("[TUN][Windows] 获取物理接口别名失败: %v", err)
return err
}
savedPhysIfAlias = physIf

for _, prefix := range addrs {
addr := prefix.Addr()
if addr.Is4() {
// IPv4 排除路由：经物理接口直连
// metric=5：高于 TUN 默认路由(1)，确保这些 IP 走物理链路
out, err := exec.Command("netsh", "interface", "ipv4", "add", "route",
prefix.String(), physIf,
"nexthop="+gw4, "metric=5", "store=active").CombinedOutput()
if err != nil {
s := string(out)
if !isAlreadyExists(s) {
log.Warn("[TUN][Windows] 添加 IPv4 排除路由 %s 失败: %v, output: %s", prefix, err, s)
}
} else {
log.Info("[TUN][Windows] 添加 IPv4 排除路由 %s 接口=%s nexthop=%s metric=5", prefix, physIf, gw4)
}
} else {
// IPv6 排除路由
gw6, err := getDefaultGatewayIPv6()
if err != nil {
log.Warn("[TUN][Windows] 获取 IPv6 网关失败，跳过 %s: %v", prefix, err)
continue
}
physIf6, err := getPhysicalInterfaceAliasIPv6()
if err != nil {
log.Warn("[TUN][Windows] 获取 IPv6 物理接口失败，跳过 %s: %v", prefix, err)
continue
}
out, err := exec.Command("netsh", "interface", "ipv6", "add", "route",
prefix.String(), physIf6,
"nexthop="+gw6, "metric=5", "store=active").CombinedOutput()
if err != nil {
s := string(out)
if !isAlreadyExists(s) {
log.Warn("[TUN][Windows] 添加 IPv6 排除路由 %s 失败: %v, output: %s", prefix, err, s)
}
} else {
log.Info("[TUN][Windows] 添加 IPv6 排除路由 %s 接口=%s nexthop=%s metric=5", prefix, physIf6, gw6)
}
}
}
return nil
}

// CleanupExcludeRoutes 删除 SetupExcludeRoutes 添加的所有静态路由。
func CleanupExcludeRoutes() error {
physIf := savedPhysIfAlias
for _, prefix := range excludedRoutes {
addr := prefix.Addr()
if addr.Is4() {
var args []string
if physIf != "" {
args = []string{"interface", "ipv4", "delete", "route", prefix.String(), physIf}
} else {
args = []string{"interface", "ipv4", "delete", "route", prefix.String()}
}
out, err := exec.Command("netsh", args...).CombinedOutput()
if err != nil {
log.Warn("[TUN][Windows] 删除 IPv4 排除路由 %s 失败: %v, output: %s", prefix, err, string(out))
}
} else {
out, err := exec.Command("netsh", "interface", "ipv6", "delete", "route", prefix.String()).CombinedOutput()
if err != nil {
log.Warn("[TUN][Windows] 删除 IPv6 排除路由 %s 失败: %v, output: %s", prefix, err, string(out))
}
}
}
excludedRoutes = nil
savedPhysIfAlias = ""
return nil
}

// SetupTunDNS 通过 netsh 为 TUN 接口设置 DNS 服务器地址。
func SetupTunDNS(dnsServers []netip.Addr, tunName string) error {
if len(dnsServers) == 0 {
return nil
}

ifAlias, err := getTunInterfaceName(tunName)
if err != nil {
log.Warn("[TUN][Windows] SetupTunDNS: 获取接口别名失败: %v", err)
return err
}

// 分离 IPv4 / IPv6 DNS
var v4, v6 []string
for _, dns := range dnsServers {
if dns.Is4() {
v4 = append(v4, dns.String())
} else {
v6 = append(v6, dns.String())
}
}

// IPv4 DNS：先设置主 DNS，再依次添加备用
for i, addr := range v4 {
var out []byte
if i == 0 {
// set 设置主 DNS（同时清空旧配置）
out, err = exec.Command("netsh", "interface", "ipv4", "set", "dns",
"name="+ifAlias, "source=static", "address="+addr, "register=none").CombinedOutput()
} else {
// add 追加备用 DNS
out, err = exec.Command("netsh", "interface", "ipv4", "add", "dns",
"name="+ifAlias, "address="+addr, fmt.Sprintf("index=%d", i+1)).CombinedOutput()
}
if err != nil {
log.Warn("[TUN][Windows] 设置 IPv4 DNS[%d]=%s 失败: %v, output: %s", i, addr, err, string(out))
} else {
log.Info("[TUN][Windows] 接口 '%s' IPv4 DNS[%d] 已设置: %s", ifAlias, i, addr)
}
}

// IPv6 DNS
for i, addr := range v6 {
var out []byte
if i == 0 {
out, err = exec.Command("netsh", "interface", "ipv6", "set", "dns",
"name="+ifAlias, "source=static", "address="+addr, "register=none").CombinedOutput()
} else {
out, err = exec.Command("netsh", "interface", "ipv6", "add", "dns",
"name="+ifAlias, "address="+addr, fmt.Sprintf("index=%d", i+1)).CombinedOutput()
}
if err != nil {
log.Warn("[TUN][Windows] 设置 IPv6 DNS[%d]=%s 失败: %v, output: %s", i, addr, err, string(out))
} else {
log.Info("[TUN][Windows] 接口 '%s' IPv6 DNS[%d] 已设置: %s", ifAlias, i, addr)
}
}
return nil
}

// CleanupTunDNS 通过 netsh 将 TUN 接口的 DNS 恢复为 DHCP 自动获取。
func CleanupTunDNS(tunName string) error {
ifAlias, err := getTunInterfaceName(tunName)
if err != nil {
// 接口可能已被删除，不视为错误
return nil
}

out, err := exec.Command("netsh", "interface", "ipv4", "set", "dns",
"name="+ifAlias, "source=dhcp").CombinedOutput()
if err != nil {
log.Warn("[TUN][Windows] 重置接口 IPv4 DNS 失败: %v, output: %s", err, string(out))
} else {
log.Info("[TUN][Windows] 接口 '%s' IPv4 DNS 已重置为 DHCP", ifAlias)
}

out, err = exec.Command("netsh", "interface", "ipv6", "set", "dns",
"name="+ifAlias, "source=dhcp").CombinedOutput()
if err != nil {
log.Warn("[TUN][Windows] 重置接口 IPv6 DNS 失败: %v, output: %s", err, string(out))
} else {
log.Info("[TUN][Windows] 接口 '%s' IPv6 DNS 已重置为 DHCP", ifAlias)
}
return nil
}

// ── 内部辅助函数 ──────────────────────────────────────────────────────────────

// getInterfaceAliasByIP 通过 PowerShell Get-NetIPAddress 查找拥有指定 IPv4 地址的接口别名。
func getInterfaceAliasByIP(ip string) (string, error) {
psCmd := fmt.Sprintf(
"(Get-NetIPAddress -IPAddress '%s' -AddressFamily IPv4 -ErrorAction SilentlyContinue | Select-Object -First 1).InterfaceAlias",
ip)
out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
if err != nil {
return "", fmt.Errorf("查询 IP=%s 的接口失败: %w, output: %s", ip, err, string(out))
}
alias := strings.TrimSpace(string(out))
if alias == "" {
return "", fmt.Errorf("未找到拥有 IP=%s 的接口", ip)
}
return alias, nil
}

// getDefaultGateway 获取当前物理接口的默认 IPv4 网关（排除 TUN 网关）。
func getDefaultGateway() (string, error) {
psCmd := `(Get-NetRoute -DestinationPrefix '0.0.0.0/0' |` +
` Where-Object { $_.NextHop -ne '` + tunGateway4 + `' -and $_.NextHop -ne '0.0.0.0' } |` +
` Sort-Object InterfaceMetric |` +
` Select-Object -First 1).NextHop`
out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
if err != nil {
return "", fmt.Errorf("获取默认 IPv4 网关失败: %w, output: %s", err, string(out))
}
gw := strings.TrimSpace(string(out))
if gw == "" {
return "", fmt.Errorf("默认 IPv4 网关为空（可能无网络连接）")
}
return gw, nil
}

// getDefaultGatewayIPv6 获取当前物理接口的默认 IPv6 网关。
func getDefaultGatewayIPv6() (string, error) {
psCmd := `(Get-NetRoute -DestinationPrefix '::/0' |` +
` Where-Object { $_.NextHop -ne '::' } |` +
` Sort-Object InterfaceMetric |` +
` Select-Object -First 1).NextHop`
out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
if err != nil {
return "", fmt.Errorf("获取默认 IPv6 网关失败: %w, output: %s", err, string(out))
}
gw := strings.TrimSpace(string(out))
if gw == "" {
return "", fmt.Errorf("默认 IPv6 网关为空")
}
return gw, nil
}

// getPhysicalInterfaceAlias 获取承载默认 IPv4 路由的物理接口别名（排除 TUN）。
func getPhysicalInterfaceAlias() (string, error) {
psCmd := `(Get-NetRoute -DestinationPrefix '0.0.0.0/0' |` +
` Where-Object { $_.NextHop -ne '` + tunGateway4 + `' -and $_.NextHop -ne '0.0.0.0' } |` +
` Sort-Object InterfaceMetric |` +
` Select-Object -First 1).InterfaceAlias`
out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
if err != nil {
return "", fmt.Errorf("获取物理 IPv4 接口别名失败: %w, output: %s", err, string(out))
}
alias := strings.TrimSpace(string(out))
if alias == "" {
return "", fmt.Errorf("物理 IPv4 接口别名为空")
}
return alias, nil
}

// getPhysicalInterfaceAliasIPv6 获取承载默认 IPv6 路由的物理接口别名。
func getPhysicalInterfaceAliasIPv6() (string, error) {
psCmd := `(Get-NetRoute -DestinationPrefix '::/0' |` +
` Where-Object { $_.NextHop -ne '::' } |` +
` Sort-Object InterfaceMetric |` +
` Select-Object -First 1).InterfaceAlias`
out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
if err != nil {
return "", fmt.Errorf("获取物理 IPv6 接口别名失败: %w, output: %s", err, string(out))
}
alias := strings.TrimSpace(string(out))
if alias == "" {
return "", fmt.Errorf("物理 IPv6 接口别名为空")
}
return alias, nil
}

// getTunInterfaceName 通过 Get-NetAdapter 按名称模糊匹配 TUN 接口的完整别名。
func getTunInterfaceName(baseName string) (string, error) {
psCmd := fmt.Sprintf(
"(Get-NetAdapter | Where-Object { $_.Name -like '*%s*' -and $_.Status -ne 'Not Present' } | Select-Object -First 1).Name",
baseName)
out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
if err != nil {
return "", fmt.Errorf("查找接口失败: %w, output: %s", err, string(out))
}
name := strings.TrimSpace(string(out))
if name == "" {
return "", fmt.Errorf("未找到匹配接口: %s", baseName)
}
return name, nil
}

// isAlreadyExists 检查 netsh 输出是否表示"对象已存在"。
func isAlreadyExists(output string) bool {
return strings.Contains(output, "对象已存在") ||
strings.Contains(output, "already exists") ||
strings.Contains(output, "Element already exists")
}
