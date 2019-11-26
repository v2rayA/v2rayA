package transparentProxy

import (
	"V2RayA/global"
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
			e = errors.New(err.Error() + string(out))
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
# 删除策略路由
ip rule del fwmark 1 table 100 
ip route del local 0.0.0.0/0 dev lo table 100
# 删除iptables链
iptables -t mangle -F V2RAY
iptables -t mangle -D PREROUTING -j V2RAY
iptables -t mangle -X V2RAY
iptables -t mangle -F V2RAY_MASK
iptables -t mangle -D OUTPUT -j V2RAY_MASK
iptables -t mangle -X V2RAY_MASK
`
	if global.ServiceControlMode == global.DockerMode {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
	}
	return execCommands(commands, false)
}

func WriteRules() error {
	commands := `
# 设置策略路由
ip rule add fwmark 1 table 100 
ip route add local 0.0.0.0/0 dev lo table 100

# 代理局域网设备
iptables -t mangle -N V2RAY
iptables -t mangle -A V2RAY -d 0.0.0.0/8 -j RETURN
iptables -t mangle -A V2RAY -d 10.0.0.0/8 -j RETURN
iptables -t mangle -A V2RAY -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A V2RAY -d 169.254.0.0/16 -j RETURN
iptables -t mangle -A V2RAY -d 172.16.0.0/12 -j RETURN
iptables -t mangle -A V2RAY -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A V2RAY -d 224.0.0.0/4 -j RETURN
iptables -t mangle -A V2RAY -d 240.0.0.0/4 -j RETURN
iptables -t mangle -A V2RAY -d 255.255.255.255/32 -j RETURN 
iptables -t mangle -A V2RAY -i docker+ -j RETURN
iptables -t mangle -A V2RAY -p udp -j TPROXY --on-port 12345 --tproxy-mark 1 # 给 UDP 打标记 1，转发至 12345 端口
iptables -t mangle -A V2RAY -p tcp -j TPROXY --on-port 12345 --tproxy-mark 1 # 给 TCP 打标记 1，转发至 12345 端口
iptables -t mangle -A PREROUTING -j V2RAY # 应用规则

# 代理网关本机
iptables -t mangle -N V2RAY_MASK
iptables -t mangle -A V2RAY_MASK -d 0.0.0.0/8 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 10.0.0.0/8 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 169.254.0.0/16 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 172.16.0.0/12 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 224.0.0.0/4 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 240.0.0.0/4 -j RETURN
iptables -t mangle -A V2RAY_MASK -d 224.0.0.0/4 -j RETURN 
iptables -t mangle -A V2RAY_MASK -d 255.255.255.255/32 -j RETURN 
iptables -t mangle -A V2RAY_MASK -j RETURN -m mark --mark 0xff    # 直连 SO_MARK 为 0xff 的流量(0xff 是 16 进制数，数值上等同与上面V2Ray 配置的 255)，此规则目的是避免代理本机(网关)流量出现回环问题
iptables -t mangle -A V2RAY -i docker+ -j RETURN
iptables -t mangle -A V2RAY_MASK -p udp -j MARK --set-mark 1   # 给 UDP 打标记,重路由
iptables -t mangle -A V2RAY_MASK -p tcp -j MARK --set-mark 1   # 给 TCP 打标记,重路由
iptables -t mangle -A OUTPUT -j V2RAY_MASK # 应用规则
` // 来自https://guide.v2fly.org/app/tproxy.html
	if global.ServiceControlMode == global.DockerMode {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
	}
	//避免docker冲突
	//TODO
	if err := execCommands(commands, true); err != nil {
		_ = DeleteRules()
		return err
	}
	return nil
}
