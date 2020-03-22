package entity

import (
	"V2RayA/core/dnsPoison"
	"V2RayA/core/v2ray/asset"
	"V2RayA/common/netTools"
	"errors"
	"log"
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
)

type ExtraInfo struct {
	DohIps       []string
	DohDomains   []string
	ServerIps    []string
	ServerDomain string
}

func SetupDnsPoisonWithExtraInfo(info *ExtraInfo) {
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
	})
	_ = StartDNSPoison([]*router.CIDR{
		{Ip: []byte{119, 29, 29, 29}, Prefix: 32},
		{Ip: []byte{114, 114, 114, 114}, Prefix: 32},
	},
		whitedms)
}

func StartDNSPoison(externWhiteDnsServers []*router.CIDR, externWhiteDomains []*router.Domain) (err error) {
	defer func() {
		if err != nil {
			err = errors.New("StartDNSPoison: " + err.Error())
		}
	}()
	mutex.Lock()
	if done != nil {
		select {
		case <-done:
			//done has closed
		default:
			mutex.Unlock()
			return errors.New("DNSPoison is running")
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
				ifnames, err := netTools.GetDefaultInterface()
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
				_, wlDms, err := asset.GetWhitelistCn(nil, whiteDomains)
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
				time.Sleep(5 * time.Second)
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
			//done has closed
		default:
			close(done)
		}
	}
}
