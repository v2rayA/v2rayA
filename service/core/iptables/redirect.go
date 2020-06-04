package iptables

import (
	"v2rayA/common/cmds"
)

type redirect struct{ iptablesSetter }

var Redirect redirect

func (r *redirect) GetSetupCommands() SetupCommands {
	commands := `
iptables -t nat -N V2RAY
# 出方向白名单端口
iptables -t nat -A V2RAY -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
iptables -t nat -A V2RAY -i docker+ -j RETURN
iptables -t nat -A V2RAY -i veth+ -j RETURN
iptables -t nat -A V2RAY -i br-+ -j RETURN
iptables -t nat -A V2RAY -d 10.0.0.0/8 -j RETURN
iptables -t nat -A V2RAY -d 100.64.0.0/10 -j RETURN
iptables -t nat -A V2RAY -d 127.0.0.0/8 -j RETURN
iptables -t nat -A V2RAY -d 169.254.0.0/16 -j RETURN
iptables -t nat -A V2RAY -d 172.16.0.0/12 -j RETURN
iptables -t nat -A V2RAY -d 192.0.0.0/24 -j RETURN
iptables -t nat -A V2RAY -d 192.0.2.0/24 -j RETURN
iptables -t nat -A V2RAY -d 192.88.99.0/24 -j RETURN
iptables -t nat -A V2RAY -d 192.168.0.0/16 -j RETURN
iptables -t nat -A V2RAY -d 198.18.0.0/15 -j RETURN
iptables -t nat -A V2RAY -d 198.51.100.0/24 -j RETURN
iptables -t nat -A V2RAY -d 203.0.113.0/24 -j RETURN
iptables -t nat -A V2RAY -d 224.0.0.0/4 -j RETURN
iptables -t nat -A V2RAY -d 240.0.0.0/4 -j RETURN
iptables -t nat -A V2RAY -d 255.255.255.255/32 -j RETURN
iptables -t nat -A V2RAY -m mark --mark 0xff -j RETURN
iptables -t nat -A V2RAY -p tcp -j REDIRECT --to-ports 32345

iptables -t nat -A PREROUTING -p tcp -j V2RAY
iptables -t nat -A OUTPUT -p tcp -j V2RAY
`
	if cmds.IsCommandValid("sysctl") {
		commands += `
#禁用ipv6
sysctl -w net.ipv6.conf.all.disable_ipv6=1
sysctl -w net.ipv6.conf.default.disable_ipv6=1
	`
	}
	return SetupCommands(commands)
}

func (r *redirect) GetCleanCommands() CleanCommands {
	commands := `
iptables -t nat -F V2RAY
iptables -t nat -D PREROUTING -p tcp -j V2RAY
iptables -t nat -D OUTPUT -p tcp -j V2RAY
iptables -t nat -X V2RAY
`
	if cmds.IsCommandValid("sysctl") {
		commands += `
sysctl -w net.ipv6.conf.all.disable_ipv6=0
sysctl -w net.ipv6.conf.default.disable_ipv6=0
`
	}
	return CleanCommands(commands)
}
