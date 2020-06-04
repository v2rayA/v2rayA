package dnsPoison

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"log"
	"sync"
	"time"
	v2router "v2ray.com/core/app/router"
	"v2ray.com/core/common/strmatcher"
)

type DnsPoison struct {
	handles        map[string]*handle
	reqID          uint32
	inner          sync.Mutex
	reservedIpPool *ReservedIpPool
}

func New() *DnsPoison {
	return &DnsPoison{
		handles:        make(map[string]*handle),
		reservedIpPool: NewReservedIpPool(),
	}
}

func (d *DnsPoison) Exists(ifname string) bool {
	_, ok := d.handles[ifname]
	return ok
}

func (d *DnsPoison) Clear() {
	handles := d.ListHandles()
	for _, h := range handles {
		_ = d.DeleteHandles(h)
	}
	log.Println("DnsPoison: Clear")
}

func (d *DnsPoison) Prepare(ifname string) (err error) {
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

func (d *DnsPoison) ListHandles() (ifnames []string) {
	d.inner.Lock()
	defer d.inner.Unlock()
	for ifname := range d.handles {
		ifnames = append(ifnames, ifname)
	}
	return
}

func (d *DnsPoison) DeleteHandles(ifname string) (err error) {
	d.inner.Lock()
	defer d.inner.Unlock()
	if !d.Exists(ifname) {
		return newError("handle not exists")
	}
	close(d.handles[ifname].done)
	delete(d.handles, ifname)
	log.Println("DnsPoison:", ifname, "closed")
	return
}

func (d *DnsPoison) Run(ifname string, whitelistDnsServers *v2router.GeoIPMatcher, whitelistDomains *strmatcher.MatcherGroup) (err error) {
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
	log.Println("DnsPoison[" + ifname + "]: running")
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
