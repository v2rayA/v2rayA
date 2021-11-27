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
	commands = fmt.Sprintf(`iptables -w 2 -t nat -I V2RAY -d %s -j RETURN`, cidr)
	if !strings.Contains(cidr, ".") {
		//ipv6
		commands = strings.Replace(commands, "iptables", "ip6tables", 1)
	}
	cmds.ExecCommands(commands, false)
}

func (r *redirect) RemoveIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -D V2RAY -d %s -j RETURN`, cidr)
	cmds.ExecCommands(commands, false)
}

func (r *redirect) GetSetupCommands() SetupCommands {
	commands := `
iptables -w 2 -t nat -N V2RAY
iptables -w 2 -t nat -A V2RAY -d 0.0.0.0/32 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 10.0.0.0/8 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 100.64.0.0/10 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 127.0.0.0/8 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 169.254.0.0/16 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 172.16.0.0/12 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 192.0.0.0/24 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 192.0.2.0/24 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 192.88.99.0/24 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 192.168.0.0/16 -j RETURN
# fakedns
# iptables -w 2 -t nat -A V2RAY -d 198.18.0.0/15 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 198.51.100.0/24 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 203.0.113.0/24 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 224.0.0.0/4 -j RETURN
iptables -w 2 -t nat -A V2RAY -d 240.0.0.0/4 -j RETURN
iptables -w 2 -t nat -A V2RAY -m mark --mark 0x80/0x80 -j RETURN
iptables -w 2 -t nat -A V2RAY -p tcp -j REDIRECT --to-ports 32345

iptables -w 2 -t nat -I PREROUTING -p tcp -j V2RAY
iptables -w 2 -t nat -I OUTPUT -p tcp -j V2RAY
`
	if IsIPv6Supported() {
		commands += `
ip6tables -t nat -N V2RAY
ip6tables -w 2 -t nat -A V2RAY -d ::/128 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d ::1/128 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d 64:ff9b::/96 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d 100::/64 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d 2001::/32 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d 2001:20::/28 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d 2001:db8::/32 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d 2002::/16 -j RETURN
# fakedns
# ip6tables -w 2 -t nat -A V2RAY -d fc00::/7 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d fe80::/10 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -d ff00::/8 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -m mark --mark 0x80/0x80 -j RETURN
ip6tables -w 2 -t nat -A V2RAY -p tcp -j REDIRECT --to-ports 32345

ip6tables -w 2 -t nat -I PREROUTING -p tcp -j V2RAY
ip6tables -w 2 -t nat -I OUTPUT -p tcp -j V2RAY
`
	}
	return SetupCommands(commands)
}

func (r *redirect) GetCleanCommands() CleanCommands {
	commands := `
iptables -w 2 -t nat -F V2RAY
iptables -w 2 -t nat -D PREROUTING -p tcp -j V2RAY
iptables -w 2 -t nat -D OUTPUT -p tcp -j V2RAY
iptables -w 2 -t nat -X V2RAY
`
	if IsIPv6Supported() {
		commands += `
ip6tables -w 2 -t nat -F V2RAY
ip6tables -w 2 -t nat -D PREROUTING -p tcp -j V2RAY
ip6tables -w 2 -t nat -D OUTPUT -p tcp -j V2RAY
ip6tables -w 2 -t nat -X V2RAY
`
	}
	return CleanCommands(commands)
}
