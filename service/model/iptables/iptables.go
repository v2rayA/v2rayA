package iptables

import (
	"V2RayA/global"
	"V2RayA/persistence/configure"
	"V2RayA/tools/cmds"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type IpTablesMangle struct {
	bakMangle *string
}

func execCommands(commands string, stopWhenError bool) error {
	lines := strings.Split(commands, "\n")
	var e error
	for _, line := range lines {
		if len(line) <= 0 || strings.HasPrefix(line, "#") {
			continue
		}
		out, err := exec.Command("sh", "-c", line).CombinedOutput()
		if err != nil {
			e = errors.New(line + " " + err.Error() + " " + string(out))
			if stopWhenError {
				return e
			}
		}
	}
	return e
}

func backupRules() (string, error) {
	out, err := exec.Command("sh", "-c", "iptables-save -t mangle").CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (t *IpTablesMangle) BackupRules() error {
	s, err := backupRules()
	if err != nil {
		return err
	}
	t.bakMangle = &s
	return nil
}

func restoreRules(r string) error {
	tmpfile, err := ioutil.TempFile("", "V2RayAbak.*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.WriteString(r); err != nil {
		tmpfile.Close()
		return err
	}
	out, err := exec.Command("sh", "-c", "iptables-restore < "+tmpfile.Name()).CombinedOutput()
	if err != nil {
		return errors.New(err.Error() + string(out))
	}
	if err = tmpfile.Close(); err != nil {
		return err
	}
	commands := `
# 删除策略路由
ip rule del fwmark 1 table 100 
ip route del local 0.0.0.0/0 dev lo table 100
`
	err = execCommands(commands, false)
	return err
}

func (t *IpTablesMangle) RestoreRules() error {
	if t.bakMangle == nil {
		return errors.New("你还没有备份过iptables")
	}
	return restoreRules(*t.bakMangle)
}

func DeleteRules() error {
	commands := `
ip rule del fwmark 1 table 100 
ip route del local 0.0.0.0/0 dev lo table 100

iptables -t mangle -F SSTP_OUT
iptables -t mangle -D OUTPUT -j SSTP_OUT
iptables -t mangle -X SSTP_OUT
iptables -t mangle -F SSTP_PRE
iptables -t mangle -D PREROUTING -j SSTP_PRE
iptables -t mangle -X SSTP_PRE
iptables -t mangle -F LOG
iptables -t mangle -F SETMARK
iptables -t mangle -X SETMARK
`
	if cmds.IsCommandValid("ip6tables") {
		commands += `
ip -6 rule del fwmark 1 table 100
ip -6 route del local ::/0 dev lo table 100

ip6tables -t mangle -F SSTP_OUT
ip6tables -t mangle -D OUTPUT -j SSTP_OUT
ip6tables -t mangle -X SSTP_OUT
ip6tables -t mangle -F SSTP_PRE
ip6tables -t mangle -D PREROUTING -j SSTP_PRE
ip6tables -t mangle -X SSTP_PRE
ip6tables -t mangle -F LOG
ip6tables -t mangle -F SETMARK
ip6tables -t mangle -X SETMARK
`
	}
	if global.ServiceControlMode == global.DockerMode {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	return execCommands(commands, false)
}

func WriteRules() error {
	commands := `
# 设置策略路由
ip rule add fwmark 1 table 100
ip route add local 0.0.0.0/0 dev lo table 100

# 建链
iptables -t mangle -N SSTP_OUT
iptables -t mangle -N SSTP_PRE
iptables -t mangle -A OUTPUT -j SSTP_OUT
iptables -t mangle -A PREROUTING -j SSTP_PRE

# 打上 iptables 标记，mark 了的会走代理
iptables -t mangle -N SETMARK
iptables -t mangle -A SETMARK -i docker+ -j RETURN
iptables -t mangle -A SETMARK -i br-+ -j RETURN
`
	//s := configure.GetSettingNotNil()
	//if s.AntiPollution != configure.AntipollutionNone {
	if true {
		commands += `
iptables -t mangle -A SETMARK -p udp --dport 53 -j MARK --set-mark 1
iptables -t mangle -A SETMARK -p tcp --dport 53 -j MARK --set-mark 1
`
	} else {
		commands += `
iptables -t mangle -A SETMARK -p udp --dport 53 -j RETURN
iptables -t mangle -A SETMARK -p tcp --dport 53 -j RETURN
`
	}
	commands += `
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
# 本机出方向规则，白名单端口
iptables -t mangle -A SSTP_OUT -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
iptables -t mangle -A SSTP_OUT -p udp -m multiport --sports {{UDP_PORTS}} -j RETURN
# 本机发出去的 TCP 和 UDP 走一下 SETMARK 链
iptables -t mangle -A SSTP_OUT -p tcp -m mark ! --mark 1 -j SETMARK
iptables -t mangle -A SSTP_OUT -p udp -m mark ! --mark 1 -j SETMARK

# 走过TPROXY的通行
iptables -t mangle -A SSTP_PRE -m mark --mark 0xff -j RETURN
# 让内网主机发出的 TCP 和 UDP 走一下 SETMARK 链
iptables -t mangle -A SSTP_PRE -p tcp -m mark ! --mark 1 -j SETMARK
iptables -t mangle -A SSTP_PRE -p udp -m mark ! --mark 1 -j SETMARK
# 将所有打了标记的 TCP 和 UDP 包透明地转发到代理的监听端口
iptables -t mangle -A SSTP_PRE -m mark --mark 1 -p tcp -j TPROXY --on-port 12345
iptables -t mangle -A SSTP_PRE -m mark --mark 1 -p udp -j TPROXY --on-port 12345
`
	if cmds.IsCommandValid("ip6tables") {
		commands += `
# 开启ipv6 forward
sysctl -w net.ipv6.conf.all.forwarding=1

# 设置策略路由
ip -6 rule add fwmark 1 table 100
ip -6 route add local ::/0 dev lo table 100

# 建链
ip6tables -t mangle -N SSTP_OUT
ip6tables -t mangle -N SSTP_PRE
ip6tables -t mangle -A OUTPUT -j SSTP_OUT
ip6tables -t mangle -A PREROUTING -j SSTP_PRE

# 打上 iptables 标记，mark 了的会走代理
ip6tables -t mangle -N SETMARK
ip6tables -t mangle -A SETMARK -i docker+ -j RETURN
	`
		if true {
			commands += `
ip6tables -t mangle -A SETMARK -p udp --dport 53 -j MARK --set-mark 1
ip6tables -t mangle -A SETMARK -p tcp --dport 53 -j MARK --set-mark 1
	`
		} else {
			commands += `
ip6tables -t mangle -A SETMARK -p udp --dport 53 -j RETURN
ip6tables -t mangle -A SETMARK -p tcp --dport 53 -j RETURN
	`
		}
		commands += `
ip6tables -t mangle -A SETMARK -d ::/128 -j RETURN
ip6tables -t mangle -A SETMARK -d ::1/128 -j RETURN
ip6tables -t mangle -A SETMARK -d ::ffff:0:0/96 -j RETURN
ip6tables -t mangle -A SETMARK -d ::ffff:0:0:0/96 -j RETURN
ip6tables -t mangle -A SETMARK -d 64:ff9b::/96 -j RETURN
ip6tables -t mangle -A SETMARK -d 100::/64 -j RETURN
ip6tables -t mangle -A SETMARK -d 2001::/32 -j RETURN
ip6tables -t mangle -A SETMARK -d 2001:20::/28 -j RETURN
ip6tables -t mangle -A SETMARK -d 2001:db8::/32 -j RETURN
ip6tables -t mangle -A SETMARK -d 2002::/16 -j RETURN
ip6tables -t mangle -A SETMARK -d fc00::/7 -j RETURN
ip6tables -t mangle -A SETMARK -d fe80::/10 -j RETURN
ip6tables -t mangle -A SETMARK -d ff00::/8 -j RETURN
ip6tables -t mangle -A SETMARK -p tcp -j MARK --set-mark 1
ip6tables -t mangle -A SETMARK -p udp -j MARK --set-mark 1

# 走过TPROXY的通行
#禁用本地ipv6
ip6tables -t mangle -A SSTP_OUT -p tcp -j DROP
ip6tables -t mangle -A SSTP_OUT -p udp -j DROP
ip6tables -t mangle -A SSTP_OUT -m mark --mark 0xff -j RETURN
# 本机出方向规则，白名单端口
ip6tables -t mangle -A SSTP_OUT -p tcp -m multiport --sports {{TCP_PORTS}} -j RETURN
ip6tables -t mangle -A SSTP_OUT -p udp -m multiport --sports {{UDP_PORTS}} -j RETURN
# 本机发出去的 TCP 和 UDP 走一下 SETMARK 链
ip6tables -t mangle -A SSTP_OUT -p tcp -m mark ! --mark 1 -j SETMARK
ip6tables -t mangle -A SSTP_OUT -p udp -m mark ! --mark 1 -j SETMARK

# 走过TPROXY的通行
ip6tables -t mangle -A SSTP_PRE -m mark --mark 0xff -j RETURN
# 让内网主机发出的 TCP 和 UDP 走一下 SETMARK 链
ip6tables -t mangle -A SSTP_PRE -p tcp -m mark ! --mark 1 -j SETMARK
ip6tables -t mangle -A SSTP_PRE -p udp -m mark ! --mark 1 -j SETMARK
# 将所有打了标记的 TCP 和 UDP 包透明地转发到代理的监听端口
ip6tables -t mangle -A SSTP_PRE -m mark --mark 1 -p tcp -j TPROXY --on-port 12345
ip6tables -t mangle -A SSTP_PRE -m mark --mark 1 -p udp -j TPROXY --on-port 12345
	`
	}
	//参考http://briteming.hatenablog.com/entry/2019/06/18/175518
	//先看要不要把自己的端口加进去
	selfPort := strings.Split(global.GetEnvironmentConfig().Address, ":")[1]
	wl := configure.GetPortWhiteListNotNil()
	if !wl.Has(selfPort, "tcp") {
		wl.TCP = append(wl.TCP, selfPort)
	}
	commands = strings.ReplaceAll(commands, "{{TCP_PORTS}}", strings.Join(wl.TCP, ","))
	if len(wl.UDP) > 0 {
		commands = strings.ReplaceAll(commands, "{{UDP_PORTS}}", strings.Join(wl.UDP, ","))
	} else { //没有UDP端口就把这行删了
		lines := strings.Split(commands, "\n")
		for i, line := range lines {
			if strings.Contains(line, "{{UDP_PORTS}}") {
				lines = append(lines[:i], lines[i+1:]...)
				break
			}
		}
		commands = strings.Join(lines, "\n")
	}
	if global.ServiceControlMode == global.DockerMode {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	if err := execCommands(commands, true); err != nil {
		_ = DeleteRules()
		if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
			err = errors.New("内核未编译xt_TPROXY")
		}
		return err
	}
	return nil
}
