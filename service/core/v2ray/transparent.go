package v2ray

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/tun"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func deleteTransparentProxyRules() {
	iptables.CloseWatcher()
	if !conf.GetEnvironmentConfig().Lite {
		removeResolvHijacker()
		iptables.Tproxy.GetCleanCommands().Run(false)
		iptables.Redirect.GetCleanCommands().Run(false)
		iptables.DropSpoofing.GetCleanCommands().Run(false)
		tun.Default.Close()
	}
	iptables.SystemProxy.GetCleanCommands().Run(false)
	time.Sleep(30 * time.Millisecond)
}

func writeTransparentProxyRules(tmpl *Template) (err error) {
	defer func() {
		if err != nil {
			log.Warn("writeTransparentProxyRules: %v", err)
			deleteTransparentProxyRules()
		}
	}()
	if specialMode.ShouldUseSupervisor() {
		if err = iptables.DropSpoofing.GetSetupCommands().Run(true); err != nil {
			log.Warn("DropSpoofing can't be enable: %v", err)
			return err
		}
	}
	setting := configure.GetSettingNotNil()
	switch setting.TransparentType {
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
	case configure.TransparentGvisorTun, configure.TransparentSystemTun:
		mode, _, _ := strings.Cut(string(setting.TransparentType), "_")
		tun.Default.SetFakeIP(setting.TunFakeIP)
		tun.Default.SetIPv6(setting.TunIPv6)
		tun.Default.SetStrictRoute(setting.TunStrictRoute)
		tun.Default.SetAutoRoute(setting.TunAutoRoute)

		// Extract and resolve DNS servers from v2ray configuration
		// This prevents DNS server traffic from being intercepted by TUN, avoiding routing loops
		log.Info("[TUN] Extracting DNS servers from configuration...")
		var dnsHosts []string
		if tmpl != nil {
			dnsHosts = ExtractDnsServerHostsFromTemplate(tmpl)
		}
		if len(dnsHosts) > 0 {
			dnsExcludes := tun.ResolveDnsServersToExcludes(dnsHosts)
			for _, prefix := range dnsExcludes {
				tun.Default.AddIPWhitelist(prefix.Addr())
				log.Info("[TUN] Added DNS server IP to exclusion list: %s", prefix.Addr())
			}
		}

		// Add server addresses to exclusion list BEFORE starting TUN
		// Resolve domain names to IPs to ensure proper routing exclusion
		_, serverInfos, _ := getConnectedServerObjs()
		for _, info := range serverInfos {
			host := info.Info.GetHostname()
			if addr, err := netip.ParseAddr(host); err == nil {
				// Already an IP address
				tun.Default.AddIPWhitelist(addr)
			} else {
				// Domain name - resolve to IPs first
				log.Info("[TUN] Resolving server domain: %s", host)
				if ips, err := net.LookupIP(host); err == nil {
					log.Info("[TUN] Resolved %s to %d IP address(es)", host, len(ips))
					for _, ip := range ips {
						if addr, ok := netip.AddrFromSlice(ip); ok {
							tun.Default.AddIPWhitelist(addr)
						}
					}
					// Also add domain to whitelist for DNS queries
					tun.Default.AddDomainWhitelist(host)
				} else {
					log.Warn("[TUN] Failed to resolve server domain %s: %v, adding as domain whitelist", host, err)
					tun.Default.AddDomainWhitelist(host)
				}
			}
		}

		// Now start TUN with the exclusion list configured
		if err = tun.Default.Start(tun.Stack(mode)); err != nil {
			return fmt.Errorf("not support \"%s tun\" mode of transparent proxy: %w", mode, err)
		}
	case configure.TransparentSystemProxy:
		if err = iptables.SystemProxy.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("not support \"system proxy\" mode of transparent proxy: %w", err)
		}
	default:
		return fmt.Errorf("undefined \"%v\" mode of transparent proxy", setting.TransparentType)
	}

	if specialMode.ShouldLocalDnsListen() {
		if couldListenLocalhost, e := specialMode.CouldLocalDnsListen(); couldListenLocalhost {
			if e != nil {
				log.Warn("only listen at 127.2.0.17: %v", e)
			}
			resetResolvHijacker()
		} else if specialMode.ShouldUseFakeDns() {
			return fmt.Errorf("fakedns cannot be enabled: %w", e)
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
