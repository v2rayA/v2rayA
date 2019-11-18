package transparentProxy

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type IpTablesMangle struct {
	bakMangle *string
}

func execCommands(commands string) error {
	lines := strings.Split(commands, "\n")
	for _, line := range lines {
		if len(line) <= 0 || strings.HasPrefix(line, "#") {
			continue
		}
		_, err := exec.Command("sh", "-c", line).CombinedOutput()
		if err != nil {
			return err
		}
	}
	return nil
}

func backupRules() (string, error) {
	out, err := exec.Command("sh", "-c", "iptables-save -t mangle").CombinedOutput()
	if err != nil {
		return "", err
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
	_, err = exec.Command("sh", "-c", "iptables-restore < "+tmpfile.Name()).CombinedOutput()
	if err != nil {
		return err
	}
	if err = tmpfile.Close(); err != nil {
		return err
	}
	commands := `
# 删除策略路由
ip rule del fwmark 1 table 100 
ip route del local 0.0.0.0/0 dev lo table 100
`
	err = execCommands(commands)
	return err
}

func (t *IpTablesMangle) RestoreRules() error {
	if t.bakMangle == nil {
		return errors.New("你还没有备份过iptables")
	}
	return restoreRules(*t.bakMangle)
}

func (t *IpTablesMangle) WriteRules() error {
	bak, err := backupRules()
	if err != nil {
		return err
	}
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
iptables -t mangle -A V2RAY -d 192.168.0.0/16 -p tcp -j RETURN # 直连局域网，避免 V2Ray 无法启动时无法连网关的 SSH，如果你配置的是其他网段（如 10.x.x.x 等），则修改成自己的
iptables -t mangle -A V2RAY -d 192.168.0.0/16 -p udp ! --dport 53 -j RETURN # 直连局域网，53 端口除外（因为要使用 V2Ray 的 
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
iptables -t mangle -A V2RAY_MASK -d 192.168.0.0/16 -p tcp -j RETURN # 直连局域网
iptables -t mangle -A V2RAY_MASK -d 192.168.0.0/16 -p udp ! --dport 53 -j RETURN # 直连局域网，53 端口除外（因为要使用 V2Ray 的 DNS）
iptables -t mangle -A V2RAY_MASK -j RETURN -m mark --mark 0xff    # 直连 SO_MARK 为 0xff 的流量(0xff 是 16 进制数，数值上等同与上面V2Ray 配置的 255)，此规则目的是避免代理本机(网关)流量出现回环问题
iptables -t mangle -A V2RAY_MASK -p udp -j MARK --set-mark 1   # 给 UDP 打标记,重路由
iptables -t mangle -A V2RAY_MASK -p tcp -j MARK --set-mark 1   # 给 TCP 打标记,重路由
iptables -t mangle -A OUTPUT -j V2RAY_MASK # 应用规则
` // 来自https://guide.v2fly.org/app/tproxy.html
	if err := execCommands(commands); err != nil {
		_ = restoreRules(bak)
		return err
	}
	return nil
}
