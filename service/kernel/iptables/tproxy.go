package iptables

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
)

var (
	TproxyNotSkipBr string
)

type tproxy interface {
	AddIPWhitelist(cidr string)
	RemoveIPWhitelist(cidr string)
	GetSetupCommands() Setter
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
	if notSkip, _ := strconv.ParseBool(TproxyNotSkipBr); notSkip {
		pos--
	}

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

func (t *legacyTproxy) GetSetupCommands() Setter {
	commands := `
ip rule add fwmark 0x40/0xc0 table 100
ip route add local 0.0.0.0/0 dev lo table 100

iptables -w 2 -t mangle -N TP_OUT
iptables -w 2 -t mangle -N TP_PRE
iptables -w 2 -t mangle -N TP_RULE
iptables -w 2 -t mangle -N TP_MARK
iptables -w 2 -t mangle -N DNS_MARK

# DNS 规则必须在透明代理规则之前插入（环路保护）
iptables -w 2 -t mangle -I OUTPUT -p udp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -I OUTPUT -p tcp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -I OUTPUT -j TP_OUT
# DNS 规则必须在透明代理规则之前插入
iptables -w 2 -t mangle -I PREROUTING -p udp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -I PREROUTING -p tcp --dport 53 -j DNS_MARK
iptables -w 2 -t mangle -I PREROUTING -j TP_PRE

iptables -w 2 -t mangle -A TP_OUT -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t mangle -A TP_OUT -p tcp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_OUT -p udp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE

iptables -w 2 -t mangle -A TP_PRE -i lo -m mark ! --mark 0x40/0xc0 -j RETURN
iptables -w 2 -t mangle -A TP_PRE -p tcp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_PRE -p udp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
# DNS 流量重定向到新 DNS 模块端口 52353（必须在通用 TPROXY 规则之前）
iptables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip 127.2.0.17
iptables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip 127.2.0.17
# 通用 TPROXY 规则
iptables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip 127.0.0.1
iptables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip 127.0.0.1
iptables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark
iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
	for _, v := range GetExcludedInterfaces() {
		commands += fmt.Sprintf("iptables -w 2 -t mangle -A TP_RULE -i %s -j RETURN\n", strings.ReplaceAll(v, "*", "+"))
	}
	commands += `
iptables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j TP_MARK
iptables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j TP_MARK
iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`

	if IsEnabledTproxyWhiteIpGroups() {
		whiteIpv4List, _ := GetWhiteListIPs()
		for _, v := range whiteIpv4List {
			commands += fmt.Sprintf("iptables -w 2 -t mangle -A TP_RULE -d %s -j RETURN\n", v)
		}
	}
	commands += `
iptables -w 2 -t mangle -A TP_RULE -j TP_MARK

iptables -w 2 -t mangle -A TP_MARK -p tcp -m tcp --syn -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A TP_MARK -p udp -m conntrack --ctstate NEW -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A TP_MARK -j CONNMARK --save-mark
# DNS_MARK 链：环路保护 + 标记 DNS 流量
iptables -w 2 -t mangle -A DNS_MARK -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t mangle -A DNS_MARK -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A DNS_MARK -j ACCEPT
`
	if IsIPv6Supported() {
		commands += `
ip -6 rule add fwmark 0x40/0xc0 table 100
ip -6 route add local ::/0 dev lo table 100

ip6tables -w 2 -t mangle -N TP_OUT
ip6tables -w 2 -t mangle -N TP_PRE
ip6tables -w 2 -t mangle -N TP_RULE
ip6tables -w 2 -t mangle -N TP_MARK
ip6tables -w 2 -t mangle -N DNS_MARK

# IPv6 DNS 规则必须在透明代理规则之前插入
ip6tables -w 2 -t mangle -I OUTPUT -p udp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -I OUTPUT -p tcp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -I OUTPUT -j TP_OUT
# IPv6 DNS 规则必须在透明代理规则之前插入
ip6tables -w 2 -t mangle -I PREROUTING -p udp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -I PREROUTING -p tcp --dport 53 -j DNS_MARK
ip6tables -w 2 -t mangle -I PREROUTING -j TP_PRE

ip6tables -w 2 -t mangle -A TP_OUT -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t mangle -A TP_OUT -p tcp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_OUT -p udp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE

ip6tables -w 2 -t mangle -A TP_PRE -i lo -m mark ! --mark 0x40/0xc0 -j RETURN
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
# DNS 流量重定向到新 DNS 模块端口 52353（必须在通用 TPROXY 规则之前）
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip ::1
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip ::1
# 通用 TPROXY 规则
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip ::1
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip ::1
ip6tables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark
ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
		for _, v := range GetExcludedInterfaces() {
			commands += fmt.Sprintf("ip6tables -w 2 -t mangle -A TP_RULE -i %s -j RETURN\n", strings.ReplaceAll(v, "*", "+"))
		}
		commands += `
ip6tables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j TP_MARK
ip6tables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j TP_MARK
ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
		if IsEnabledTproxyWhiteIpGroups() {
			_, whiteIpv6List := GetWhiteListIPs()
			for _, v := range whiteIpv6List {
				commands += fmt.Sprintf("ip6tables -w 2 -t mangle -A TP_RULE -d %s -j RETURN\n", v)
			}
		}
		commands += `
ip6tables -w 2 -t mangle -A TP_RULE -j TP_MARK

ip6tables -w 2 -t mangle -A TP_MARK -p tcp -m tcp --syn -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A TP_MARK -p udp -m conntrack --ctstate NEW -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A TP_MARK -j CONNMARK --save-mark
# IPv6 DNS_MARK 链：环路保护 + 标记 DNS 流量
ip6tables -w 2 -t mangle -A DNS_MARK -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t mangle -A DNS_MARK -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A DNS_MARK -j ACCEPT
`
	}
	return Setter{
		Cmds: commands,
	}
}

func (t *legacyTproxy) GetCleanCommands() Setter {
	commands := `
ip rule del fwmark 0x40/0xc0 table 100
ip route del local 0.0.0.0/0 dev lo table 100

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
ip -6 rule del fwmark 0x40/0xc0 table 100
ip -6 route del local ::/0 dev lo table 100

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
	commands += "conntrack -D --mark 0x40 2>/dev/null || true\n"
	return Setter{
		Cmds: commands,
	}
}

func (t *nftTproxy) AddIPWhitelist(cidr string) {
	command := fmt.Sprintf("nft add element inet v2raya interface { %s }", cidr)
	if !strings.Contains(cidr, ".") {
		command = strings.Replace(command, "interface", "interface6", 1)
	}
	cmds.ExecCommands(command, false)
}

func (t *nftTproxy) RemoveIPWhitelist(cidr string) {
	command := fmt.Sprintf("nft delete element inet v2raya interface { %s }", cidr)
	if !strings.Contains(cidr, ".") {
		command = strings.Replace(command, "interface", "interface6", 1)
	}
	cmds.ExecCommands(command, false)
}

func (t *nftTproxy) GetSetupCommands() Setter {
	table := `
	table inet v2raya {
`
	if IsEnabledTproxyWhiteIpGroups() {
		whiteIpv4List, whiteIpv6List := GetWhiteListIPs()
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
    set interface {
        type ipv4_addr
        flags interval
        auto-merge
    }

    set interface6 {
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
        # DNS 流量重定向到新 DNS 模块端口 52353（必须在通用 TPROXY 规则之前）
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 th dport 53 tproxy ip to 127.2.0.17:52353
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 th dport 53 tproxy ip6 to [::1]:52353
        # 通用 TPROXY 规则
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 tproxy ip to 127.0.0.1:52345
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 tproxy ip6 to [::1]:52345
    }

    chain dns_mark {
        meta mark & 0x80 == 0x80 return
        meta mark set mark | 0x40
        accept
    }

    chain output {
        type route hook output priority mangle - 5; policy accept;
        # DNS 规则必须在透明代理规则之前匹配
        meta nfproto { ipv4, ipv6 } meta l4proto { tcp, udp } th dport 53 jump dns_mark
        meta nfproto { ipv4, ipv6 } jump tp_out
    }

    chain prerouting {
        type filter hook prerouting priority mangle - 5; policy accept;
        # DNS 规则必须在透明代理规则之前匹配
        meta nfproto { ipv4, ipv6 } meta l4proto { tcp, udp } th dport 53 jump dns_mark
        meta nfproto { ipv4, ipv6 } jump tp_pre
    }

    chain tp_rule {
        meta mark set ct mark
        meta mark & 0xc0 == 0x40 return
`
	for _, v := range GetExcludedInterfaces() {
		table += fmt.Sprintf("        iifname \"%s\" return\n", v)
	}
	table += `
        # anti-pollution
        ip daddr @interface return
	`
	if IsEnabledTproxyWhiteIpGroups() {
		table += `
        ip daddr @whitelist return
        ip6 daddr @whitelist6 return
	`
	}
	table += `
        ip6 daddr @interface6 return
        jump tp_mark
    }

    chain tp_mark {
        tcp flags & (fin | syn | rst | ack) == syn meta mark set mark | 0x40
        meta l4proto udp ct state new meta mark set mark | 0x40
        ct mark set mark
    }
}
`
	table = strings.ReplaceAll(table, "# anti-pollution", `
        meta l4proto { tcp, udp } th dport 53 jump tp_mark
        meta mark & 0xc0 == 0x40 return
		`)

	if !IsIPv6Supported() {
		// drop ipv6 packets hooks
		table = strings.ReplaceAll(table, "meta nfproto { ipv4, ipv6 }", "meta nfproto ipv4")
	}

	nftablesConf := asset.GetNftablesConfigPath()
	os.WriteFile(nftablesConf, []byte(table), 0644)

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
ip rule del fwmark 0x40/0xc0 table 100
ip route del local 0.0.0.0/0 dev lo table 100
`
	if IsIPv6Supported() {
		command += `
ip -6 rule del fwmark 0x40/0xc0 table 100
ip -6 route del local ::/0 dev lo table 100
		`
	}

	command += `nft delete table inet v2raya`
	if !IsIPv6Supported() {
		command = strings.Replace(command, "inet", "ip", 1)
	}
	command += "\nconntrack -D --mark 0x40 2>/dev/null || true"
	return Setter{Cmds: command}
}
