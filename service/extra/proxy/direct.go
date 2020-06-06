package proxy

import (
	"log"
	"net"
	"syscall"
	"github.com/mzz2017/v2rayA/global"
)

// Direct proxy
type Direct struct {
	iface *net.Interface // interface specified by user
	ip    net.IP
}

// Default dialer
var Default = &Direct{}

// NewDirect returns a Direct dialer
func NewDirect(intface string) (*Direct, error) {
	if intface == "" {
		return &Direct{}, nil
	}

	ip := net.ParseIP(intface)
	if ip != nil {
		return &Direct{ip: ip}, nil
	}

	iface, err := net.InterfaceByName(intface)
	if err != nil {
		return nil, newError(intface).Base(err)
	}

	return &Direct{iface: iface}, nil
}

// Addr returns forwarder's address
func (d *Direct) Addr() string { return "DIRECT" }

// Dial connects to the address addr on the network net
func (d *Direct) Dial(network, addr string) (c net.Conn, err error) {
	if d.iface == nil || d.ip != nil {
		c, err = dial(network, addr, d.ip)
		if err == nil {
			return
		}
	}

	for _, ip := range d.IFaceIPs() {
		c, err = dial(network, addr, ip)
		if err == nil {
			d.ip = ip
			break
		}
	}

	// no ip available (so no dials made), maybe the interface link is down
	if c == nil && err == nil {
		err = newError("dial failed, maybe the interface link is down, please check it")
	}

	return c, err
}

func dial(network, addr string, localIP net.IP) (net.Conn, error) {
	if network == "uot" {
		network = "udp"
	}

	var la net.Addr
	switch network {
	case "tcp":
		la = &net.TCPAddr{IP: localIP}
	case "udp":
		la = &net.UDPAddr{IP: localIP}
	}

	dialer := &net.Dialer{
		LocalAddr: la,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				//TODO: force to set 0xff. any chances to customize this value?
				err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, 0xff)
				if err != nil {
					if global.IsDebug() {
						log.Printf("control: %s", err)
					}
					return
				}
			})
		},
	}
	c, err := dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if c, ok := c.(*net.TCPConn); ok {
		c.SetKeepAlive(true)
	}

	return c, err
}

// DialUDP connects to the given address
func (d *Direct) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	// TODO: support specifying local interface
	la := ""
	if d.ip != nil {
		la = d.ip.String() + ":0"
	}

	pc, err := net.ListenPacket(network, la)
	if err != nil {
		log.Printf("ListenPacket error: %s\n", err)
		return nil, nil, err
	}

	uAddr, err := net.ResolveUDPAddr("udp", addr)
	return pc, uAddr, err
}

// IFaceIPs returns ip addresses according to the specified interface
func (d *Direct) IFaceIPs() (ips []net.IP) {
	ipnets, err := d.iface.Addrs()
	if err != nil {
		return
	}

	for _, ipnet := range ipnets {
		ips = append(ips, ipnet.(*net.IPNet).IP) //!ip.IsLinkLocalUnicast()
	}

	return
}
