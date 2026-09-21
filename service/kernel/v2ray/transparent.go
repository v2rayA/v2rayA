package v2ray

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/bounddevice"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// hostMu serializes host changes with teardown across process generations.
var hostMu sync.Mutex

// RecoverHostState restores a pending startup's host changes without runtime state.
func RecoverHostState(state *configure.HostState) error {
	hostMu.Lock()
	defer hostMu.Unlock()
	iptables.CloseWatcher()
	var errs []error
	if state.TransparentType == configure.TransparentTun {
		errs = append(errs, recoverTunHostState(state))
	}
	if !conf.GetEnvironmentConfig().Lite {
		errs = append(errs, cleanupResidualTransparentProxyRules())
	}
	hijackerMu.Lock()
	if hijacker != nil {
		_ = hijacker.Close()
		hijacker = nil
	}
	if _, err := os.Lstat(resolvBackupPath); err == nil {
		if !restoreResolv() {
			errs = append(errs, fmt.Errorf("could not restore resolver backup %s", resolvBackupPath))
		}
	} else if !os.IsNotExist(err) {
		errs = append(errs, err)
	}
	hijackerMu.Unlock()
	if state.TransparentType == configure.TransparentSystemProxy {
		var snapshot interface{}
		found, err := configure.GetSystemProxySnapshot(&snapshot)
		if err != nil {
			errs = append(errs, err)
		} else if found {
			errs = append(errs, iptables.SystemProxy.GetCleanCommands().Run(true))
		}
	}
	return errors.Join(errs...)
}

var boundDeviceGuard struct {
	sync.Mutex
	guard *bounddevice.Guard
}

func stopBoundDeviceGuard() {
	boundDeviceGuard.Lock()
	defer boundDeviceGuard.Unlock()
	if err := boundDeviceGuard.guard.Close(); err != nil {
		log.Warn("detach bound-device REDIRECT bypass: %v", err)
	}
	boundDeviceGuard.guard = nil
}

func startBoundDeviceGuard() error {
	boundDeviceGuard.Lock()
	defer boundDeviceGuard.Unlock()
	if boundDeviceGuard.guard != nil {
		return nil
	}
	guard, err := bounddevice.Start("/sys/fs/cgroup")
	if err != nil {
		return err
	}
	boundDeviceGuard.guard = guard
	return nil
}

