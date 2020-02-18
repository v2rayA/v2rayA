package service

import (
	"V2RayA/model/dnsPoison"
	"V2RayA/model/v2ray/asset"
	"V2RayA/tools/netTools"
	"errors"
	"log"
	"sync"
	"time"
)

var poison = dnsPoison.New()
var poisonDone chan interface{}
var poisonMutex sync.Mutex
var limit = make(chan interface{}, 1)

func StartDNSPoison() (err error) {
	defer func() {
		if err != nil {
			err = errors.New("StartDNSPoison: " + err.Error())
		}
	}()
	poisonMutex.Lock()
	if poisonDone != nil {
		select {
		case <-poisonDone:
			//poisonDone has closed
		default:
			poisonMutex.Unlock()
			return errors.New("DNSPoison正在运行")
		}
	}
	poisonDone = make(chan interface{})
	poisonMutex.Unlock()
	go func(poison *dnsPoison.DnsPoison) {
		//并发限制1
		limit <- nil
		defer func() { <-limit }()
	out:
		for {
			f := func() {
				ifname, err := netTools.GetDefaultInterface()
				if err != nil {
					return
				}
				if poison.Exists(ifname) {
					return
				}
				poison.Clear()
				err = poison.Prepare(ifname)
				if err != nil {
					log.Println("StartDNSPoisonConroutine:", err)
					return
				}
				//准备白名单
				wlIps, wlDms, err := asset.GetWhitelistCn()
				if err != nil {
					log.Println("StartDNSPoisonConroutine:", err)
					return
				}
				go func() {
					err = poison.Run(ifname, wlIps, wlDms)
					if err != nil {
						log.Println("StartDNSPoisonConroutine:", err)
					}
				}()
			}
			f()
			select {
			case <-poisonDone:
				poison.Clear()
				break out
			default:
				log.Println("sleep")
				time.Sleep(5 * time.Second)
			}
		}
	}(poison)
	return nil
}

func StopDNSPoison() {
	poisonMutex.Lock()
	defer poisonMutex.Unlock()
	if poisonDone != nil {
		select {
		case <-poisonDone:
			//poisonDone has closed
		default:
			close(poisonDone)
		}
	}
}
