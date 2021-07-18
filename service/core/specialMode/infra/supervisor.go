package infra

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"log"
	"sync"
	"time"
	v2router "github.com/v2fly/v2ray-core/v4/app/router"
	"github.com/v2fly/v2ray-core/v4/common/strmatcher"
)

type DnsSupervisor struct {
	handles        map[string]*handle
	reqID          uint32
	inner          sync.Mutex
	reservedIpPool *ReservedIpPool
}

func New() *DnsSupervisor {
	return &DnsSupervisor{
		handles:        make(map[string]*handle),
		reservedIpPool: NewReservedIpPool(),
	}
}

func (d *DnsSupervisor) Exists(ifname string) bool {
	_, ok := d.handles[ifname]
	return ok
}

func (d *DnsSupervisor) Clear() {
	handles := d.ListHandles()
	for _, h := range handles {
		_ = d.DeleteHandles(h)
	}
	log.Println("DnsSupervisor: Clear")
}

func (d *DnsSupervisor) Prepare(ifname string) (err error) {
	d.inner.Lock()
	defer d.inner.Unlock()
	if d.Exists(ifname) {
		return newError(ifname + " exists")
	}
	h, err := pcapgo.NewEthernetHandle(ifname)
	if err != nil {
		return
	}
	d.handles[ifname] = newHandle(d, h)
	return
}

func (d *DnsSupervisor) ListHandles() (ifnames []string) {
	d.inner.Lock()
	defer d.inner.Unlock()
	for ifname := range d.handles {
		ifnames = append(ifnames, ifname)
	}
	return
}

func (d *DnsSupervisor) DeleteHandles(ifname string) (err error) {
	d.inner.Lock()
	defer d.inner.Unlock()
	if !d.Exists(ifname) {
		return newError("handle not exists")
	}
	close(d.handles[ifname].done)
	delete(d.handles, ifname)
	log.Println("DnsSupervisor:", ifname, "closed")
	return
}

func (d *DnsSupervisor) Run(ifname string, whitelistDnsServers *v2router.GeoIPMatcher, whitelistDomains *strmatcher.MatcherGroup) (err error) {
	defer func() {
		recover()
	}()
	d.inner.Lock()
	handle, ok := d.handles[ifname]
	if !ok {
		return newError(ifname + " not exsits")
	}
	if handle.running {
		return newError(ifname + " is running")
	}
	handle.running = true
	log.Println("[DnsSupervisor] " + ifname + ": running")
	pkgsrc := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	pkgsrc.NoCopy = true
	d.inner.Unlock()
	packets := pkgsrc.Packets()
	go func() {
		for {
			//心跳包，防止内存泄漏
			packets <- gopacket.NewPacket(nil, layers.LinkTypeEthernet, gopacket.DecodeOptions{})
			select {
			case <-handle.done:
				return
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}()
out:
	for packet := range packets {
		select {
		case <-handle.done:
			break out
		default:
		}
		go handle.handlePacket(packet, ifname, whitelistDnsServers, whitelistDomains)
	}
	return
}
