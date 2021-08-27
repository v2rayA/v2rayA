package ss

import (
	"github.com/v2rayA/v2rayA/pkg/plugin/infra"
	"net"
)

type Proxy struct {
	Shadowsocks Shadowsocks
}

func NewProxy(s string) (p Proxy, err error) {
	d, _ := infra.NewDirect("")
	shadowsocks, err := NewShadowsocks(s, d)
	if err != nil {
		return
	}
	return Proxy{Shadowsocks: *shadowsocks}, nil
}

// Dial connects to the given address via the infra.
func (p Proxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.Shadowsocks.Dial(network, addr)
	return c, p.Shadowsocks.Addr(), err
}

// DialUDP connects to the given address via the infra.
func (p Proxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.Shadowsocks.DialUDP(network, addr)
}

func (p Proxy) NextDialer(dstAddr string) infra.Dialer {
	n, _ := infra.NewDirect("")
	return n
}
