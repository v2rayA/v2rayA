package specialMode

import (
	"fmt"
	"github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/specialMode/infra"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
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
	// TODO
	return true
}

func ShouldUseSupervisor() bool {
	if conf.GetEnvironmentConfig().Lite {
		return false
	}
	return configure.GetSettingNotNil().SpecialMode == configure.SpecialModeSupervisor
}

func CheckAndSetupDNSSupervisor() {
	if conf.GetEnvironmentConfig().Lite {
		return
	}
	if !ShouldUseSupervisor() {
		return
	}
	_ = StartDNSSupervisor(nil,
		nil)
}

func StartDNSSupervisor(externWhiteDnsServers []*router.CIDR, externWhiteDomains []*router.Domain) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("StartDNSSupervisor: %w", err)
			log.Warn("%v", err)
		}
	}()
	mutex.Lock()
	if done != nil {
		select {
		case <-done:
			//done has closed
		default:
			mutex.Unlock()
			return fmt.Errorf("DNSSupervisor is running")
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
			//随时准备应对interface变化
			f := func() {
				ifces, err := net.Interfaces()
				if err != nil {
					return
				}
				var ifnames = make([]string, 0, len(ifces))
				for _, ifce := range ifces {
					if ifce.Flags&net.FlagUp == net.FlagUp {
						ifnames = append(ifnames, ifce.Name)
					}
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
				log.Trace("[DnsSupervisor] preparing whitelist")
				wlDms, err := infra.GetWhitelistCn(nil)
				//var wlDms = new(strmatcher.MatcherGroup)
				if err != nil {
					log.Warn("StartDNSSupervisorConroutine: %v", err)
					return
				}
				ipMatcher := new(router.GeoIPMatcher)
				_ = ipMatcher.Init(whiteDnsServerIps)
				for _, ifname := range ifnames {
					if _, ok := mHandles[ifname]; !ok {
						err = poison.Prepare(ifname)
						if err != nil {
							log.Warn("StartDNSSupervisorConroutine[%v]: %v", ifname, err)
							return
						}
						go func(ifname string) {
							wg.Add(1)
							defer wg.Done()
							err = poison.Run(ifname, ipMatcher, wlDms)
							if err != nil {
								log.Warn("StartDNSSupervisorConroutine[%v]: %v", ifname, err)
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