// cleanupResidualTransparentProxyRules cleans up any residual iptables/nftables rules
// that may have been left behind after an abnormal termination (e.g., kill -9, system crash, panic).
// It uses "2>/dev/null || true" to ensure no errors are raised if rules/chains don't exist.
func cleanupResidualTransparentProxyRules() error {
	tunCleanupResidual()
	commands := `
# 清理 DNS_MARK 链（TProxy 模式）
iptables -w 2 -t mangle -F DNS_MARK 2>/dev/null || true
iptables -w 2 -t mangle -D PREROUTING -p udp --dport 53 -j DNS_MARK 2>/dev/null || true
iptables -w 2 -t mangle -D PREROUTING -p tcp --dport 53 -j DNS_MARK 2>/dev/null || true
iptables -w 2 -t mangle -D OUTPUT -p udp --dport 53 -j DNS_MARK 2>/dev/null || true
iptables -w 2 -t mangle -D OUTPUT -p tcp --dport 53 -j DNS_MARK 2>/dev/null || true
iptables -w 2 -t mangle -X DNS_MARK 2>/dev/null || true
	# 清理 DNS_REDIRECT 链（Redirect 模式）
	iptables -w 2 -t nat -F DNS_REDIRECT 2>/dev/null || true
	iptables -w 2 -t nat -D PREROUTING -p udp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	iptables -w 2 -t nat -D PREROUTING -p tcp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	iptables -w 2 -t nat -D OUTPUT -p udp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	iptables -w 2 -t nat -D OUTPUT -p tcp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	iptables -w 2 -t nat -X DNS_REDIRECT 2>/dev/null || true
	# 清理直接 REDIRECT 规则（所有透明代理模式通用的 DNS 重定向）
	# 清理 mark 0x80 豁免规则（配合上述 REDIRECT 的防回环豁免，成对删除避免重复累积）
	iptables -w 2 -t nat -D OUTPUT -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
	iptables -w 2 -t nat -D PREROUTING -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
# IPv6 清理
ip6tables -w 2 -t mangle -F DNS_MARK 2>/dev/null || true
ip6tables -w 2 -t mangle -D PREROUTING -p udp --dport 53 -j DNS_MARK 2>/dev/null || true
ip6tables -w 2 -t mangle -D PREROUTING -p tcp --dport 53 -j DNS_MARK 2>/dev/null || true
ip6tables -w 2 -t mangle -D OUTPUT -p udp --dport 53 -j DNS_MARK 2>/dev/null || true
ip6tables -w 2 -t mangle -D OUTPUT -p tcp --dport 53 -j DNS_MARK 2>/dev/null || true
ip6tables -w 2 -t mangle -X DNS_MARK 2>/dev/null || true
	ip6tables -w 2 -t nat -F DNS_REDIRECT 2>/dev/null || true
	ip6tables -w 2 -t nat -D PREROUTING -p udp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	ip6tables -w 2 -t nat -D PREROUTING -p tcp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	ip6tables -w 2 -t nat -D OUTPUT -p udp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	ip6tables -w 2 -t nat -D OUTPUT -p tcp --dport 53 -j DNS_REDIRECT 2>/dev/null || true
	ip6tables -w 2 -t nat -X DNS_REDIRECT 2>/dev/null || true
	# 清理 IPv6 直接 REDIRECT 规则` + dnsRedirectDeleteCommands() + `
	# 清理 IPv6 mark 0x80 豁免规则
	ip6tables -w 2 -t nat -D OUTPUT -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
	ip6tables -w 2 -t nat -D PREROUTING -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
# 清理 TProxy 链
iptables -w 2 -t mangle -F TP_OUT 2>/dev/null || true
iptables -w 2 -t mangle -D OUTPUT -j TP_OUT 2>/dev/null || true
iptables -w 2 -t mangle -X TP_OUT 2>/dev/null || true
iptables -w 2 -t mangle -F TP_PRE 2>/dev/null || true
iptables -w 2 -t mangle -D PREROUTING -j TP_PRE 2>/dev/null || true
iptables -w 2 -t mangle -X TP_PRE 2>/dev/null || true
iptables -w 2 -t mangle -F TP_RULE 2>/dev/null || true
iptables -w 2 -t mangle -X TP_RULE 2>/dev/null || true
iptables -w 2 -t mangle -F TP_MARK 2>/dev/null || true
iptables -w 2 -t mangle -X TP_MARK 2>/dev/null || true
# 清理 Redirect 链
iptables -w 2 -t nat -F TP_OUT 2>/dev/null || true
iptables -w 2 -t nat -D OUTPUT -j TP_OUT 2>/dev/null || true
iptables -w 2 -t nat -X TP_OUT 2>/dev/null || true
iptables -w 2 -t nat -F TP_PRE 2>/dev/null || true
iptables -w 2 -t nat -D PREROUTING -j TP_PRE 2>/dev/null || true
iptables -w 2 -t nat -X TP_PRE 2>/dev/null || true
iptables -w 2 -t nat -F TP_RULE 2>/dev/null || true
iptables -w 2 -t nat -X TP_RULE 2>/dev/null || true
iptables -w 2 -F DROP_SPOOFING 2>/dev/null || true
iptables -w 2 -D INPUT -j DROP_SPOOFING 2>/dev/null || true
iptables -w 2 -D FORWARD -j DROP_SPOOFING 2>/dev/null || true
iptables -w 2 -X DROP_SPOOFING 2>/dev/null || true
ip rule del fwmark 0x40/0xc0 table 100 2>/dev/null || true
ip route del local 0.0.0.0/0 dev lo table 100 2>/dev/null || true
nft delete table inet v2raya 2>/dev/null || true
`
	return iptables.Setter{Cmds: commands}.Run(true)
}

// dnsRedirectPorts remembers every port the DNS REDIRECT rules were installed
// for during this process, so a change of DnsListenAddr between two runs
// does not leave the old port's rules behind: the settings the cleanup
// reads have already changed by then.
var (
	dnsRedirectPortsMu sync.Mutex
	dnsRedirectPorts   = map[string]struct{}{"52353": {}}
)

func rememberDnsRedirectPort(port string) {
	dnsRedirectPortsMu.Lock()
	dnsRedirectPorts[port] = struct{}{}
	dnsRedirectPortsMu.Unlock()
}

