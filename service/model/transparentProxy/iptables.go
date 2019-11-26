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
ip rule del fwmark 1 table 100 
ip route del local 0/0 dev lo table 100

sudo iptables -t mangle -F SSTP_OUT
sudo iptables -t mangle -D OUTPUT -j SSTP_OUT
sudo iptables -t mangle -X SSTP_OUT
sudo iptables -t mangle -F SSTP_PRE
sudo iptables -t mangle -D PREROUTING -j SSTP_PRE
sudo iptables -t mangle -X SSTP_PRE
sudo iptables -t mangle -F LOG
sudo iptables -t mangle -F SETMARK
sudo iptables -t mangle -X SETMARK
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
ip route add local 0/0 dev lo table 100

# 建链
iptables -t mangle -N SSTP_OUT
iptables -t mangle -N SSTP_PRE
iptables -t mangle -A OUTPUT -j SSTP_OUT
iptables -t mangle -A PREROUTING -j SSTP_PRE

# 打上 iptables 标记，mark 了的会走代理
iptables -t mangle -N SETMARK
iptables -t mangle -A SETMARK -m mark --mark 1 -j RETURN
iptables -t mangle -A SETMARK -m mark --mark 0xff -j RETURN
iptables -t mangle -A SETMARK -i docker+ -j RETURN
iptables -t mangle -A SETMARK -p udp --dport 53 -j MARK --set-mark 1
iptables -t mangle -A SETMARK -d 0.0.0.0/32 -j RETURN
iptables -t mangle -A SETMARK -d 10.0.0.0/8 -j RETURN
iptables -t mangle -A SETMARK -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A SETMARK -d 169.254.0.0/16 -j RETURN
iptables -t mangle -A SETMARK -d 172.16.0.0/12 -j RETURN
iptables -t mangle -A SETMARK -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A SETMARK -d 224.0.0.0/4 -j RETURN
iptables -t mangle -A SETMARK -d 240.0.0.0/4 -j RETURN
iptables -t mangle -A SETMARK -d 255.255.255.255/32 -j RETURN
iptables -t mangle -A SETMARK -s 224.0.0.0/4 -j RETURN
iptables -t mangle -A SETMARK -s 240.0.0.0/4 -j RETURN
iptables -t mangle -A SETMARK -s 255.255.255.255/32 -j RETURN
iptables -t mangle -A SETMARK -p udp -j MARK --set-mark 1
iptables -t mangle -A SETMARK -p tcp -j MARK --set-mark 1

# 本机发出去的 TCP 和 UDP 走一下 SETMARK 链
iptables -t mangle -A SSTP_OUT -p tcp -m mark ! --mark 1 -j SETMARK
iptables -t mangle -A SSTP_OUT -p udp -m mark ! --mark 1 -j SETMARK

# 让内网主机发出的 TCP 和 UDP 走一下 SETMARK 链
iptables -t mangle -A SSTP_PRE -p tcp -m mark ! --mark 1 -j SETMARK
iptables -t mangle -A SSTP_PRE -p udp -m mark ! --mark 1 -j SETMARK
# 将所有打了标记的 TCP 和 UDP 包透明地转发到代理的监听端口
iptables -t mangle -A SSTP_PRE -m mark --mark 1 -p tcp -j TPROXY --on-ip 127.0.0.1 --on-port 12345 --tproxy-mark 1
iptables -t mangle -A SSTP_PRE -m mark --mark 1 -p udp -j TPROXY --on-ip 127.0.0.1 --on-port 12345 --tproxy-mark 1
` //参考http://briteming.hatenablog.com/entry/2019/06/18/175518
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
