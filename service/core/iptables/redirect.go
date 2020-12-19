package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/cmds"
)

type redirect struct{}

var Redirect redirect

func (r *redirect) AddIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -t nat -I V2RAY -d %s -j RETURN`, cidr)
	cmds.ExecCommands(commands, false)
}

func (r *redirect) RemoveIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -t mangle -D V2RAY -d %s -j RETURN`, cidr)
	cmds.ExecCommands(commands, false)
}

func (r *redirect) GetSetupCommands() SetupCommands {
	commands := `
iptables -t nat -N V2RAY
# 出方向白名单端口
iptables -t nat -A V2RAY -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
# 白名单自动插入
iptables -t nat -A V2RAY -m mark --mark 0xff -j RETURN
iptables -t nat -A V2RAY -p tcp -j REDIRECT --to-ports 32345

iptables -t nat -I PREROUTING -p tcp -j V2RAY
iptables -t nat -I OUTPUT -p tcp -j V2RAY
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
