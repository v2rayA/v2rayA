package specialMode

import (
	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2rayA/v2rayA/common/netTools"
	"github.com/v2rayA/v2rayA/core/specialMode/infra"
	"github.com/v2rayA/v2rayA/db/configure"
	"log"
	"sync"
	"time"
)

var (
	poison            = infra.New()
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

func CouldUseSupervisor() bool {
	return true
}

func ShouldUseSupervisor() bool {
	return configure.GetSettingNotNil().SpecialMode == configure.SpecialModeSupervisor
}

func CheckAndSetupDNSSupervisor() {
	if !ShouldUseSupervisor() {
		return
	}
	_ = StartDNSSupervisor(nil,
		nil)
}

func StartDNSSupervisor(externWhiteDnsServers []*router.CIDR, externWhiteDomains []*router.Domain) (err error) {
	defer func() {
		if err != nil {
			err = newError("StartDNSSupervisor").Base(err)
		}
	}()
	mutex.Lock()
	if done != nil {
		select {
		case <-done:
			//done has closed
		default:
			mutex.Unlock()
			return newError("DNSSupervisor is running")
		}
	}
	done = make(chan interface{})
	whiteDnsServerIps = externWhiteDnsServers
	whiteDomains = externWhiteDomains
	mutex.Unlock()
	go func(poison *infra.DnsSupervisor) {
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
				log.Println("[DnsSupervisor] preparing whitelist")
				wlDms, err := infra.GetWhitelistCn(nil)
				//var wlDms = new(strmatcher.MatcherGroup)
				if err != nil {
					log.Println("StartDNSSupervisorConroutine:", err)
					return
				}
				ipMatcher := new(router.GeoIPMatcher)
				_ = ipMatcher.Init(whiteDnsServerIps)
				for _, ifname := range ifnames {
					if _, ok := mHandles[ifname]; !ok {
						err = poison.Prepare(ifname)
						if err != nil {
							log.Println("StartDNSSupervisorConroutine["+ifname+"]:", err)
							return
						}
						go func(ifname string) {
							wg.Add(1)
							defer wg.Done()
							err = poison.Run(ifname, ipMatcher, wlDms)
							if err != nil {
								log.Println("StartDNSSupervisorConroutine["+ifname+"]:", err)
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

func StopDNSSupervisor() {
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
}
