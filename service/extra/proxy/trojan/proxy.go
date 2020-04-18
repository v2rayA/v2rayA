package trojan

import (
	"V2RayA/extra/proxy"
	"net"
)

type Proxy struct {
	Trojan Trojan
}

func NewProxy(s string) (p Proxy, err error) {
	d, _ := proxy.NewDirect("")
	trojan, err := NewTrojan(s, d, nil)
	if err != nil {
		return
	}
	return Proxy{Trojan: *trojan}, nil
}

// Dial connects to the given address via the proxy.
func (p Proxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.Trojan.Dial(network, addr)
	return c, p.Trojan.Addr(), err
}

// DialUDP connects to the given address via the proxy.
func (p Proxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.Trojan.DialUDP(network, addr)
}

func (p Proxy) NextDialer(dstAddr string) proxy.Dialer {
	n, _ := proxy.NewDirect("")
	return n
}
