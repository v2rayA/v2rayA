package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"os"
	"strconv"
	"strings"
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
	if configure.GetSettingNotNil().AntiPollution != configure.AntipollutionClosed {
		pos += 3
	}
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

iptables -w 2 -t mangle -I OUTPUT -j TP_OUT
iptables -w 2 -t mangle -I PREROUTING -j TP_PRE

iptables -w 2 -t mangle -A TP_OUT -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t mangle -A TP_OUT -p tcp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_OUT -p udp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE

iptables -w 2 -t mangle -A TP_PRE -i lo -m mark ! --mark 0x40/0xc0 -j RETURN
iptables -w 2 -t mangle -A TP_PRE -p tcp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_PRE -p udp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
iptables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip 127.0.0.1
iptables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip 127.0.0.1

iptables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark
iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -i docker+ -j RETURN
`
	if notSkip, _ := strconv.ParseBool(TproxyNotSkipBr); !notSkip {
		commands += `iptables -w 2 -t mangle -A TP_RULE -i br+ -j RETURN`
	}
	commands += `
iptables -w 2 -t mangle -A TP_RULE -i veth+ -j RETURN
iptables -w 2 -t mangle -A TP_RULE -i ppp+ -j RETURN
`
	if configure.GetSettingNotNil().AntiPollution != configure.AntipollutionClosed {
		commands += `
iptables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j TP_MARK
iptables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j TP_MARK
iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
	}
	commands += `
iptables -w 2 -t mangle -A TP_RULE -d 0.0.0.0/32 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 10.0.0.0/8 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 100.64.0.0/10 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 127.0.0.0/8 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 169.254.0.0/16 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 172.16.0.0/12 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 192.0.0.0/24 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 192.0.2.0/24 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 192.88.99.0/24 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 192.168.0.0/16 -j RETURN
# fakedns
# iptables -w 2 -t mangle -A TP_RULE -d 198.18.0.0/15 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 198.51.100.0/24 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 203.0.113.0/24 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 224.0.0.0/4 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -d 240.0.0.0/4 -j RETURN
iptables -w 2 -t mangle -A TP_RULE -j TP_MARK

iptables -w 2 -t mangle -A TP_MARK -p tcp -m tcp --syn -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A TP_MARK -p udp -m conntrack --ctstate NEW -j MARK --set-xmark 0x40/0x40
iptables -w 2 -t mangle -A TP_MARK -j CONNMARK --save-mark
`
	if IsIPv6Supported() {
		commands += `
ip -6 rule add fwmark 0x40/0xc0 table 100
ip -6 route add local ::/0 dev lo table 100

ip6tables -w 2 -t mangle -N TP_OUT
ip6tables -w 2 -t mangle -N TP_PRE
ip6tables -w 2 -t mangle -N TP_RULE
ip6tables -w 2 -t mangle -N TP_MARK

ip6tables -w 2 -t mangle -I OUTPUT -j TP_OUT
ip6tables -w 2 -t mangle -I PREROUTING -j TP_PRE

ip6tables -w 2 -t mangle -A TP_OUT -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t mangle -A TP_OUT -p tcp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_OUT -p udp -m addrtype --src-type LOCAL ! --dst-type LOCAL -j TP_RULE

ip6tables -w 2 -t mangle -A TP_PRE -i lo -m mark ! --mark 0x40/0xc0 -j RETURN
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m addrtype ! --src-type LOCAL ! --dst-type LOCAL -j TP_RULE
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip ::1
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345 --on-ip ::1

ip6tables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark
ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
		if notSkip, _ := strconv.ParseBool(TproxyNotSkipBr); !notSkip {
			commands += `ip6tables -w 2 -t mangle -A TP_RULE -i br+ -j RETURN`
		}
		commands += `
ip6tables -w 2 -t mangle -A TP_RULE -i docker+ -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -i veth+ -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -i ppp+ -j RETURN
`
		if configure.GetSettingNotNil().AntiPollution != configure.AntipollutionClosed {
			commands += `
ip6tables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j TP_MARK
ip6tables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j TP_MARK
ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN
`
		}
		commands += `
ip6tables -w 2 -t mangle -A TP_RULE -d ::/128 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d ::1/128 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d 64:ff9b::/96 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d 100::/64 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d 2001::/32 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d 2001:20::/28 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d fe80::/10 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -d ff00::/8 -j RETURN
ip6tables -w 2 -t mangle -A TP_RULE -j TP_MARK

ip6tables -w 2 -t mangle -A TP_MARK -p tcp -m tcp --syn -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A TP_MARK -p udp -m conntrack --ctstate NEW -j MARK --set-xmark 0x40/0x40
ip6tables -w 2 -t mangle -A TP_MARK -j CONNMARK --save-mark
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
`
	}
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
	// 198.18.0.0/15 and fc00::/7 are reserved for private use but used by fakedns
	table := `
table inet v2raya {
    set whitelist {
        type ipv4_addr
        flags interval
        auto-merge
        elements = {
            0.0.0.0/32,
            10.0.0.0/8,
            100.64.0.0/10,
            127.0.0.0/8,
            169.254.0.0/16,
            172.16.0.0/12,
            192.0.0.0/24,
            192.0.2.0/24,
            192.88.99.0/24,
            192.168.0.0/16,
            198.51.100.0/24,
            203.0.113.0/24,
            224.0.0.0/4,
            240.0.0.0/4
        }
    }

    set whitelist6 {
        type ipv6_addr
        flags interval
        auto-merge
        elements = {
            ::/128,
            ::1/128,
            64:ff9b::/96,
            100::/64,
            2001::/32,
            2001:20::/28,
            fe80::/10,
            ff00::/8
        }
    }

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
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 tproxy ip to 127.0.0.1:52345
        meta l4proto { tcp, udp } mark & 0xc0 == 0x40 tproxy ip6 to [::1]:52345
    }

    chain output {
        type route hook output priority mangle - 5; policy accept;
        meta nfproto { ipv4, ipv6 } jump tp_out
    }

    chain prerouting {
        type filter hook prerouting priority mangle - 5; policy accept;
        meta nfproto { ipv4, ipv6 } jump tp_pre
    }

    chain tp_rule {
        meta mark set ct mark
        meta mark & 0xc0 == 0x40 return
`
	if notSkip, _ := strconv.ParseBool(TproxyNotSkipBr); !notSkip {
		table += `        iifname "br-*" return`
	}
	table += `
        iifname "docker*" return
        iifname "veth*" return
        iifname "wg*" return
        iifname "ppp*" return
        # anti-pollution
        ip daddr @interface return
        ip daddr @whitelist return
        ip6 daddr @interface6 return
        ip6 daddr @whitelist6 return
        jump tp_mark
    }

    chain tp_mark {
        tcp flags & (fin | syn | rst | ack) == syn meta mark set mark | 0x40
        meta l4proto udp ct state new meta mark set mark | 0x40
        ct mark set mark
    }
}
`
	if configure.GetSettingNotNil().AntiPollution != configure.AntipollutionClosed {
		table = strings.ReplaceAll(table, "# anti-pollution", `
        meta l4proto { tcp, udp } th dport 53 jump tp_mark
        meta mark & 0xc0 == 0x40 return
		`)
	}

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
	return Setter{Cmds: command}
}
