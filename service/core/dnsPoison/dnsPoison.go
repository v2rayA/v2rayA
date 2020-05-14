package dnsPoison

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
	"log"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	v2router "v2ray.com/core/app/router"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/strmatcher"
	"v2rayA/common"
	"v2rayA/common/netTools"
)

type handle struct {
	done    chan interface{}
	running bool
	*pcapgo.EthernetHandle
	inspectedWhiteDomains map[string]interface{}
	domainMutex           sync.RWMutex
}

func newHandle(ethernetHandle *pcapgo.EthernetHandle) *handle {
	return &handle{
		done:                  make(chan interface{}),
		EthernetHandle:        ethernetHandle,
		inspectedWhiteDomains: make(map[string]interface{}),
	}
}

type DnsPoison struct {
	handles map[string]*handle
	reqID   uint32
	inner   sync.Mutex
}

func New() *DnsPoison {
	return &DnsPoison{
		handles: make(map[string]*handle),
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
	d.handles[ifname] = newHandle(h)
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

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return newError("failed to set resuse_addr")
	}

	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return newError("failed to set resuse_port")
	}

	var sockaddr syscall.Sockaddr

	switch len(ip) {
	case net.IPv4len:
		a4 := &syscall.SockaddrInet4{
			Port: int(port),
		}
		copy(a4.Addr[:], ip)
		sockaddr = a4
	case net.IPv6len:
		a6 := &syscall.SockaddrInet6{
			Port: int(port),
		}
		copy(a6.Addr[:], ip)
		sockaddr = a6
	default:
		return newError("unexpected length of ip")
	}

	return syscall.Bind(int(fd), sockaddr)
}
func Fqdn(domain string) string {
	if len(domain) > 0 && domain[len(domain)-1] == '.' {
		return domain
	}
	return domain + "."
}

type handleResult int

const (
	Pass handleResult = iota
	Spoofing
	AddWhitelist
	RemoveWhitelist
)

type domainHandleResult struct {
	domain string
	result handleResult
}

func handleSendMessage(interfaceHandle *handle, m *dnsmessage.Message, sAddr, sPort, dAddr, dPort *gopacket.Endpoint, whitelistDomains *strmatcher.MatcherGroup) (result handleResult) {
	result = Pass
	// dns请求一般只有一个question
	q := m.Questions[0]
	if (q.Type != dnsmessage.TypeA && q.Type != dnsmessage.TypeAAAA) ||
		q.Class != dnsmessage.ClassINET {
		return
	}
	dm := q.Name.String()
	dmNoTQDN := strings.TrimSuffix(dm, ".")
	if dm == "" {
		return
	} else if index := whitelistDomains.Match(dmNoTQDN); index > 0 {
		// whitelistDomains
		return
	} else if index := strings.Index(dmNoTQDN, "."); index <= 0 {
		// 跳过随机的顶级域名
		return
	}
	if q.Type == dnsmessage.TypeA {
		//在已探测白名单中的放行
		interfaceHandle.domainMutex.RLock()
		if _, ok := interfaceHandle.inspectedWhiteDomains[q.Name.String()]; ok {
			interfaceHandle.domainMutex.RUnlock()
			return
		}
		interfaceHandle.domainMutex.RUnlock()
	}
	//TODO: 不支持IPv6，AAAA投毒返回空，加速解析
	go poison(m, dAddr, dPort, sAddr, sPort)
	return Spoofing
}

func handleReceiveMessage(interfaceHandle *handle, m *dnsmessage.Message) (results []*domainHandleResult, msg string) {
	//探测该域名是否被污染为本地回环
	spoofed := false
	emptyRecord := true
	var msgs []string
	q := m.Questions[0]
	dm := q.Name.String()
	dms := []string{dm}
	//CNAME has multiple answers including an AResource
	for _, a := range m.Answers {
		switch a := a.Body.(type) {
		case *dnsmessage.CNAMEResource:
			cname := a.CNAME.String()
			msgs = append(msgs, "CNAME:"+strings.TrimSuffix(cname, "."))
			dms = append(dms, cname)
		case *dnsmessage.AResource:
			msgs = append(msgs, "A:"+fmt.Sprintf("%v.%v.%v.%v", a.A[0], a.A[1], a.A[2], a.A[3]))
			if netTools.IsIntranet4(a.A) {
				spoofed = true
			}
			emptyRecord = false
		}
	}
	if emptyRecord {
		//空记录不能影响白名单探测
		return nil, msg
	}
	defer func() {
		if results != nil {
			msg = "[" + strings.Join(msgs, ", ") + "]"
		}
	}()
	todolist := make([]string, 0, len(dms))
	//读写分离，减少锁竞争
	interfaceHandle.domainMutex.RLock()
	iSpoofed := common.BoolToInt(spoofed)
	for _, d := range dms {
		if _, ok := interfaceHandle.inspectedWhiteDomains[d]; common.BoolToInt(ok)^iSpoofed == 0 {
			todolist = append(todolist, d)
		}
	}
	interfaceHandle.domainMutex.RUnlock()
	if len(todolist) > 0 {
		interfaceHandle.domainMutex.Lock()
		for _, d := range todolist {
			if _, ok := interfaceHandle.inspectedWhiteDomains[d]; common.BoolToInt(ok)^iSpoofed == 0 {
				if ok {
					delete(interfaceHandle.inspectedWhiteDomains, d)
					results = append(results, &domainHandleResult{domain: d, result: RemoveWhitelist})
				} else {
					interfaceHandle.inspectedWhiteDomains[d] = nil
					results = append(results, &domainHandleResult{domain: d, result: AddWhitelist})
				}
			}
		}
		interfaceHandle.domainMutex.Unlock()
		return results, msg
	}
	return nil, msg
}

