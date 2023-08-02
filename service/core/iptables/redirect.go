package iptables

import (
	"fmt"
	"os"
	"strings"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
)

type redirect interface {
	AddIPWhitelist(cidr string)
	RemoveIPWhitelist(cidr string)
	GetSetupCommands() Setter
	GetCleanCommands() Setter
}

type legacyRedirect struct{}
type nftRedirect struct{}

var Redirect redirect

func init() {
	if IsNftablesSupported() {
		Redirect = &nftRedirect{}
	} else {
		Redirect = &legacyRedirect{}
	}
}

func (r *legacyRedirect) AddIPWhitelist(cidr string) {
	// avoid duplication
	r.RemoveIPWhitelist(cidr)
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t nat -I TP_RULE -d %s -j RETURN`, cidr)
	if !strings.Contains(cidr, ".") {
		//ipv6
		commands = strings.Replace(commands, "iptables", "ip6tables", 1)
	}
	cmds.ExecCommands(commands, false)
}

func (r *legacyRedirect) RemoveIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -D TP_RULE -d %s -j RETURN`, cidr)
	cmds.ExecCommands(commands, false)
}

func (r *legacyRedirect) GetSetupCommands() Setter {
	commands := `
iptables -w 2 -t nat -N TP_OUT
iptables -w 2 -t nat -N TP_PRE
iptables -w 2 -t nat -N TP_RULE
iptables -w 2 -t nat -A TP_RULE -d 0.0.0.0/32 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 10.0.0.0/8 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 100.64.0.0/10 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 127.0.0.0/8 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 169.254.0.0/16 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 172.16.0.0/12 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 192.0.0.0/24 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 192.0.2.0/24 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 192.88.99.0/24 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 192.168.0.0/16 -j RETURN
# fakedns
# iptables -w 2 -t nat -A TP_RULE -d 198.18.0.0/15 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 198.51.100.0/24 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 203.0.113.0/24 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 224.0.0.0/4 -j RETURN
iptables -w 2 -t nat -A TP_RULE -d 240.0.0.0/4 -j RETURN
iptables -w 2 -t nat -A TP_RULE -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 52345

iptables -w 2 -t nat -I PREROUTING -p tcp -j TP_PRE
iptables -w 2 -t nat -I OUTPUT -p tcp -j TP_OUT
iptables -w 2 -t nat -A TP_PRE -j TP_RULE
iptables -w 2 -t nat -A TP_OUT -j TP_RULE
`
	if IsIPv6Supported() {
		commands += `
ip6tables -w 2 -t nat -N TP_OUT
ip6tables -w 2 -t nat -N TP_PRE
ip6tables -w 2 -t nat -N TP_RULE
ip6tables -w 2 -t nat -A TP_RULE -d ::/128 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d ::1/128 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d 64:ff9b::/96 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d 100::/64 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d 2001::/32 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d 2001:20::/28 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d 2001:db8::/32 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d 2002::/16 -j RETURN
# fakedns
# ip6tables -w 2 -t nat -A TP_RULE -d fc00::/7 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d fe80::/10 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -d ff00::/8 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 52345

ip6tables -w 2 -t nat -I PREROUTING -p tcp -j TP_PRE
ip6tables -w 2 -t nat -I OUTPUT -p tcp -j TP_OUT
ip6tables -w 2 -t nat -A TP_PRE -j TP_RULE
ip6tables -w 2 -t nat -A TP_OUT -j TP_RULE
`
	}
	return Setter{
		Cmds: commands,
	}
}

func (r *legacyRedirect) GetCleanCommands() Setter {
	commands := `
iptables -w 2 -t nat -F TP_OUT
iptables -w 2 -t nat -D OUTPUT -p tcp -j TP_OUT
iptables -w 2 -t nat -X TP_OUT
iptables -w 2 -t nat -F TP_PRE
iptables -w 2 -t nat -D PREROUTING -p tcp -j TP_PRE
iptables -w 2 -t nat -X TP_PRE
iptables -w 2 -t nat -F TP_RULE
iptables -w 2 -t nat -X TP_RULE
`
	if IsIPv6Supported() {
		commands += `
ip6tables -w 2 -t nat -F TP_OUT
ip6tables -w 2 -t nat -D OUTPUT -p tcp -j TP_OUT
ip6tables -w 2 -t nat -X TP_OUT
ip6tables -w 2 -t nat -F TP_PRE
ip6tables -w 2 -t nat -D PREROUTING -p tcp -j TP_PRE
ip6tables -w 2 -t nat -X TP_PRE
ip6tables -w 2 -t nat -F TP_RULE
ip6tables -w 2 -t nat -X TP_RULE
`
	}
	return Setter{
		Cmds: commands,
	}
}

func (t *nftRedirect) AddIPWhitelist(cidr string) {
	command := fmt.Sprintf("nft add element inet v2raya interface { %s }", cidr)
	if !strings.Contains(cidr, ".") {
		command = strings.Replace(command, "interface", "interface6", 1)
	}
	cmds.ExecCommands(command, false)
}

func (t *nftRedirect) RemoveIPWhitelist(cidr string) {
	command := fmt.Sprintf("nft delete element inet v2raya interface { %s }", cidr)
	if !strings.Contains(cidr, ".") {
		command = strings.Replace(command, "interface", "interface6", 1)
	}
	cmds.ExecCommands(command, false)
}

func (r *nftRedirect) GetSetupCommands() Setter {
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

    chain tp_rule {
        ip daddr @whitelist return
        ip daddr @interface return
        ip6 daddr @whitelist6 return
        ip6 daddr @interface6 return
        meta mark & 0x80 == 0x80 return
        meta l4proto tcp redirect to :52345
    }

    chain tp_pre {
        type nat hook prerouting priority dstnat - 5
        meta nfproto { ipv4, ipv6 } meta l4proto tcp jump tp_rule
    }

    chain tp_out {
        type nat hook output priority -105
        meta nfproto { ipv4, ipv6 } meta l4proto tcp jump tp_rule
    }
}
`
	if !IsIPv6Supported() {
		table = strings.ReplaceAll(table, "meta nfproto { ipv4, ipv6 }", "meta nfproto ipv4")
	}

	nftablesConf := asset.GetNftablesConfigPath()
	os.WriteFile(nftablesConf, []byte(table), 0644)

	command := `nft -f ` + nftablesConf

	return Setter{Cmds: command}
}

func (r *nftRedirect) GetCleanCommands() Setter {
	command := `nft delete table inet v2raya`
	return Setter{Cmds: command}
}
