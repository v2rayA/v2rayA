//go:build windows
// +build windows

package tun

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Windows TUN 接口地址及网关（sing-tun 固定分配）
const (
	tunAddr4    = "172.19.0.1"
	tunGateway4 = "172.19.0.2"
	tunAddr6    = "fdfe:dcba:9876::1"
	tunGateway6 = "fdfe:dcba:9876::2"
)

var (
	excludedRoutes        []netip.Prefix
	tunDefaultRouteAdded  bool
	tunDefaultRoute6Added bool
	loopbackRouteAdded    bool
	// 记录各阶段使用的接口别名，供清理时精确匹配
	savedTunIfAlias      string
	savedLoopbackIfAlias string
	savedPhysIfAlias     string
	savedTunIfAlias6     string
	// tunAutoRoute 记录当前是否由 sing-tun 自动管理路由
	// true  = AutoRoute 开启，sing-tun 通过 winipcfg 管理路由，手动设置均为空操作
	// false = AutoRoute 关闭，需手动 netsh 路由管理
	tunAutoRoute bool

	// dynExcludeSet 记录已经动态添加过绕过路由的 IP 地址，防止重复添加。
	dynExcludeSet sync.Map // key: netip.Addr (as string) → struct{}
	// dynExcludedRoutes 记录所有动态添加的排除路由，供 CleanupExcludeRoutes 清理。
	dynExcludedRoutes []netip.Prefix
	dynExcludeMu      sync.Mutex
	// 缓存物理接口和默认网关，避免每次动态路由时都调用 PowerShell。
	cachedPhysIf  string
	cachedGw4     string
	cachedPhysIf6 string
	cachedGw6     string
	cachedIfMu    sync.Mutex
)

// setTunRouteAutoMode 通知 Windows 路由模块当前是否处于 AutoRoute 模式。
// 源头 singTun.go Start() 在调用 SetupTunRouteRules 前调用。
func setTunRouteAutoMode(auto bool) {
	tunAutoRoute = auto
}

// SetupTunRouteRules 在 Windows 上通过 netsh 手动建立 TUN 默认路由。
//
// 当 AutoRoute 开启时，sing-tun 已通过 winipcfg 管理路由，本函数为空操作。
// 当 AutoRoute 关闭时（用户显式禁用），添加以下路由（均使用 store=active，进程退出后自动失效）：
//
//   - 127.0.0.0/8 → 127.0.0.1  metric 0  （保证回环流量不经 TUN）
//   - 0.0.0.0/0   → 172.19.0.2 metric 1  （全局流量走 TUN）
func SetupTunRouteRules() error {
	if tunAutoRoute {
		// sing-tun 的 AutoRoute 已通过 winipcfg 高效管理路由，无需手动干预
		log.Info("[TUN][Windows] AutoRoute 开启，sing-tun 管理路由，跳过手动 SetupTunRouteRules")
		return nil
	}
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

	// --- 3. IPv6 默认路由（仅当 TUN 有 IPv6 地址时） ---
	if tunIf6, err := getInterfaceAliasByIPv6(tunAddr6); err != nil {
		log.Info("[TUN][Windows] 未检测到 TUN IPv6 地址，跳过 IPv6 默认路由: %v", err)
	} else {
		out, err = exec.Command("netsh", "interface", "ipv6", "add", "route",
			"::/0", tunIf6,
			"nexthop="+tunGateway6, "metric=1", "store=active").CombinedOutput()
		if err != nil {
			s := string(out)
			if isAlreadyExists(s) {
				tunDefaultRoute6Added = true
				log.Info("[TUN][Windows] 默认 IPv6 路由 ::/0 接口=%s 已存在", tunIf6)
			} else {
				log.Warn("[TUN][Windows] 添加 IPv6 默认路由失败: %v, output: %s", err, s)
				return err
			}
		} else {
			tunDefaultRoute6Added = true
			log.Info("[TUN][Windows] 添加默认 IPv6 路由 ::/0 接口=%s nexthop=%s metric=1", tunIf6, tunGateway6)
		}
	}
	return nil
}