func packetFilter(pPacket *gopacket.Packet, whitelistDnsServers *v2router.GeoIPMatcher) (m *dnsmessage.Message, pSAddr, pSPort, pDAddr, pDPort *gopacket.Endpoint) {
	packet := *pPacket
	trans := packet.TransportLayer()
	//跳过非传输层的包
	if trans == nil {
		return
	}
	transflow := trans.TransportFlow()
	sPort, dPort := transflow.Endpoints()
	//跳过非常规DNS端口53端口的包
	if dPort.String() != "53" && sPort.String() != "53" {
		return
	}
	sAddr, dAddr := packet.NetworkLayer().NetworkFlow().Endpoints()
	// TODO: 暂不支持IPv6
	sIp := net.ParseIP(sAddr.String()).To4()
	if len(sIp) != net.IPv4len {
		return
	}
	// Domain-Name-Server whitelistIps
	if ok := whitelistDnsServers.Match(sIp); ok {
		return
	}
	dIp := net.ParseIP(dAddr.String()).To4()
	if len(dIp) != net.IPv4len {
		return
	}
	// Domain-Name-Server whitelistIps
	if ok := whitelistDnsServers.Match(dIp); ok {
		return
	}
	//尝试解析为dnsmessage格式
	var dmessage dnsmessage.Message
	err := dmessage.Unpack(trans.LayerPayload())
	if err != nil {
		return
	}
	return &dmessage, &sAddr, &sPort, &dAddr, &dPort
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
				time.Sleep(1 * time.Second)
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
		m, sAddr, sPort, dAddr, dPort := packetFilter(&packet, whitelistDnsServers)
		if m == nil {
			continue
		}
		// dns请求一般只有一个question
		dm := m.Questions[0].Name.String()
		if !m.Response {
			result := handleSendMessage(handle, m, sAddr, sPort, dAddr, dPort, whitelistDomains)
			// TODO: 不显示AAAA的投毒，因为暂时不支持IPv6
			if result == Spoofing && m.Questions[0].Type == dnsmessage.TypeA {
				log.Println("dnsPoison["+ifname+"]:", sAddr.String()+":"+sPort.String(), "->", dAddr.String()+":"+dPort.String(), dm, "已投毒")
			}
		} else {
			results, msg := handleReceiveMessage(handle, m)
			if results != nil {
				log.Println("dnsPoison["+ifname+"]:", dAddr.String()+":"+dPort.String(), "<-", sAddr.String()+":"+sPort.String(), "("+dm, "=>", msg+")")
				for _, r := range results {
					switch r.result {
					case AddWhitelist:
						log.Println("dnsPoison["+ifname+"]: 探测到", r.domain, "加入白名单", dm+msg)
					case RemoveWhitelist:
						log.Println("dnsPoison["+ifname+"]: 探测到", r.domain, "从白名单移除", dm+msg)
					}
				}
			}
		}
	}
	return
}

func poison(m *dnsmessage.Message, lAddr, lPort, rAddr, rPort *gopacket.Endpoint) {
	q := m.Questions[0]
	m.RCode = dnsmessage.RCodeSuccess
	switch q.Type {
	case dnsmessage.TypeAAAA:
		//返回空回答
	case dnsmessage.TypeA:
		//对A查询返回一个公网地址以使得后续tcp连接经过网关嗅探，以dns污染解决dns污染
		m.Answers = []dnsmessage.Resource{{
			Header: dnsmessage.ResourceHeader{
				Name:  q.Name,
				Type:  q.Type,
				Class: q.Class,
				TTL:   0,
			},
			Body: &dnsmessage.AResource{A: [4]byte{1, 2, 3, 4}},
		}}
	}
	m.Response = true
	m.RecursionAvailable = true
	packed, _ := m.Pack()
	// write back
	lport, _ := strconv.Atoi(lPort.String())
	conn, err := newDialer(lAddr.String(), uint32(lport), 30*time.Second).Dial("udp", rAddr.String()+":"+rPort.String())
	if err != nil {
		return
	}
	defer conn.Close()
	_, _ = conn.Write(packed)
}
