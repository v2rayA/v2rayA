package simpleobfs

import (
	"github.com/v2rayA/v2rayA/extra/proxy"
	"net"
)

type Proxy struct {
	SimpleObfs SimpleObfs
}

func NewProxy(s string) (p Proxy, err error) {
	d, _ := proxy.NewDirect("")
	simpleObfs, err := NewSimpleObfs(s, d)
	if err != nil {
		return
	}
	return Proxy{SimpleObfs: *simpleObfs}, nil
}

// Dial connects to the given address via the proxy.
func (p Proxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.SimpleObfs.Dial(network, addr)
	return c, p.SimpleObfs.Addr(), err
}

// DialUDP connects to the given address via the proxy.
func (p Proxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.SimpleObfs.DialUDP(network, addr)
}

func (p Proxy) NextDialer(dstAddr string) proxy.Dialer {
	n, _ := proxy.NewDirect("")
	return n
}