// CleanupTunRouteRules 删除 SetupTunRouteRules 添加的路由条目。
func CleanupTunRouteRules() error {
	if tunAutoRoute {
		return nil // sing-tun 在 Close() 时自行清除路由
	}
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
	if tunDefaultRoute6Added && savedTunIfAlias6 != "" {
		out, err := exec.Command("netsh", "interface", "ipv6", "delete", "route",
			"::/0", savedTunIfAlias6).CombinedOutput()
		if err != nil {
			log.Warn("[TUN][Windows] 删除 TUN IPv6 默认路由失败: %v, output: %s", err, string(out))
		} else {
			log.Info("[TUN][Windows] 已删除默认 IPv6 路由 ::/0 接口=%s", savedTunIfAlias6)
		}
		tunDefaultRoute6Added = false
	}
	savedTunIfAlias = ""
	savedTunIfAlias6 = ""
	savedLoopbackIfAlias = ""
	return nil
}

// SetupExcludeRoutes 为代理服务端 IP 添加"绕过 TUN"的静态路由。
//
// 当 AutoRoute 开启时，sing-tun 已通过 Inet4RouteExcludeAddress 处理，无需单独添加。
// 当 AutoRoute 关闭时，Windows 没有 fwmark 机制，需为每个服务端 IP 显式添加经物理网关的路由，
// 避免代理流量被 TUN 再次捕获（回环）。
func SetupExcludeRoutes(addrs []netip.Prefix) error {
	if tunAutoRoute {
		log.Info("[TUN][Windows] AutoRoute 开启，Inet4RouteExcludeAddress 已处理排除，跳过 SetupExcludeRoutes")
		return nil
	}
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

// CleanupExcludeRoutes 删除 SetupExcludeRoutes 和 DynAddExcludeRoute 添加的所有静态路由。
func CleanupExcludeRoutes() error {
	if tunAutoRoute {
		return nil
	}
	physIf := savedPhysIfAlias
	if physIf == "" {
		if alias, err := getPhysicalInterfaceAlias(); err == nil {
			physIf = alias
		} else {
			log.Warn("[TUN][Windows] Skip fetching physical interface alias during cleanup: %v", err)
		}
	}
	for _, prefix := range excludedRoutes {
		addr := prefix.Addr()
		if addr.Is4() {
			args := []string{"interface", "ipv4", "delete", "route", prefix.String()}
			if physIf != "" {
				args = append(args, physIf)
			}
			out, err := exec.Command("netsh", args...).CombinedOutput()
			if err != nil {
				log.Warn("[TUN][Windows] Failed to delete IPv4 exclude route %s: %v, output: %s", prefix, err, string(out))
			}
		} else {
			out, err := exec.Command("netsh", "interface", "ipv6", "delete", "route", prefix.String()).CombinedOutput()
			if err != nil {
				log.Warn("[TUN][Windows] Failed to delete IPv6 exclude route %s: %v, output: %s", prefix, err, string(out))
			}
		}
	}
	excludedRoutes = nil
	savedPhysIfAlias = ""

	// Clean up dynamic routes as well
	cleanupDynExcludeRoutes()
	return nil
}

// cleanupDynExcludeRoutes 删除所有 DynAddExcludeRoute 动态添加的路由并重置状态。
func cleanupDynExcludeRoutes() {
	dynExcludeMu.Lock()
	routes := dynExcludedRoutes
	physIf := cachedPhysIf
	dynExcludedRoutes = nil
	dynExcludeMu.Unlock()
	dynExcludeSet.Range(func(k, _ any) bool {
		dynExcludeSet.Delete(k)
		return true
	})
	// 重置缓存的物理接口信息
	cachedIfMu.Lock()
	cachedPhysIf = ""
	cachedGw4 = ""
	cachedPhysIf6 = ""
	cachedGw6 = ""
	cachedIfMu.Unlock()

	if physIf == "" {
		if alias, err := getPhysicalInterfaceAlias(); err == nil {
			physIf = alias
		} else {
			log.Warn("[TUN][Windows] Skip fetching physical interface alias during dynamic route cleanup: %v", err)
		}
	}

	for _, prefix := range routes {
		if prefix.Addr().Is4() {
			args := []string{"interface", "ipv4", "delete", "route", prefix.String()}
			if physIf != "" {
				args = append(args, physIf)
			}
			exec.Command("netsh", args...).Run() //nolint:errcheck
		} else {
			exec.Command("netsh", "interface", "ipv6", "delete", "route", prefix.String()).Run() //nolint:errcheck
		}
	}
}

// DynAddExcludeRoute 在运行时为指定 IP 添加一条经物理接口直连的静态路由，
// 用于防止 TUN 处理器的"直连"拨号因 TUN 接口路由而再次进入 TUN，导致路由回环。
//
// 该函数幂等（同一个 IP 只添加一次），路由通过 goroutine 异步添加，
// 不阻塞连接处理主路径。在 TUN 关闭时由 CleanupExcludeRoutes 统一清理。
func DynAddExcludeRoute(addr netip.Addr) {
	addr = addr.Unmap()
	if !addr.IsValid() || addr.IsLoopback() || addr.IsUnspecified() {
		return
	}
	key := addr.String()
	if _, loaded := dynExcludeSet.LoadOrStore(key, struct{}{}); loaded {
		// 已经添加过
		return
	}
	prefix := netip.PrefixFrom(addr, addr.BitLen())
	dynExcludeMu.Lock()
	dynExcludedRoutes = append(dynExcludedRoutes, prefix)
	dynExcludeMu.Unlock()

	go func() {
		// 获取或缓存物理网关/接口
		cachedIfMu.Lock()
		gw4 := cachedGw4
		physIf := cachedPhysIf
		gw6 := cachedGw6
		physIf6 := cachedPhysIf6
		cachedIfMu.Unlock()

		if addr.Is4() {
			if gw4 == "" || physIf == "" {
				var err error
				gw4, err = getDefaultGateway()
				if err != nil {
					log.Warn("[TUN][Windows] DynAddExcludeRoute: 获取 IPv4 网关失败: %v", err)
					return
				}
				physIf, err = getPhysicalInterfaceAlias()
				if err != nil {
					log.Warn("[TUN][Windows] DynAddExcludeRoute: 获取物理接口失败: %v", err)
					return
				}
				cachedIfMu.Lock()
				cachedGw4 = gw4
				cachedPhysIf = physIf
				cachedIfMu.Unlock()
			}
			out, err := exec.Command("netsh", "interface", "ipv4", "add", "route",
				prefix.String(), physIf, "nexthop="+gw4, "metric=5", "store=active").CombinedOutput()
			if err != nil {
				if !isAlreadyExists(string(out)) {
					log.Warn("[TUN][Windows] DynAddExcludeRoute: 添加 IPv4 路由 %s 失败: %v", prefix, err)
				}
			} else {
				log.Info("[TUN][Windows] DynAddExcludeRoute: 已添加 IPv4 路由 %s → %s (%s)", prefix, gw4, physIf)
			}
		} else {
			if gw6 == "" || physIf6 == "" {
				var err error
				gw6, err = getDefaultGatewayIPv6()
				if err != nil {
					log.Warn("[TUN][Windows] DynAddExcludeRoute: 获取 IPv6 网关失败: %v", err)
					return
				}
				physIf6, err = getPhysicalInterfaceAliasIPv6()
				if err != nil {
					log.Warn("[TUN][Windows] DynAddExcludeRoute: 获取 IPv6 物理接口失败: %v", err)
					return
				}
				cachedIfMu.Lock()
				cachedGw6 = gw6
				cachedPhysIf6 = physIf6
				cachedIfMu.Unlock()
			}
			out, err := exec.Command("netsh", "interface", "ipv6", "add", "route",
				prefix.String(), physIf6, "nexthop="+gw6, "metric=5", "store=active").CombinedOutput()
			if err != nil {
				if !isAlreadyExists(string(out)) {
					log.Warn("[TUN][Windows] DynAddExcludeRoute: 添加 IPv6 路由 %s 失败: %v", prefix, err)
				}
			} else {
				log.Info("[TUN][Windows] DynAddExcludeRoute: 已添加 IPv6 路由 %s → %s (%s)", prefix, gw6, physIf6)
			}
		}
	}()
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

// getInterfaceAliasByIPv6 通过 PowerShell 查找拥有指定 IPv6 地址的接口别名。
func getInterfaceAliasByIPv6(ip string) (string, error) {
	psCmd := fmt.Sprintf(
		"(Get-NetIPAddress -IPAddress '%s' -AddressFamily IPv6 -ErrorAction SilentlyContinue | Select-Object -First 1).InterfaceAlias",
		ip)
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("查询 IPv6=%s 的接口失败: %w, output: %s", ip, err, string(out))
	}
	alias := strings.TrimSpace(string(out))
	if alias == "" {
		return "", fmt.Errorf("未找到拥有 IPv6=%s 的接口", ip)
	}
	return alias, nil
}

// isAlreadyExists 检查 netsh 输出是否表示"对象已存在"。
func isAlreadyExists(output string) bool {
	return strings.Contains(output, "对象已存在") ||
		strings.Contains(output, "already exists") ||
		strings.Contains(output, "Element already exists")
}
