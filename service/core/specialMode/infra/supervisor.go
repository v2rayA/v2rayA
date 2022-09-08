package infra

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/v2fly/v2ray-core/v5/common/strmatcher"
	v2router "github.com/v2rayA/v2ray-lib/router"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"sync"
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
	log.Trace("DnsSupervisor: Clear")
}

func (d *DnsSupervisor) Prepare(ifname string) (err error) {
	d.inner.Lock()
	defer d.inner.Unlock()
	if d.Exists(ifname) {
		return fmt.Errorf("Prepare: %v exists", ifname)
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
		return fmt.Errorf("DeleteHandles: handle not exists")
	}
	close(d.handles[ifname].done)
	delete(d.handles, ifname)
	log.Trace("DnsSupervisor:%v deleted", ifname)
	return
}

func (d *DnsSupervisor) Run(ifname string, whitelistDnsServers *v2router.GeoIPMatcher, whitelistDomains strmatcher.MatcherGroup) (err error) {
	defer func() {
		recover()
	}()
	d.inner.Lock()
	handle, ok := d.handles[ifname]
	if !ok {
		d.inner.Unlock()
		return fmt.Errorf("Run: %v not exsits", ifname)
	}
	if handle.running {
		d.inner.Unlock()
		return fmt.Errorf("Run: %v is running", ifname)
	}
	handle.running = true
	log.Trace("[DnsSupervisor] " + ifname + ": running")
	// we only decode UDP packets
	pkgsrc := gopacket.NewPacketSource(handle, layers.LinkTypeEthernet)
	pkgsrc.NoCopy = true
	//pkgsrc.Lazy = true
	d.inner.Unlock()
	packets := pkgsrc.Packets()
	go func() {
		<-handle.done
		packets <- gopacket.NewPacket(nil, layers.LinkTypeEthernet, pkgsrc.DecodeOptions)
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
	log.Trace("DnsSupervisor:%v closed", ifname)
	return
}
