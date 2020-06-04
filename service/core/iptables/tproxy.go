package iptables

import "v2rayA/common/cmds"

type tproxy struct{ iptablesSetter }

var Tproxy tproxy

func (t *tproxy) GetSetupCommands() SetupCommands {
	commands := `
# 设置策略路由
ip rule add fwmark 1 table 100
ip route add local 0.0.0.0/0 dev lo table 100

# 建链
iptables -t mangle -N SSTP_OUT
iptables -t mangle -N SSTP_PRE
iptables -t mangle -N SSTP_ONCE
iptables -t mangle -A OUTPUT -j SSTP_OUT
iptables -t mangle -A PREROUTING -m socket --transparent -j SSTP_ONCE
iptables -t mangle -A PREROUTING -j SSTP_PRE

# 打上 iptables 标记，mark 了的会走代理
iptables -t mangle -N SETMARK
# 出方向白名单端口
iptables -t mangle -A SETMARK -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
iptables -t mangle -A SETMARK -p udp -m multiport --sports {{UDP_PORTS}} -j RETURN
iptables -t mangle -A SETMARK -i docker+ -j RETURN
iptables -t mangle -A SETMARK -i veth+ -j RETURN
iptables -t mangle -A SETMARK -i br-+ -j RETURN
iptables -t mangle -A SETMARK -p udp --dport 53 -j MARK --set-mark 1
iptables -t mangle -A SETMARK -p tcp --dport 53 -j MARK --set-mark 1
iptables -t mangle -A SETMARK -d 10.0.0.0/8 -j RETURN
iptables -t mangle -A SETMARK -d 100.64.0.0/10 -j RETURN
iptables -t mangle -A SETMARK -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A SETMARK -d 169.254.0.0/16 -j RETURN
iptables -t mangle -A SETMARK -d 172.16.0.0/12 -j RETURN
iptables -t mangle -A SETMARK -d 192.0.0.0/24 -j RETURN
iptables -t mangle -A SETMARK -d 192.0.2.0/24 -j RETURN
iptables -t mangle -A SETMARK -d 192.88.99.0/24 -j RETURN
iptables -t mangle -A SETMARK -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A SETMARK -d 198.18.0.0/15 -j RETURN
iptables -t mangle -A SETMARK -d 198.51.100.0/24 -j RETURN
iptables -t mangle -A SETMARK -d 203.0.113.0/24 -j RETURN
iptables -t mangle -A SETMARK -d 224.0.0.0/4 -j RETURN
iptables -t mangle -A SETMARK -d 240.0.0.0/4 -j RETURN
iptables -t mangle -A SETMARK -d 255.255.255.255/32 -j RETURN
iptables -t mangle -A SETMARK -p tcp -j MARK --set-mark 1
iptables -t mangle -A SETMARK -p udp -j MARK --set-mark 1

# 走过TPROXY的通行
iptables -t mangle -A SSTP_OUT -m mark --mark 0xff -j RETURN
# 本机发出去的 TCP 和 UDP 走一下 SETMARK 链
iptables -t mangle -A SSTP_OUT -p tcp -m mark ! --mark 1 -j SETMARK
iptables -t mangle -A SSTP_OUT -p udp -m mark ! --mark 1 -j SETMARK

# 走过TPROXY的通行
iptables -t mangle -A SSTP_PRE -m mark --mark 0xff -j RETURN
# 让内网主机发出的 TCP 和 UDP 走一下 SETMARK 链
iptables -t mangle -A SSTP_PRE -p tcp -m mark ! --mark 1 -j SETMARK
iptables -t mangle -A SSTP_PRE -p udp -m mark ! --mark 1 -j SETMARK
# 将所有打了标记的 TCP 和 UDP 包透明地转发到代理的监听端口
iptables -t mangle -A SSTP_PRE -m mark --mark 1 -p tcp -j TPROXY --on-port 32345 --tproxy-mark 1
iptables -t mangle -A SSTP_PRE -m mark --mark 1 -p udp -j TPROXY --on-port 32345 --tproxy-mark 1

# 略过已建立且被tproxy标记的的socket
iptables -t mangle -A SSTP_ONCE -j MARK --set-mark 1
iptables -t mangle -A SSTP_ONCE -j ACCEPT

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

func (t *tproxy) GetCleanCommands() CleanCommands {
	commands := `
ip rule del fwmark 1 table 100 
ip route del local 0.0.0.0/0 dev lo table 100

iptables -t mangle -F SSTP_ONCE
iptables -t mangle -D PREROUTING -m socket --transparent -j SSTP_ONCE
iptables -t mangle -X SSTP_ONCE
iptables -t mangle -F SSTP_OUT
iptables -t mangle -D OUTPUT -j SSTP_OUT
iptables -t mangle -X SSTP_OUT
iptables -t mangle -F SSTP_PRE
iptables -t mangle -D PREROUTING -j SSTP_PRE
iptables -t mangle -X SSTP_PRE
iptables -t mangle -F SETMARK
iptables -t mangle -X SETMARK
`
	if cmds.IsCommandValid("sysctl") {
		commands += `
sysctl -w net.ipv6.conf.all.disable_ipv6=0
sysctl -w net.ipv6.conf.default.disable_ipv6=0
`
	}
	return CleanCommands(commands)
}