// dnsRedirectDeleteCommands deletes the DNS REDIRECT rules for the historical
// default, the port the module listens on now, and every port installed
// earlier in this process.
func dnsRedirectDeleteCommands() string {
	dnsRedirectPortsMu.Lock()
	ports := make([]string, 0, len(dnsRedirectPorts)+1)
	for p := range dnsRedirectPorts {
		ports = append(ports, p)
	}
	dnsRedirectPortsMu.Unlock()
	if p := dnsModulePort(configure.GetSettingNotNil()); !slices.Contains(ports, p) {
		ports = append(ports, p)
	}
	slices.Sort(ports)
	var b strings.Builder
	for _, port := range ports {
		for _, bin := range []string{"iptables", "ip6tables"} {
			for _, chain := range []string{"PREROUTING", "OUTPUT"} {
				for _, proto := range []string{"udp", "tcp"} {
					fmt.Fprintf(&b, "%s -w 2 -t nat -D %s -p %s --dport 53 -j REDIRECT --to-port %s 2>/dev/null || true\n", bin, chain, proto, port)
				}
			}
		}
	}
	return b.String()
}

// cleanDnsRedirectRules removes the direct nat OUTPUT/PREROUTING DNS
// redirect rules (REDIRECT --to-port 52353) and the fwmark 0x80 exemption
// rules that writeTransparentProxyRules adds with -A/-I. Tproxy/Redirect
// GetCleanCommands() only clean their own chains, so without this the
// rules survive a normal proxy stop: local DNS queries keep being hijacked
// to :52353, and once the DNS module exits they hit a dead port and DNS
// fails. Idempotent (2>/dev/null || true).
func cleanDnsRedirectRules() {
	commands := dnsRedirectDeleteCommands() + `
iptables -w 2 -t nat -D OUTPUT -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
iptables -w 2 -t nat -D PREROUTING -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
ip6tables -w 2 -t nat -D OUTPUT -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
ip6tables -w 2 -t nat -D PREROUTING -m mark --mark 0x80/0x80 -j RETURN 2>/dev/null || true
`
	iptables.Setter{Cmds: commands}.Run(false)
}

func deleteTransparentProxyRulesKeepSystemProxy() {
	stopTunCore()
	iptables.CloseWatcher()
	if !conf.GetEnvironmentConfig().Lite {
		removeResolvHijacker()
		cleanDnsRedirectRules()
		iptables.Tproxy.GetCleanCommands().Run(false)
		iptables.Redirect.GetCleanCommands().Run(false)
		iptables.DropSpoofing.GetCleanCommands().Run(false)
	}
	stopBoundDeviceGuard()
	time.Sleep(30 * time.Millisecond)
}

func deleteTransparentProxyRules() {
	deleteTransparentProxyRulesKeepSystemProxy()
	var snapshot interface{}
	found, err := configure.GetSystemProxySnapshot(&snapshot)
	if err != nil {
		log.Warn("read original system proxy: %v", err)
		return
	}
	if !found {
		return
	}
	if err := iptables.SystemProxy.GetCleanCommands().Run(true); err != nil {
		log.Warn("restore system proxy: %v", err)
	}
}

func waitForTransparentDNS(tmpl *Template) {
	if tmpl != nil && tmpl.DnsModuleConfig != nil {
		dnsAddr := tunDnsTarget(tmpl.Setting)
		if err := waitForDnsPort(dnsAddr, 5*time.Second); err != nil {
			// The probe resolves a name, so a dead or slow upstream fails it
			// even though the listener is up. Waiting is worth it when DNS is
			// healthy, but it must not be the reason the core cannot start.
			log.Warn("DNS module did not answer on %s yet, applying transparent proxy rules anyway: %v", dnsAddr, err)
		} else {
			log.Trace("DNS module is ready on %s, setting up transparent proxy rules", dnsAddr)
		}
	}
}

