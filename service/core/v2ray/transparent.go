package v2ray

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"log"
	"strings"
	"time"
)

func DeleteTransparentProxyRules() {
	removeResolvHijacker()
	iptables.CloseWatcher()
	iptables.Tproxy.GetCleanCommands().Clean()
	iptables.Redirect.GetCleanCommands().Clean()
	iptables.DropSpoofing.GetCleanCommands().Clean()
	time.Sleep(100 * time.Millisecond)
}

func WriteTransparentProxyRules(preprocess *func(c *iptables.SetupCommands)) (err error) {
	defer func() {
		if err != nil {
			log.Println(err)
			DeleteTransparentProxyRules()
		}
	}()
	if specialMode.ShouldUseSupervisor() {
		if err = iptables.DropSpoofing.GetSetupCommands().Setup(preprocess); err != nil {
			err = newError("[WARNING] DropSpoofing can't be enable").Base(err)
			return err
		}
	}
	setting := configure.GetSettingNotNil()
	if setting.TransparentType == configure.TransparentTproxy {
		if err = iptables.Tproxy.GetSetupCommands().Setup(preprocess); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = newError("you does not compile xt_TPROXY in kernel")
			}
			return fmt.Errorf("not support \"tproxy\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(&iptables.Tproxy)
	} else if setting.TransparentType == configure.TransparentRedirect {
		if err = iptables.Redirect.GetSetupCommands().Setup(preprocess); err != nil {
			return newError("not support \"redirect\" mode of transparent proxy: ").Base(err)
		}
		iptables.SetWatcher(&iptables.Redirect)
	}
	if specialMode.ShouldLocalDnsListen() {
		if couldListenLocalhost, e := specialMode.CouldLocalDnsListen(); couldListenLocalhost {
			if e != nil {
				log.Printf("[Warning] %v, only listen at 127.2.0.17", e)
			}
			resetResolvHijacker()
		} else if specialMode.ShouldUseFakeDns() {
			return fmt.Errorf("fakedns cannot be enabled: %w", e)
		} else {
			log.Printf("[Warning] %v", e)
		}
	}
	return nil
}

func nextPortsGroup(ports []string, groupSize int) (group []string, remain []string) {
	var cnt int
	for i := range ports {
		if strings.ContainsRune(ports[i], ':') {
			cnt += 2
		} else {
			cnt++
		}
		if cnt == groupSize {
			return ports[:i+1], ports[i+1:]
		} else if cnt > groupSize {
			return ports[:i], ports[i:]
		}
	}
	if len(ports) > 0 {
		return ports, nil
	}
	return nil, nil
}

func CheckAndSetupTransparentProxy(checkRunning bool) (err error) {
	setting := configure.GetSettingNotNil()
	preprocess := func(c *iptables.SetupCommands) {
		commands := string(*c)
		//先看要不要把自己的端口加进去
		selfPort := strings.Split(global.GetEnvironmentConfig().Address, ":")[1]
		wl := configure.GetPortWhiteListNotNil()
		if !wl.Has(selfPort, "tcp") {
			wl.TCP = append(wl.TCP, selfPort)
		}
		lines := strings.Split(commands, "\n")
		for i, line := range lines {
			if strings.Contains(line, "{{TCP_PORTS}}") {
				raw := line
				lines[i] = ""
				var grp []string
				r := wl.TCP
				for r != nil {
					grp, r = nextPortsGroup(r, 15)
					if grp != nil {
						lines[i] += strings.Replace(raw, "{{TCP_PORTS}}", strings.Join(grp, ","), 1) + "\n"
					}
				}
				lines[i] = strings.TrimSuffix(lines[i], "\n")
			} else if strings.Contains(line, "{{UDP_PORTS}}") {
				raw := line
				lines[i] = ""
				var grp []string
				r := wl.UDP
				for r != nil {
					grp, r = nextPortsGroup(r, 15)
					if grp != nil {
						lines[i] += strings.Replace(raw, "{{UDP_PORTS}}", strings.Join(grp, ","), 1) + "\n"
					}
				}
				lines[i] = strings.TrimSuffix(lines[i], "\n")
			}
		}
		commands = strings.Join(lines, "\n")
		//if setting.AntiPollution == configure.AntipollutionClosed {
		//	commands = common.TrimLineContains(commands, "udp")
		//}
		if specialMode.ShouldUseSupervisor() || specialMode.ShouldUseFakeDns() {
			commands = common.TrimLineContains(commands, "198.18.0.0/15")
		}
		*c = iptables.SetupCommands(commands)
	}
	if (!checkRunning || IsV2RayRunning()) && setting.Transparent != configure.TransparentClose {
		DeleteTransparentProxyRules()
		err = WriteTransparentProxyRules(&preprocess)
	}
	return
}

func CheckAndStopTransparentProxy() {
	DeleteTransparentProxyRules()
}
