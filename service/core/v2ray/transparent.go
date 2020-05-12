package v2ray

import (
	"log"
	"strings"
	"time"
	"v2rayA/common/netTools/netstat"
	"v2rayA/common/netTools/ports"
	"v2rayA/core/iptables"
	"v2rayA/global"
	"v2rayA/persistence/configure"
)

func DeleteTransparentProxyRules() {
	iptables.Tproxy.GetCleanCommands().Clean()
	iptables.Redirect.GetCleanCommands().Clean()
	time.Sleep(100 * time.Millisecond)
}

func WriteTransparentProxyRules(preprocess *func(c *iptables.SetupCommands)) error {
	setting := configure.GetSettingNotNil()
	if !(!global.SupportTproxy || setting.EnhancedMode) {
		if err := iptables.Tproxy.GetSetupCommands().Setup(preprocess); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = newError("not compile xt_TPROXY in kernel")
			}
			DeleteTransparentProxyRules()
			log.Println(err)
			global.SupportTproxy = false
		}
	}
	if !global.SupportTproxy || setting.EnhancedMode {
		if err := iptables.Redirect.GetSetupCommands().Setup(preprocess); err != nil {
			log.Println(err)
			DeleteTransparentProxyRules()
			return newError("not support transparent proxy: ").Base(err)
		}
	}
	return nil
}

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	preprocess := func(c *iptables.SetupCommands) {
		commands := string(*c)
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
					lines[i] = ""
				}
			}
			commands = strings.Join(lines, "\n")
		}
		*c = iptables.SetupCommands(commands)
	}
	setting := configure.GetSettingNotNil()
	if (!checkRunning || IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		var (
			o bool
			s *netstat.Socket
		)
		o, s, err = ports.IsPortOccupied([]string{"32345:tcp,udp"})
		if err != nil {
			return
		}
		if o {
			p, e := s.Process()
			if e == nil && p.Name != "v2ray" {
				err = newError("transparent proxy cannot be set up, port 32345 is occupied by ", p.Name)
				return
			}
		}
		DeleteTransparentProxyRules()
		err = WriteTransparentProxyRules(&preprocess)
	}
	return
}

func CheckAndStopTransparentProxy() {
	DeleteTransparentProxyRules()
}
