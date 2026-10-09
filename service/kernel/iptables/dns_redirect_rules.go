package iptables

import "fmt"

// The DNS fragments below are what the opt-out removes. With it on, port 53
// is redirected to the DNS module by a chain jumped into ahead of the
// generic transparent rules; with it off the chain is never created and
// port 53 returns from TP_RULE instead of being redirected to 52345. Each
// takes the iptables binary name, so IPv4 and IPv6 stay one set of rules.

// legacyDnsRedirectChain creates the chain PREROUTING and OUTPUT jump into.
func legacyDnsRedirectChain(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return bin + " -w 2 -t nat -N DNS_REDIRECT\n"
}

// legacyDnsRedirectChainRules fills it: the loop guard, then the redirect to
// the module's port for both transports.
func legacyDnsRedirectChainRules(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`# DNS 重定向到新 DNS 模块端口 52353（必须在通用 REDIRECT 规则之前）
%s -w 2 -t nat -A DNS_REDIRECT -m mark --mark 0x80/0x80 -j RETURN
%s -w 2 -t nat -A DNS_REDIRECT -p tcp -j REDIRECT --to-port 52353
%s -w 2 -t nat -A DNS_REDIRECT -p udp -j REDIRECT --to-port 52353
`, bin, bin, bin)
}

// legacyDnsBypass returns port 53 from TP_RULE. The generic rule there
// redirects every TCP packet, so without this the queries would end up at
// the transparent proxy's port rather than the resolver the user chose.
func legacyDnsBypass(bin string, dnsHijack bool) string {
	if dnsHijack {
		return ""
	}
	return fmt.Sprintf(`%s -w 2 -t nat -A TP_RULE -p udp --dport 53 -j RETURN
%s -w 2 -t nat -A TP_RULE -p tcp --dport 53 -j RETURN
`, bin, bin)
}

// legacyDnsRedirectHooks are the -I rules that send port 53 into the chain.
// They come after the TP_PRE/TP_OUT insertions: -I always puts a rule at the
// head of the chain, so the later insertion ends up in front — otherwise a
// local or LAN TCP query would hit TP_OUT/TP_PRE first, be redirected to
// 52345 and miss the DNS module.
func legacyDnsRedirectHooks(bin string, dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return fmt.Sprintf(`# DNS 跳转在 TP_PRE/TP_OUT 之后插入：-I 总是插到链首，后插入者排在前面，
# 因此 DNS 规则实际位于通用透明代理规则之前（否则本机/局域网 TCP DNS 会
# 先命中 TP_OUT/TP_PRE 被重定向到 52345，绕过 DNS 模块）。
%s -w 2 -t nat -I PREROUTING -p udp --dport 53 -j DNS_REDIRECT
%s -w 2 -t nat -I PREROUTING -p tcp --dport 53 -j DNS_REDIRECT
%s -w 2 -t nat -I OUTPUT -p udp --dport 53 -j DNS_REDIRECT
%s -w 2 -t nat -I OUTPUT -p tcp --dport 53 -j DNS_REDIRECT
`, bin, bin, bin, bin)
}

// nftRedirectChain is the chain the prerouting and output hooks jump into to
// redirect port 53 to the DNS module's port.
func nftRedirectChain(dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return `    chain dns_redirect {
        meta mark & 0x80 == 0x80 return
        meta l4proto { tcp, udp } th dport 53 redirect to :52353
    }

`
}

// nftRedirectHook is the statement that jumps into the chain, ahead of the
// generic transparent rule.
func nftRedirectHook(dnsHijack bool) string {
	if !dnsHijack {
		return ""
	}
	return `        # DNS 重定向到 52353（优先于通用透明代理）
        meta nfproto { ipv4, ipv6 } meta l4proto { tcp, udp } th dport 53 jump dns_redirect
`
}

// nftRedirectBypass returns port 53 from tp_rule. The rule below it redirects
// every TCP packet to the transparent proxy's port, so without this the
// queries would end up there rather than at the resolver the user chose.
func nftRedirectBypass(dnsHijack bool) string {
	if dnsHijack {
		return ""
	}
	return `        meta l4proto { tcp, udp } th dport 53 return
`
}
