package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/cmds"
	"strings"
)

type redirect struct{}

var Redirect redirect

func (r *redirect) AddIPWhitelist(cidr string) {
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

func (r *redirect) RemoveIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -D TP_RULE -d %s -j RETURN`, cidr)
	cmds.ExecCommands(commands, false)
}

func (r *redirect) GetSetupCommands() Setter {
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
iptables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 32345

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
ip6tables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 32345

ip6tables -w 2 -t nat -I PREROUTING -p tcp -j TP_PRE
ip6tables -w 2 -t nat -I OUTPUT -p tcp -j TP_OUT
ip6tables -w 2 -t nat -A TP_PRE -j TP_RULE
ip6tables -w 2 -t nat -A TP_OUT -j TP_RULE
`
	}
	return Setter{
		Cmds:      commands,
	}
}

func (r *redirect) GetCleanCommands() Setter {
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
		Cmds:      commands,
	}
}
