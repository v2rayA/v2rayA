package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"strings"
)

type tproxy struct {
	watcher *LocalIPWatcher
}

var Tproxy tproxy

func (t *tproxy) AddIPWhitelist(cidr string) {
	// avoid duplication
	t.RemoveIPWhitelist(cidr)
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -I SETMARK 6 -d %s -j RETURN`, cidr)
	if !strings.Contains(cidr, ".") {
		//ipv6
		commands = strings.Replace(commands, "iptables", "ip6tables", 1)
	}
	cmds.ExecCommands(commands, false)
}

func (t *tproxy) RemoveIPWhitelist(cidr string) {
	var commands string
	commands = fmt.Sprintf(`iptables -w 2 -t mangle -D SETMARK -d %s -j RETURN`, cidr)
	if !strings.Contains(cidr, ".") {
		//ipv6
		commands = strings.Replace(commands, "iptables", "ip6tables", 1)
	}
	cmds.ExecCommands(commands, false)
}

func (t *tproxy) GetSetupCommands() SetupCommands {
	commands := `
# 设置策略路由
ip rule add fwmark 1 table 100
ip route add local 0.0.0.0/0 dev lo table 100

# 建链
iptables -w 2 -t mangle -N TP_OUT
iptables -w 2 -t mangle -N TP_PRE
iptables -w 2 -t mangle -I OUTPUT -j TP_OUT
iptables -w 2 -t mangle -I PREROUTING -j TP_PRE

# 打上 iptables 标记，mark 了的会走代理
iptables -w 2 -t mangle -N SETMARK
# 出方向白名单端口
iptables -w 2 -t mangle -A SETMARK -i docker+ -j RETURN
iptables -w 2 -t mangle -A SETMARK -i veth+ -j RETURN
iptables -w 2 -t mangle -A SETMARK -i br-+ -j RETURN
iptables -w 2 -t mangle -A SETMARK -p udp --dport 53 -j MARK --set-mark 1
iptables -w 2 -t mangle -A SETMARK -p tcp --dport 53 -j MARK --set-mark 1
# 注意，如果要调整位置，记得调整func AddIPWhitelist的插入位置
iptables -w 2 -t mangle -A SETMARK -d 0.0.0.0/32 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 10.0.0.0/8 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 100.64.0.0/10 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 127.0.0.0/8 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 169.254.0.0/16 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 172.16.0.0/12 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 192.0.0.0/24 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 192.0.2.0/24 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 192.88.99.0/24 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 192.168.0.0/16 -j RETURN
# fakedns
# iptables -w 2 -t mangle -A SETMARK -d 198.18.0.0/15 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 198.51.100.0/24 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 203.0.113.0/24 -j RETURN
iptables -w 2 -t mangle -A SETMARK -d 224.0.0.0/4 -j RETURN
# supervisor
# iptables -w 2 -t mangle -A SETMARK -d 240.0.0.0/4 -j RETURN
iptables -w 2 -t mangle -A SETMARK -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
iptables -w 2 -t mangle -A SETMARK -p udp -m multiport --sports {{UDP_PORTS}} -j RETURN
iptables -w 2 -t mangle -A SETMARK -p tcp -j MARK --set-mark 1
iptables -w 2 -t mangle -A SETMARK -p udp -j MARK --set-mark 1

# 走过TPROXY的通行
iptables -w 2 -t mangle -A TP_OUT -m mark --mark 0xff -j RETURN
`
	if specialMode.ShouldLocalDnsListen() {
		commands += ` 
iptables -w 2 -t mangle -A TP_OUT -p udp --dport 53 -j RETURN
`
	}
	commands += `
# 本机发出去的 TCP 和 UDP 走一下 SETMARK 链
iptables -w 2 -t mangle -A TP_OUT -p tcp -m mark ! --mark 1 -j SETMARK
iptables -w 2 -t mangle -A TP_OUT -p udp -m mark ! --mark 1 -j SETMARK

# 走过TPROXY的通行
iptables -w 2 -t mangle -A TP_PRE -m mark --mark 0xff -j RETURN
`
	if specialMode.ShouldLocalDnsListen() {
		commands += ` 
iptables -w 2 -t mangle -A TP_PRE -p udp --dport 53 -j RETURN
`
	}
	commands += `
# 让内网主机发出的 TCP 和 UDP 走一下 SETMARK 链
iptables -w 2 -t mangle -A TP_PRE -p tcp -m mark ! --mark 1 -j SETMARK
iptables -w 2 -t mangle -A TP_PRE -p udp -m mark ! --mark 1 -j SETMARK
# 将所有打了标记的 TCP 和 UDP 包透明地转发到代理的监听端口
iptables -w 2 -t mangle -A TP_PRE -m mark --mark 1 -p tcp -j TPROXY --on-port 32345 --tproxy-mark 1
iptables -w 2 -t mangle -A TP_PRE -m mark --mark 1 -p udp -j TPROXY --on-port 32345 --tproxy-mark 1

`
	if IsIPv6Supported() {
		commands += `
# 设置策略路由
ip -6 rule add fwmark 1 table 100
ip -6 route add local ::/0 dev lo table 100

# 建链
ip6tables -w 2 -t mangle -N TP_OUT
ip6tables -w 2 -t mangle -N TP_PRE
ip6tables -w 2 -t mangle -I OUTPUT -j TP_OUT
ip6tables -w 2 -t mangle -I PREROUTING -j TP_PRE

# 打上 iptables 标记，mark 了的会走代理
ip6tables -w 2 -t mangle -N SETMARK
# 出方向白名单端口
ip6tables -w 2 -t mangle -A SETMARK -i docker+ -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -i veth+ -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -i br-+ -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -p udp --dport 53 -j MARK --set-mark 1
ip6tables -w 2 -t mangle -A SETMARK -p tcp --dport 53 -j MARK --set-mark 1
# 注意，如果要调整位置，记得调整func AddIPWhitelist的插入位置
ip6tables -w 2 -t mangle -A SETMARK -d ::/128 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d ::1/128 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d ::ffff:0:0/96 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d ::ffff:0:0:0/96 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d 64:ff9b::/96 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d 100::/64 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d 2001::/32 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d 2001:20::/28 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d 2001:db8::/32 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d 2002::/16 -j RETURN
# fakedns
# ip6tables -w 2 -t mangle -A SETMARK -d fc00::/7 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d fe80::/10 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -d ff00::/8 -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -p udp -m multiport --sports {{UDP_PORTS}} -j RETURN
ip6tables -w 2 -t mangle -A SETMARK -p tcp -j MARK --set-mark 1
ip6tables -w 2 -t mangle -A SETMARK -p udp -j MARK --set-mark 1

# 走过TPROXY的通行
ip6tables -w 2 -t mangle -A TP_OUT -m mark --mark 0xff -j RETURN
`
		if specialMode.ShouldLocalDnsListen() {
			commands += ` 
ip6tables -w 2 -t mangle -A TP_OUT -p udp --dport 53 -j RETURN
`
		}
		commands += `
# 本机发出去的 TCP 和 UDP 走一下 SETMARK 链
ip6tables -w 2 -t mangle -A TP_OUT -p tcp -m mark ! --mark 1 -j SETMARK
ip6tables -w 2 -t mangle -A TP_OUT -p udp -m mark ! --mark 1 -j SETMARK

# 走过TPROXY的通行
ip6tables -w 2 -t mangle -A TP_PRE -m mark --mark 0xff -j RETURN
`
		if specialMode.ShouldLocalDnsListen() {
			commands += ` 
ip6tables -w 2 -t mangle -A TP_PRE -p udp --dport 53 -j RETURN
`
		}
		commands += `
# 让内网主机发出的 TCP 和 UDP 走一下 SETMARK 链
ip6tables -w 2 -t mangle -A TP_PRE -p tcp -m mark ! --mark 1 -j SETMARK
ip6tables -w 2 -t mangle -A TP_PRE -p udp -m mark ! --mark 1 -j SETMARK
# 将所有打了标记的 TCP 和 UDP 包透明地转发到代理的监听端口
ip6tables -w 2 -t mangle -A TP_PRE -m mark --mark 1 -p tcp -j TPROXY --on-port 32345 --tproxy-mark 1
ip6tables -w 2 -t mangle -A TP_PRE -m mark --mark 1 -p udp -j TPROXY --on-port 32345 --tproxy-mark 1
	`
	}
	return SetupCommands(commands)
}

func (t *tproxy) GetCleanCommands() CleanCommands {
	commands := `
ip rule del fwmark 1 table 100 
ip route del local 0.0.0.0/0 dev lo table 100

iptables -w 2 -t mangle -F TP_OUT
iptables -w 2 -t mangle -D OUTPUT -j TP_OUT
iptables -w 2 -t mangle -X TP_OUT
iptables -w 2 -t mangle -F TP_PRE
iptables -w 2 -t mangle -D PREROUTING -j TP_PRE
iptables -w 2 -t mangle -X TP_PRE
iptables -w 2 -t mangle -F SETMARK
iptables -w 2 -t mangle -X SETMARK
`
	if IsIPv6Supported() {
		commands += `
ip -6 rule del fwmark 1 table 100
ip -6 route del local ::/0 dev lo table 100

ip6tables -w 2 -t mangle -F TP_OUT
ip6tables -w 2 -t mangle -D OUTPUT -j TP_OUT
ip6tables -w 2 -t mangle -X TP_OUT
ip6tables -w 2 -t mangle -F TP_PRE
ip6tables -w 2 -t mangle -D PREROUTING -j TP_PRE
ip6tables -w 2 -t mangle -X TP_PRE
ip6tables -w 2 -t mangle -F SETMARK
ip6tables -w 2 -t mangle -X SETMARK
`
	}
	return CleanCommands(commands)
}
