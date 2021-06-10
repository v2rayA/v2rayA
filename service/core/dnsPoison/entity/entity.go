package entity

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/netTools"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/core/dnsPoison"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
	"v2ray.com/core/app/router"
)

var (
	poison            = dnsPoison.New()
	done              chan interface{}
	mutex             sync.Mutex
	limit             = make(chan interface{}, 1)
	whiteDnsServerIps []*router.CIDR
	whiteDomains      []*router.Domain
	wg                sync.WaitGroup
)

type ExtraInfo struct {
	DohIps       []string
	DohDomains   []string
	ServerIps    []string
	ServerDomain string
}

func dnsPortValid() bool {
	occupied, socket, err := ports.IsPortOccupied([]string{"53:udp"})
	if err != nil {
		return false
	}

	if occupied {
		p, err := socket.Process()
		if err != nil {
			return false
		}
		if p.PPID == strconv.Itoa(os.Getpid()) {
			return true
		}
		log.Printf("[info] port 53 is occupied by %v(%v)\n", p.Name, p.PID)
		return false
	}
	return true
}

/*
0: neither

1: redirect + poison

2: redirect + fakedns
*/
func ShouldDnsPoisonOpen() int {
	setting := configure.GetSettingNotNil()
	if setting.Transparent == configure.TransparentClose ||
		setting.AntiPollution == configure.AntipollutionClosed ||
		(global.SupportTproxy && !setting.EnhancedMode) {
		return 0
	}
	ver, err := where.GetV2rayServiceVersion()
	if err != nil {
		ver = "0.0.0"
	}
	fakednsValid, _ := common.VersionGreaterEqual(ver, "4.35.0")
	if fakednsValid && configure.GetSettingNotNil().Transparent == configure.TransparentClose {
		fakednsValid = false
	}
	if fakednsValid && !dnsPortValid() {
		log.Println("[fakedns] unable to use fakedns: port 53 is occupied")
		fakednsValid = false
	}
	if !fakednsValid {
		return 1
	}
	return 2
}

func CheckAndSetupDnsPoisonWithExtraInfo(info *ExtraInfo) {
	if ShouldDnsPoisonOpen() != 1 {
		return
	}
	whitedms := make([]*router.Domain, 0, len(info.DohDomains))
	for _, h := range info.DohDomains {
		whitedms = append(whitedms, &router.Domain{
			Type:  router.Domain_Full,
			Value: h,
		})
	}
	if len(info.ServerDomain) > 0 {
		whitedms = append(whitedms, &router.Domain{
			Type:  router.Domain_Full,
			Value: info.ServerDomain,
		})
	}
	whitedms = append(whitedms, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "v2raya.mzz.pub",
	}, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "v.mzz.pub",
	}, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "github.com",
	}, &router.Domain{
		Type:  router.Domain_Domain,
		Value: "1password.com",
	}, &router.Domain{
		Type:  router.Domain_Regex,
		Value: `^dns\.`,
	}, &router.Domain{
		Type:  router.Domain_Regex,
		Value: `^doh\.`,
	})
	_ = StartDNSPoison(nil,
		whitedms)
}

func StartDNSPoison(externWhiteDnsServers []*router.CIDR, externWhiteDomains []*router.Domain) (err error) {
	defer func() {
		if err != nil {
			err = newError("StartDNSPoison").Base(err)
		}
	}()
	mutex.Lock()
	if done != nil {
		select {
		case <-done:
			//done has closed
		default:
			mutex.Unlock()
			return newError("DNSPoison is running")
		}
	}
	done = make(chan interface{})
	whiteDnsServerIps = externWhiteDnsServers
	whiteDomains = externWhiteDomains
	mutex.Unlock()
	go func(poison *dnsPoison.DnsPoison) {
		//并发限制1
		select {
		case limit <- nil:
		default:
			return
		}
		defer func() { <-limit }()
	out:
		for {
			//随时准备应对default interface变化
			f := func() {
				ifnames, err := netTools.GetDefaultInterfaceName()
				if err != nil {
					return
				}
				mIfnames := make(map[string]interface{})
				mHandles := make(map[string]interface{})
				needToAdd := false
				for _, ifname := range ifnames {
					mIfnames[ifname] = nil
					if !poison.Exists(ifname) {
						needToAdd = true
					}
				}
				hs := poison.ListHandles()
				for _, h := range hs {
					mHandles[h] = nil
					if _, ok := mIfnames[h]; !ok {
						_ = poison.DeleteHandles(h)
					}
				}
				if !needToAdd {
					return
				}
				//准备白名单
				log.Println("[DnsPoison] preparing whitelist")
				wlDms, err := asset.GetWhitelistCn(nil, whiteDomains)
				//var wlDms = new(strmatcher.MatcherGroup)
				if err != nil {
					log.Println("StartDNSPoisonConroutine:", err)
					return
				}
				ipMatcher := new(router.GeoIPMatcher)
				_ = ipMatcher.Init(whiteDnsServerIps)
				for _, ifname := range ifnames {
					if _, ok := mHandles[ifname]; !ok {
						err = poison.Prepare(ifname)
						if err != nil {
							log.Println("StartDNSPoisonConroutine["+ifname+"]:", err)
							return
						}
						go func(ifname string) {
							wg.Add(1)
							defer wg.Done()
							err = poison.Run(ifname, ipMatcher, wlDms)
							if err != nil {
								log.Println("StartDNSPoisonConroutine["+ifname+"]:", err)
							}
						}(ifname)
					}
				}
			}
			f()
			select {
			case <-done:
				poison.Clear()
				break out
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}(poison)
	return nil
}

func StopDNSPoison() {
	mutex.Lock()
	defer mutex.Unlock()
	if done != nil {
		select {
		case <-done:
			// channel 'done' has been closed
		default:
			close(done)
		}
	}
	wg.Wait()
	clearDNSCache()
}

func clearDNSCache() {
	switch global.ServiceControlMode {
	case global.ServiceMode:
		_, _ = exec.Command("sh -c", "service nscd restart").Output()
	case global.SystemctlMode:
		_, _ = exec.Command("sh -c", "systemctl restart nscd").Output()
	}
}
