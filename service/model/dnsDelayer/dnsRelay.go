package dnsDelayer

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type handle struct {
	done    chan interface{}
	running bool
	*pcapgo.EthernetHandle
}

func newHandle(ethernetHandle *pcapgo.EthernetHandle) handle {
	return handle{
		done:           make(chan interface{}),
		EthernetHandle: ethernetHandle,
	}
}

type DnsRelay struct {
	handles map[string]handle
	reqID   uint32
	sync.Mutex
}

func New() *DnsRelay {
	return &DnsRelay{
		handles: make(map[string]handle),
	}
}

func (d *DnsRelay) Prepare(ifname string) (err error) {
	h, err := pcapgo.NewEthernetHandle(ifname)
	if err != nil {
		return
	}
	d.Lock()
	defer d.Unlock()
	d.handles[ifname] = newHandle(h)
	return
}

func (d *DnsRelay) ListHandles() (handles []string) {
	d.Lock()
	defer d.Unlock()
	for ifname := range d.handles {
		handles = append(handles, ifname)
	}
	return
}

func (d *DnsRelay) DeleteHandles(ifname string) (err error) {
	d.Lock()
	defer d.Unlock()
	_, ok := d.handles[ifname]
	if !ok {
		return errors.New("handle not exists")
	}
	close(d.handles[ifname].done)
	return
}

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return errors.New("failed to set resuse_addr")
	}

	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return errors.New("failed to set resuse_port")
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
		return errors.New("unexpected length of ip")
	}

	return syscall.Bind(int(fd), sockaddr)
}

func (d *DnsRelay) Run(ifname string) (err error) {
	d.Lock()
	handle, ok := d.handles[ifname]
	if !ok {
		return errors.New(ifname + " not exsits")
	}
	if handle.running {
		return errors.New(ifname + " is running")
	}
	handle.running = true
	d.Unlock()
	pkgsrc := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	for packet := range pkgsrc.Packets() {
		trans := packet.TransportLayer()
		if trans == nil || trans.TransportFlow().Dst().String() != "53" {
			continue
		}
		sPort, dPort := trans.TransportFlow().Endpoints()
		sAddr, dAddr := packet.NetworkLayer().NetworkFlow().Endpoints()
		//FIXME: DNSPOD在白名单中，如果用户的dns就是DNSPOD，则DNS转发失效
		if len(net.ParseIP(dAddr.String()).To4()) != net.IPv4len ||
			//dAddr.String() == "119.29.29.29" ||
			dAddr.String() == "172.20.10.1" {
			log.Println(dAddr)
			continue
		}
		var m dnsmessage.Message
		err := m.Unpack(trans.LayerPayload())
		if err != nil {
			continue
		}

		log.Println(sAddr.String()+":"+sPort.String(), "->", dAddr.String()+":"+dPort.String(), m.Questions)
		go func(m *dnsmessage.Message, sAddr, sPort, dAddr, dPort *gopacket.Endpoint) {
			// change the protocol
			dConn, err := net.Dial("tcp", dAddr.String()+":"+dPort.String())
			if err != nil {
				log.Println(err)
				return
			}
			defer dConn.Close()
			//var option IPOption
			//switch m.Questions[0].Type {
			//case dnsmessage.TypeA:
			//	option = IPOption{
			//		IPv4Enable: true,
			//		IPv6Enable: false,
			//	}
			//case dnsmessage.TypeAAAA:
			//	option = IPOption{
			//		IPv4Enable: false,
			//		IPv6Enable: true,
			//	}
			//}
			//local, _, _ := net.SplitHostPort(dConn.LocalAddr().String())
			//reqs := buildReqMsgs(m.Questions[0].Name.String(), option, func() uint16 {
			//	return uint16(atomic.AddUint32(&d.reqID, 1))
			//}, genEDNS0Options(net.ParseIP(local)))
			//b, _ := dns.PackMessage(reqs[0].msg)
			log.Println(m.Header, m.Response)
			packed, _ := m.Pack()
			m.ID = uint16(len(packed))
			packed, _ = m.Pack()
			_, err = dConn.Write(packed)
			//_, err = dConn.Write(b.Bytes())
			if err != nil {
				log.Println(err)
				return
			}
			var buf [512]byte
			n, err := dConn.Read(buf[:])
			if err != nil {
				log.Println(err)
				return
			}
			m.Unpack(buf[:n])
			log.Println(m.Answers)
			// write back
			dport, _ := strconv.Atoi(dPort.String())
			sConn, err := newDialer(dAddr.String(), uint32(dport), 30*time.Second).Dial("udp", sAddr.String()+":"+sPort.String())
			if err != nil {
				log.Println(err)
				return
			}
			defer sConn.Close()
			_, _ = sConn.Write(buf[:n])
		}(&m, &sAddr, &sPort, &dAddr, &dPort)
		select {
		case <-handle.done:
			d.Lock()
			delete(d.handles, ifname)
			d.Unlock()
			break
		default:
		}
	}
	return
}
