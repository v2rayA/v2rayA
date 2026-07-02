package v2ray

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// cleanupResidualTransparentProxyRules cleans up any residual iptables/nftables rules
// that may have been left behind after an abnormal termination (e.g., kill -9, system crash, panic).
// It uses "2>/dev/null || true" to ensure no errors are raised if rules/chains don't exist.
func cleanupResidualTransparentProxyRules() {
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
	iptables -w 2 -t nat -D PREROUTING -p udp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
	iptables -w 2 -t nat -D PREROUTING -p tcp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
	iptables -w 2 -t nat -D OUTPUT -p udp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
	iptables -w 2 -t nat -D OUTPUT -p tcp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
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
	# 清理 IPv6 直接 REDIRECT 规则
	ip6tables -w 2 -t nat -D PREROUTING -p udp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
	ip6tables -w 2 -t nat -D PREROUTING -p tcp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
	ip6tables -w 2 -t nat -D OUTPUT -p udp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
	ip6tables -w 2 -t nat -D OUTPUT -p tcp --dport 53 -j REDIRECT --to-port 52353 2>/dev/null || true
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
	cmds.ExecCommands(commands, false)
}

func deleteTransparentProxyRulesKeepSystemProxy() {
	stopTinyTun()
	iptables.CloseWatcher()
	if !conf.GetEnvironmentConfig().Lite {
		removeResolvHijacker()
		iptables.Tproxy.GetCleanCommands().Run(false)
		iptables.Redirect.GetCleanCommands().Run(false)
		iptables.DropSpoofing.GetCleanCommands().Run(false)
	}
	time.Sleep(30 * time.Millisecond)
}

func deleteTransparentProxyRules() {
	deleteTransparentProxyRulesKeepSystemProxy()
	iptables.SystemProxy.GetCleanCommands().Run(false)
}

func writeTransparentProxyRules(tmpl *Template) (err error) {
	defer func() {
		if err != nil {
			log.Warn("writeTransparentProxyRules: %v", err)
			deleteTransparentProxyRules()
		}
	}()
	// v2raya-core 进程内启动 DNS 模块（监听 :52353），
	// v2rayA 负责在透明代理时应用 iptables/nftables 规则将 53 端口流量重定向到 52353。
	// 等待 DNS 模块就绪后应用防火墙规则。
	if tmpl != nil && tmpl.DnsModuleConfig != nil {
		dnsAddr := "127.2.0.17:52353"
		if tmpl.Setting != nil && tmpl.Setting.DnsListenAddr != "" {
			dnsAddr = tmpl.Setting.DnsListenAddr
		}
		if err := waitForDnsPort(dnsAddr, 5*time.Second); err != nil {
			return fmt.Errorf("dns module not ready: %w", err)
		}
		log.Trace("DNS module is ready on %s, setting up transparent proxy rules", dnsAddr)
	}
	cleanupResidualTransparentProxyRules()
	setting := configure.GetSettingNotNil()
	switch setting.TransparentType {
	case configure.TransparentTun:
		return startTinyTun(tmpl)
	case configure.TransparentTproxy:
		if err = iptables.Tproxy.GetSetupCommands().Run(true); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = fmt.Errorf("you does not compile xt_TPROXY in kernel")
			}
			return fmt.Errorf("not support \"tproxy\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(iptables.Tproxy)
	case configure.TransparentRedirect:
		if err = iptables.Redirect.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("not support \"redirect\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(iptables.Redirect)
	case configure.TransparentSystemProxy:
		if err = iptables.SystemProxy.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("not support \"system proxy\" mode of transparent proxy: %w", err)
		}
	default:
		return fmt.Errorf("undefined \"%v\" mode of transparent proxy", setting.TransparentType)
	}

	// 无论哪种透明代理模式，都用 nat 表的 REDIRECT 将 DNS 流量（:53）转到 DNS 模块（:52353）。
	// 同时拦截 OUTPUT（本地进程）和 PREROUTING（LAN 设备）的 DNS 查询。
	// TPROXY 模式对回环（loopback）流量的 TPROXY 拦截不可靠，而 REDIRECT 在 OUTPUT 链上稳定。
	if ShouldLocalDnsListen() {
		dnsRedirect := `
iptables -w 2 -t nat -I PREROUTING -p udp --dport 53 -j REDIRECT --to-port 52353
iptables -w 2 -t nat -I PREROUTING -p tcp --dport 53 -j REDIRECT --to-port 52353
iptables -w 2 -t nat -I OUTPUT -p udp --dport 53 -j REDIRECT --to-port 52353
iptables -w 2 -t nat -I OUTPUT -p tcp --dport 53 -j REDIRECT --to-port 52353
`
		if iptables.IsIPv6Supported() {
			dnsRedirect += `
ip6tables -w 2 -t nat -I PREROUTING -p udp --dport 53 -j REDIRECT --to-port 52353
ip6tables -w 2 -t nat -I PREROUTING -p tcp --dport 53 -j REDIRECT --to-port 52353
ip6tables -w 2 -t nat -I OUTPUT -p udp --dport 53 -j REDIRECT --to-port 52353
ip6tables -w 2 -t nat -I OUTPUT -p tcp --dport 53 -j REDIRECT --to-port 52353
`
		}
		cmds.ExecCommands(dnsRedirect, false)

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
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("udp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("dns port %s not reachable within %v", addr, timeout)
}
