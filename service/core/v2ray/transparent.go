package v2ray

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/core/dnsPoison/entity"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	resolverFile  = "/etc/resolv.conf"
	checkInterval = 3 * time.Second
)

type DnsHijacker struct {
	ticker  *time.Ticker
	fakedns bool
}

func NewDnsHijacker() *DnsHijacker {
	hij := DnsHijacker{
		ticker:  time.NewTicker(checkInterval),
		fakedns: entity.ShouldDnsPoisonOpen() == 2,
	}
	hij.HijackDNS()
	go func() {
		for range hij.ticker.C {
			hij.HijackDNS()
		}
	}()
	return &hij
}
func (h *DnsHijacker) Close() error {
	h.ticker.Stop()
	return nil
}

var hijacker *DnsHijacker

func (h *DnsHijacker) HijackDNS() error {
	alternatives := []string{
		"223.5.5.5",
		"119.29.29.29",
		"223.6.6.6",
		"180.76.76.76",
		"114.114.114.114",
		"208.67.222.222",
	}
	m := make(map[string]struct{})
	userSetDns := configure.GetDnsListNotNil()
	for _, dns := range userSetDns {
		m[dns] = struct{}{}
	}
	cnt := 0
	maxcnt := 2
	if h.fakedns {
		// fakedns
		alternatives = append([]string{"127.0.0.1"}, alternatives...)
	}
	var builder strings.Builder
	builder.WriteString("# v2rayA DNS hijack\n")
	for _, dns := range alternatives {
		// should not be duplicated with user preset dns
		if _, ok := m[dns]; !ok {
			builder.WriteString("nameserver " + dns + "\n")
			cnt++
			if cnt >= maxcnt {
				break
			}
		}
	}
	err := ioutil.WriteFile(resolverFile, []byte(builder.String()), os.FileMode(0644))
	if err != nil {
		err = fmt.Errorf("failed to hijackDNS: [write] %v", err)
	}
	return err
}

func resetDnsHijacker() {
	if hijacker != nil {
		hijacker.Close()
	}
	hijacker = NewDnsHijacker()
}

func removeDnsHijacker() {
	if hijacker != nil {
		hijacker.Close()
		hijacker = nil
	}
}

func DeleteTransparentProxyRules() {
	removeDnsHijacker()
	iptables.CloseWatcher()
	iptables.Tproxy.GetCleanCommands().Clean()
	iptables.Redirect.GetCleanCommands().Clean()
	iptables.DropSpoofing.GetCleanCommands().Clean()
	time.Sleep(100 * time.Millisecond)
}

func WriteTransparentProxyRules(preprocess *func(c *iptables.SetupCommands)) error {
	if entity.ShouldDnsPoisonOpen() == 1 {
		if e := iptables.DropSpoofing.GetSetupCommands().Setup(preprocess); e != nil {
			log.Println(newError("[WARNING] DropSpoofing can't be enable").Base(e))
			iptables.DropSpoofing.GetCleanCommands().Clean()
		}
	}
	setting := configure.GetSettingNotNil()
	if global.SupportTproxy && !setting.EnhancedMode {
		if err := iptables.Tproxy.GetSetupCommands().Setup(preprocess); err == nil {
			if setting.AntiPollution != configure.AntipollutionClosed {
				resetDnsHijacker()
			}
			iptables.SetWatcher(&iptables.Tproxy)
		} else {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = newError("not compile xt_TPROXY in kernel")
			}
			DeleteTransparentProxyRules()
			log.Println(err)
			global.SupportTproxy = false
		}
	} else {
		if err := iptables.Redirect.GetSetupCommands().Setup(preprocess); err == nil {
			if setting.AntiPollution != configure.AntipollutionClosed {
				resetDnsHijacker()
			}
			iptables.SetWatcher(&iptables.Redirect)
		} else {
			log.Println(err)
			DeleteTransparentProxyRules()
			return newError("not support transparent proxy: ").Base(err)
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
	v2rayPath, err := where.GetV2rayBinPath()
	if err != nil {
		return
	}
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
		if setting.AntiPollution == configure.AntipollutionClosed {
			commands = common.TrimLineContains(commands, "udp")
		}
		if entity.ShouldDnsPoisonOpen() == 1 {
			commands = common.TrimLineContains(commands, "240.0.0.0/4")
		}
		*c = iptables.SetupCommands(commands)
	}
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
			if e == nil && p.Name != path.Base(v2rayPath) {
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
