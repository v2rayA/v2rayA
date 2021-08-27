package trojanc

import (
	"github.com/v2rayA/v2rayA/pkg/plugin/infra"
	"net"
)

type Proxy struct {
	Trojan Trojan
}

func NewProxy(s string) (p Proxy, err error) {
	d, _ := infra.NewDirect("")
	trojan, err := NewTrojanc(s, d)
	if err != nil {
		return
	}
	return Proxy{Trojan: *trojan}, nil
}

// Dial connects to the given address via the infra.
func (p Proxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.Trojan.Dial(network, addr)
	return c, p.Trojan.Addr(), err
}

// DialUDP connects to the given address via the infra.
func (p Proxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.Trojan.DialUDP(network, addr)
}

func (p Proxy) NextDialer(dstAddr string) infra.Dialer {
	n, _ := infra.NewDirect("")
	return n
}