func writeTransparentProxyRules(tmpl *Template) (err error) {
	defer func() {
		if err != nil {
			log.Warn("writeTransparentProxyRules: %v", err)
			deleteTransparentProxyRules()
			err = common.Coded("TRANSPARENT_SETUP_FAILED", err, map[string]interface{}{
				"mode":   tmpl.Setting.TransparentType,
				"detail": err.Error(),
			})
		}
	}()
	cleanupResidualTransparentProxyRules()
	stopBoundDeviceGuard()
	setting := tmpl.Setting
	switch setting.TransparentType {
	case configure.TransparentTun:
		if err = startTunCore(tmpl); err != nil {
			return fmt.Errorf("could not set up transparent proxy in tun mode: %w", err)
		}
		if runtime.GOOS != "linux" || !setting.TunAutoRoute {
			// DNS is handled by the system-resolver setting on the TUN
			// interface; the REDIRECT rules and resolv.conf below are Linux.
			// With automatic routing off the user owns the network setup,
			// DNS included, as with the previous implementation.
			return nil
		}
	case configure.TransparentTproxy:
		if err = iptables.Tproxy.GetSetupCommands().Run(true); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = fmt.Errorf("the kernel has no xt_TPROXY module; load it or switch transparent proxy to redirect mode")
			}
			return fmt.Errorf("could not set up transparent proxy in tproxy mode: %w", err)
		}
		iptables.SetWatcher(iptables.Tproxy)
	case configure.TransparentRedirect:
		if conf.GetEnvironmentConfig().RedirectBoundDevice {
			if err = startBoundDeviceGuard(); err != nil {
				return fmt.Errorf("cannot enable bound-device REDIRECT bypass: %w", err)
			}
		}
		if err = iptables.Redirect.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("could not set up transparent proxy in redirect mode: %w", err)
		}
		iptables.SetWatcher(iptables.Redirect)
	case configure.TransparentSystemProxy:
		if err = iptables.SystemProxy.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("could not set up transparent proxy in system proxy mode: %w", err)
		}
	default:
		return fmt.Errorf("unknown transparent proxy mode %q; expected tproxy, redirect, system_proxy, or tun", setting.TransparentType)
	}

	// 无论哪种透明代理模式，都用 nat 表的 REDIRECT 将 DNS 流量（:53）转到 DNS 模块（:52353）。
	// 同时拦截 OUTPUT（本地进程）和 PREROUTING（LAN 设备）的 DNS 查询。
	// TPROXY 模式对回环（loopback）流量的 TPROXY 拦截不可靠，而 REDIRECT 在 OUTPUT 链上稳定。
	//
	// IMPORTANT (fix): mark 0x80 豁免规则必须排在 REDIRECT 规则之前，否则 v2raya-core
	// 自己向上游转发的 DNS 查询（socket 带 SO_MARK=0x80）会被自己的 REDIRECT 规则劫持回
	// :52353，形成无限回环（内存雪崩直至 OOM）。iptables 按顺序匹配：
	//   - REDIRECT 用 -A（追加到链尾），确保在 mark 豁免之后
	//   - mark 豁免用 -I（插入到链首），确保最先匹配
	if ShouldLocalDnsListen() {
		dnsPort := dnsModulePort(setting)
		rememberDnsRedirectPort(dnsPort)
		dnsRedirect := `
iptables -w 2 -t nat -A OUTPUT -p udp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
iptables -w 2 -t nat -A OUTPUT -p tcp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
iptables -w 2 -t nat -A PREROUTING -p udp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
iptables -w 2 -t nat -A PREROUTING -p tcp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
iptables -w 2 -t nat -I OUTPUT -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t nat -I PREROUTING -m mark --mark 0x80/0x80 -j RETURN
`
		if iptables.IsIPv6Supported() {
			dnsRedirect += `
ip6tables -w 2 -t nat -A OUTPUT -p udp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
ip6tables -w 2 -t nat -A OUTPUT -p tcp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
ip6tables -w 2 -t nat -A PREROUTING -p udp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
ip6tables -w 2 -t nat -A PREROUTING -p tcp --dport 53 -j REDIRECT --to-port ` + dnsPort + `
ip6tables -w 2 -t nat -I OUTPUT -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t nat -I PREROUTING -m mark --mark 0x80/0x80 -j RETURN
`
		}
		iptables.Setter{Cmds: dnsRedirect}.Run(false)

		if couldListenLocalhost, e := CouldLocalDnsListen(); couldListenLocalhost {
			if e != nil {
				log.Warn("only listen at 127.2.0.17: %v", e)
			}
			resetResolvHijacker()
		} else {
			log.Warn("writeTransparentProxyRules: %v", e)
		}
	}
	return nil
}

func IsTransparentOn(setting *configure.Setting) bool {
	if setting == nil {
		setting = configure.GetSettingNotNil()
	}
	if setting.Transparent == configure.TransparentClose {
		return false
	}
	if conf.GetEnvironmentConfig().Lite &&
		(setting.TransparentType == configure.TransparentTproxy ||
			setting.TransparentType == configure.TransparentRedirect) {
		return false
	}
	return true
}

// waitForDnsPort polls the DNS module's listening port until it's ready or a timeout expires.
// This ensures the v2raya-core DNS module is accepting queries before we apply firewall rules.
func waitForDnsPort(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	request := new(dns.Msg)
	request.SetQuestion("localhost.", dns.TypeA)
	client := &dns.Client{
		Net:     "udp",
		Timeout: 500 * time.Millisecond,
	}
	for time.Now().Before(deadline) {
		if remaining := time.Until(deadline); remaining < client.Timeout {
			client.Timeout = remaining
		}
		if _, _, err := client.Exchange(request, addr); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("dns port %s not reachable within %v", addr, timeout)
}
