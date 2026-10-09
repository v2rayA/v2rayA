package iptables

import "fmt"

// The DNS fragments below are what the opt-out removes. Without it port 53 is
// neither marked nor TPROXY'd to the DNS module, and it is returned from
// TP_RULE so it never reaches the generic rules. Each takes the iptables
// binary name, so IPv4 and IPv6 stay one set of rules.

// legacyDnsMarkChain creates the chain the OUTPUT and PREROUTING hooks jump
// into before the generic transparent rules.
func legacyDnsMarkChain(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return bin + ` -w 2 -t mangle -N DNS_MARK
`
}

// legacyDnsMarkHooks are the -I rules, inserted at the head of OUTPUT and
// PREROUTING ahead of the generic transparent rules.
func legacyDnsMarkHooks(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`# DNS 规则必须在透明代理规则之前插入（环路保护）
%s -w 2 -t mangle -I OUTPUT -p udp --dport 53 -j DNS_MARK
%s -w 2 -t mangle -I OUTPUT -p tcp --dport 53 -j DNS_MARK
%s -w 2 -t mangle -I PREROUTING -p udp --dport 53 -j DNS_MARK
%s -w 2 -t mangle -I PREROUTING -p tcp --dport 53 -j DNS_MARK
`, bin, bin, bin, bin)
}

// legacyDnsMarkChainRules fills the chain: the loop guard and the mark the
// TPROXY rules below test for.
func legacyDnsMarkChainRules(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`# DNS_MARK 链：环路保护 + 标记 DNS 流量
%s -w 2 -t mangle -A DNS_MARK -m mark --mark 0x80/0x80 -j RETURN
%s -w 2 -t mangle -A DNS_MARK -j MARK --set-xmark 0x40/0x40
%s -w 2 -t mangle -A DNS_MARK -j ACCEPT
`, bin, bin, bin)
}

func legacyDnsRuleBypass(bin string) string {
	return fmt.Sprintf(`%s -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j RETURN
%s -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j RETURN
`, bin, bin)
}

func legacyDnsRuleMark(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`%s -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j TP_MARK
%s -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j TP_MARK
`, bin, bin)
}

// legacyDnsEarlyBypass must precede connmark restoration. A connection that
// was marked while DNS interception was on must not make port 53 reach the
// generic TPROXY target after the setting is turned off.
func legacyDnsEarlyBypass(bin string, dnsHijack bool) string {
	if dnsHijack {
		return ""
	}
	return legacyDnsRuleBypass(bin)
}

// legacyDnsLoopbackReturn lets a local query that the nat OUTPUT REDIRECT
// already rewrote to the module's port through instead of being TPROXY'd to
// 52345. It only exists while that REDIRECT is installed.
func legacyDnsLoopbackReturn(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`# 本机 DNS 查询在 nat OUTPUT 已被 REDIRECT 改写端口（53→52353），经 lo 重入时
# 带着 0x40 标记但端口不再是 53，必须放行，否则命中通用 TPROXY 规则被劫持到 52345
%s -w 2 -t mangle -A TP_PRE -p tcp --dport 52353 -j RETURN
%s -w 2 -t mangle -A TP_PRE -p udp --dport 52353 -j RETURN
`, bin, bin)
}

// legacyDnsTproxyToModule hands marked port-53 traffic to the DNS module,
// ahead of the generic TPROXY rules.
func legacyDnsTproxyToModule(bin string, dnsHijack bool, onIP string) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`# DNS 流量重定向到新 DNS 模块端口 52353（必须在通用 TPROXY 规则之前）
%s -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip %s
%s -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip %s
`, bin, onIP, bin, onIP)
}

// nftDnsEarlyBypass must precede conntrack restoration for the same reason as
// legacyDnsEarlyBypass: a stale mark must not send port 53 to generic TPROXY.
func nftDnsEarlyBypass(dnsHijack bool) string {
	if dnsHijack {
		return ""
	}
	return `        meta l4proto { tcp, udp } th dport 53 return
`
}

// nftDnsMarkChain is the chain the output and prerouting hooks jump into to
// mark port 53 for the TPROXY rules.
func nftDnsMarkChain(dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return `    chain dns_mark {
        meta mark & 0x80 == 0x80 return
        meta mark set mark | 0x40
        return
    }

`
}

// nftDnsMarkHooks are the statements that jump into dns_mark ahead of the
// generic transparent proxy rules.
func nftDnsMarkHooks(dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return `        # DNS 规则必须在透明代理规则之前匹配
        meta nfproto { ipv4, ipv6 } meta l4proto { tcp, udp } th dport 53 jump dns_mark
`
}

// nftDnsTproxyToModule hands marked port-53 traffic to the DNS module, ahead
// of the generic TPROXY rules.
func nftDnsTproxyToModule(dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return `        # 本机 DNS 查询在 nat OUTPUT 已被 REDIRECT 改写端口（53→52353），经 lo 重入时
        # 带着 0x40 标记但端口不再是 53，必须放行，否则命中通用 TPROXY 规则被劫持到 52345
        meta l4proto { tcp, udp } th dport 52353 return
        # DNS 流量重定向到新 DNS 模块端口 52353（必须在通用 TPROXY 规则之前）
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 th dport 53 tproxy ip to 127.2.0.17:52353
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 th dport 53 tproxy ip6 to [::1]:52353
`
}
