package iptables

import (
	"fmt"
	"os"
	"strings"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
)

type tproxy interface {
	AddIPWhitelist(cidr string)
	RemoveIPWhitelist(cidr string)
	GetSetupCommands(dnsHijack bool) Setter
	GetCleanCommands() Setter
}

type legacyTproxy struct{}

type nftTproxy struct{}

var Tproxy tproxy

func init() {
	if IsNftablesSupported() {
		Tproxy = &nftTproxy{}
	} else {
		Tproxy = &legacyTproxy{}
	}
}

func (t *legacyTproxy) AddIPWhitelist(cidr string) {
	// avoid duplication
	t.RemoveIPWhitelist(cidr)
	pos := 7

	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -I TP_RULE %v -d %s -j RETURN`, pos, cidr)
	if !strings.Contains(cidr, ".") {
		//ipv6
		commands = strings.Replace(commands, "iptables", "ip6tables", 1)
	}
	cmds.ExecCommands(commands, false)
}

func (t *legacyTproxy) RemoveIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -D TP_RULE -d %s -j RETURN`, cidr)
	if !strings.Contains(cidr, ".") {
		//ipv6
		commands = strings.Replace(commands, "iptables", "ip6tables", 1)
	}
	cmds.ExecCommands(commands, false)
}

func (t *legacyTproxy) GetSetupCommands(dnsHijack bool) Setter {
	excludedInterfaces, whiteIpv4List, whiteIpv6List, err := legacySetupValues()
	if err != nil {
		return NewErrorSetter(err)
	}

	commands := `
ip rule add fwmark 0x40/0xc0 table 100
ip route add local 0.0.0.0/0 dev lo table 100

iptables -w 2 -t mangle -N TP_OUT
iptables -w 2 -t mangle -N TP_PRE
iptables -w 2 -t mangle -N TP_RULE
` + legacyDnsEarlyBypass("iptables", dnsHijack) + `
iptables -w 2 -t mangle -N TP_MARK
` + legacyDnsMarkChain("iptables", dnsHijack) + legacyDnsMarkHooks("iptables", dnsHijack) + `
iptables -w 2 -t mangle -I OUTPUT -j TP_OUT
iptables -w 2 -t mangle -I PREROUTING -j TP_PRE

iptables -w 2 -t mangle -A TP_OUT -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t mangle -A TP_OUT -p tcp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_OUT -p udp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE

iptables -w 2 -t mangle -A TP_PRE -i lo -m mark ! --mark 0x40/0xc0 -j RETURN
iptables -w 2 -t mangle -A TP_PRE -p tcp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_PRE -p udp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
` + legacyDnsLoopbackReturn("iptables", dnsHijack) + legacyDnsTproxyToModule("iptables", dnsHijack, "127.2.0.17") + `
# 通用 TPROXY 规则
iptables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip 127.0.0.1
iptables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip 127.0.0.1
iptables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark
iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
	for _, v := range excludedInterfaces {
		commands += fmt.Sprintf("iptables -w 2 -t mangle -A TP_RULE -i %s -j RETURN\n", strings.ReplaceAll(v, "*", "+"))
	}
	// OUTPUT 路径使用 -o 匹配输出网卡（本地流量在 OUTPUT 链中 -i 始终为 lo）
	for _, v := range excludedInterfaces {
		commands += fmt.Sprintf("iptables -w 2 -t mangle -A TP_RULE -o %s -j RETURN\n", strings.ReplaceAll(v, "*", "+"))
	}
	commands += legacyDnsRuleMark("iptables", dnsHijack) + `
iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`

	commands += legacyWhitelist4Marker + "\n"
	commands += `
iptables -w 2 -t mangle -A TP_RULE -j TP_MARK

iptables -w 2 -t mangle -A TP_MARK -p tcp -m tcp --syn -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A TP_MARK -p udp -m conntrack --ctstate NEW -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A TP_MARK -j CONNMARK --save-mark
` + legacyDnsMarkChainRules("iptables", dnsHijack)
	if legacyIPv6Supported() {
		commands += `
ip -6 rule add fwmark 0x40/0xc0 table 100
ip -6 route add local ::/0 dev lo table 100

ip6tables -w 2 -t mangle -N TP_OUT
ip6tables -w 2 -t mangle -N TP_PRE
ip6tables -w 2 -t mangle -N TP_RULE
` + legacyDnsEarlyBypass("ip6tables", dnsHijack) + `
ip6tables -w 2 -t mangle -N TP_MARK
` + legacyDnsMarkChain("ip6tables", dnsHijack) + legacyDnsMarkHooks("ip6tables", dnsHijack) + `
ip6tables -w 2 -t mangle -I OUTPUT -j TP_OUT
ip6tables -w 2 -t mangle -I PREROUTING -j TP_PRE

ip6tables -w 2 -t mangle -A TP_OUT -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t mangle -A TP_OUT -p tcp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_OUT -p udp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE

ip6tables -w 2 -t mangle -A TP_PRE -i lo -m mark ! --mark 0x40/0xc0 -j RETURN
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
` + legacyDnsLoopbackReturn("ip6tables", dnsHijack) + legacyDnsTproxyToModule("ip6tables", dnsHijack, "::1") + `
# 通用 TPROXY 规则
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip ::1
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip ::1
ip6tables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark
ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
		for _, v := range excludedInterfaces {
			commands += fmt.Sprintf("ip6tables -w 2 -t mangle -A TP_RULE -i %s -j RETURN\n", strings.ReplaceAll(v, "*", "+"))
		}
		// IPv6 OUTPUT 路径使用 -o 匹配输出网卡
		for _, v := range excludedInterfaces {
			commands += fmt.Sprintf("ip6tables -w 2 -t mangle -A TP_RULE -o %s -j RETURN\n", strings.ReplaceAll(v, "*", "+"))
		}
		commands += legacyDnsRuleMark("ip6tables", dnsHijack) + `
ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
		commands += legacyWhitelist6Marker + "\n"
		commands += `
ip6tables -w 2 -t mangle -A TP_RULE -j TP_MARK

ip6tables -w 2 -t mangle -A TP_MARK -p tcp -m tcp --syn -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A TP_MARK -p udp -m conntrack --ctstate NEW -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A TP_MARK -j CONNMARK --save-mark
` + legacyDnsMarkChainRules("ip6tables", dnsHijack)
	}
	return newLegacyWhitelistSetter(withDnsModulePort(commands), "mangle", whiteIpv4List, whiteIpv6List)
}

func (t *legacyTproxy) GetCleanCommands() Setter {
	commands := `
ip rule del fwmark 0x40/0xc0 table 100 2>/dev/null || true
ip route del local 0.0.0.0/0 dev lo table 100 2>/dev/null || true

iptables -w 2 -t mangle -F TP_OUT
iptables -w 2 -t mangle -D OUTPUT -j TP_OUT
iptables -w 2 -t mangle -X TP_OUT
iptables -w 2 -t mangle -F TP_PRE
iptables -w 2 -t mangle -D PREROUTING -j TP_PRE
iptables -w 2 -t mangle -X TP_PRE
iptables -w 2 -t mangle -F TP_RULE
iptables -w 2 -t mangle -X TP_RULE
iptables -w 2 -t mangle -F TP_MARK
iptables -w 2 -t mangle -X TP_MARK
iptables -w 2 -t mangle -F DNS_MARK
iptables -w 2 -t mangle -D PREROUTING -p udp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -D PREROUTING -p tcp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -D OUTPUT -p udp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -D OUTPUT -p tcp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -X DNS_MARK
`
	if IsIPv6Supported() {
		commands += `
ip -6 rule del fwmark 0x40/0xc0 table 100 2>/dev/null || true
ip -6 route del local ::/0 dev lo table 100 2>/dev/null || true

ip6tables -w 2 -t mangle -F TP_OUT
ip6tables -w 2 -t mangle -D OUTPUT -j TP_OUT
ip6tables -w 2 -t mangle -X TP_OUT
ip6tables -w 2 -t mangle -F TP_PRE
ip6tables -w 2 -t mangle -D PREROUTING -j TP_PRE
ip6tables -w 2 -t mangle -X TP_PRE
ip6tables -w 2 -t mangle -F TP_RULE
ip6tables -w 2 -t mangle -X TP_RULE
ip6tables -w 2 -t mangle -F TP_MARK
ip6tables -w 2 -t mangle -X TP_MARK
ip6tables -w 2 -t mangle -F DNS_MARK
ip6tables -w 2 -t mangle -D PREROUTING -p udp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -D PREROUTING -p tcp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -D OUTPUT -p udp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -D OUTPUT -p tcp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -X DNS_MARK
`
	}
	commands += `conntrack -D --mark 0x40 2>/dev/null || true
ipset destroy v2raya_white4 2>/dev/null || true
ipset destroy v2raya_white6 2>/dev/null || true
`
	return Setter{
		Cmds: withDnsModulePort(commands),
	}
}

func (t *nftTproxy) AddIPWhitelist(cidr string) {
	command := fmt.Sprintf("nft add element inet v2raya local_ips { %s }", cidr)
	if !strings.Contains(cidr, ".") {
		command = strings.Replace(command, "local_ips", "local_ips6", 1)
	}
	cmds.ExecCommands(command, false)
}

func (t *nftTproxy) RemoveIPWhitelist(cidr string) {
	command := fmt.Sprintf("nft delete element inet v2raya local_ips { %s }", cidr)
	if !strings.Contains(cidr, ".") {
		command = strings.Replace(command, "local_ips", "local_ips6", 1)
	}
	cmds.ExecCommands(command, false)
}

func (t *nftTproxy) GetSetupCommands(dnsHijack bool) Setter {
	excludedInterfaces, whiteIpv4List, whiteIpv6List, err := getTproxySetupValues()
	if err != nil {
		return NewErrorSetter(err)
	}

	table := `
	table inet v2raya {
`
	if len(whiteIpv4List) > 0 {
		table += `
    set whitelist {
        type ipv4_addr
        flags interval
        auto-merge
        elements = {
`
		table += strings.Join(whiteIpv4List, ",")
		table += `
        }
    }
`
	}
	if len(whiteIpv6List) > 0 {
		table += `
    set whitelist6 {
        type ipv6_addr
        flags interval
        auto-merge
        elements = {
`
		table += strings.Join(whiteIpv6List, ",")
		table += `
        }
    }
`
	}

	// 198.18.0.0/15 and fc00::/7 are reserved for private use but used by fakedns

	table += `
    # 用于记录本机网卡 IP 地址（由 watcher 动态维护），
    # 避免发往本机地址的流量被重定向造成环路
    set local_ips {
        type ipv4_addr
        flags interval
        auto-merge
    }

    set local_ips6 {
        type ipv6_addr
        flags interval
        auto-merge
    }

    chain tp_out {
        meta mark & 0x80 == 0x80 return
        meta l4proto { tcp, udp } fib saddr type local fib daddr type != local jump tp_rule
    }

    chain tp_pre {
        iifname "lo" mark & 0xc0 != 0x40 return
        meta l4proto { tcp, udp } fib saddr type != local fib daddr type != local jump tp_rule
` + nftDnsTproxyToModule(dnsHijack) + `
        # 通用 TPROXY 规则
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 tproxy ip to 127.0.0.1:52345
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 tproxy ip6 to [::1]:52345
    }

` + nftDnsMarkChain(dnsHijack) + `
    chain output {
        type route hook output priority mangle - 5; policy accept;
` + nftDnsMarkHooks(dnsHijack) + `
        meta nfproto { ipv4, ipv6 } jump tp_out
    }

    chain prerouting {
        type filter hook prerouting priority mangle - 5; policy accept;
` + nftDnsMarkHooks(dnsHijack) + `
        meta nfproto { ipv4, ipv6 } jump tp_pre
    }

    chain tp_rule {
` + nftDnsEarlyBypass(dnsHijack) + `        meta mark set ct mark
        meta mark & 0xc0 == 0x40 return
`
	for _, v := range excludedInterfaces {
		table += fmt.Sprintf("        iifname \"%s\" return\n", v)
	}
	for _, v := range excludedInterfaces {
		table += fmt.Sprintf("        oifname \"%s\" return\n", v)
	}
	table += `
	        # anti-pollution
	        ip daddr @local_ips return
		`
	if len(whiteIpv4List) > 0 {
		table += `
        ip daddr @whitelist return
`
	}
	if len(whiteIpv6List) > 0 {
		table += `
        ip6 daddr @whitelist6 return
`
	}
	table += `
        ip6 daddr @local_ips6 return
        jump tp_mark
    }

    chain tp_mark {
        tcp flags & (fin | syn | rst | ack) == syn meta mark set mark | 0x40
        meta l4proto udp ct state new meta mark set mark | 0x40
        ct mark set mark
    }
}
`
	// Port 53 is marked for the DNS module while interception is on. With the
	// opt-out it is already returned at the start of tp_rule, before a stale
	// conntrack mark can reach the generic TPROXY rule.
	dnsPortRule := "        meta l4proto { tcp, udp } th dport 53 jump tp_mark\n"
	if !dnsHijack {
		dnsPortRule = ""
	}
	table = strings.ReplaceAll(table, "# anti-pollution", `
`+dnsPortRule+`        meta mark & 0xc0 == 0x40 return
		`)

	if !IsIPv6Supported() {
		// drop ipv6 packets hooks
		table = strings.ReplaceAll(table, "meta nfproto { ipv4, ipv6 }", "meta nfproto ipv4")
	}

	nftablesConf := asset.GetNftablesConfigPath()
	os.WriteFile(nftablesConf, []byte(withDnsModulePort(table)), 0644)

	command := `
ip rule add fwmark 0x40/0xc0 table 100
ip route add local 0.0.0.0/0 dev lo table 100
`
	if IsIPv6Supported() {
		command += `
ip -6 rule add fwmark 0x40/0xc0 table 100
ip -6 route add local ::/0 dev lo table 100
`
	}

	command += `nft -f ` + nftablesConf
	return Setter{Cmds: command}
}

func (t *nftTproxy) GetCleanCommands() Setter {
	command := `
ip rule del fwmark 0x40/0xc0 table 100 2>/dev/null || true
ip route del local 0.0.0.0/0 dev lo table 100 2>/dev/null || true
`
	if IsIPv6Supported() {
		command += `
ip -6 rule del fwmark 0x40/0xc0 table 100 2>/dev/null || true
ip -6 route del local ::/0 dev lo table 100 2>/dev/null || true
		`
	}

	command += `nft delete table inet v2raya`
	command += "\nconntrack -D --mark 0x40 2>/dev/null || true"
	return Setter{Cmds: command}
}
