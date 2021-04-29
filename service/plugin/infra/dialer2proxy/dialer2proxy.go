package dialer2proxy

import (
	"github.com/v2rayA/v2rayA/plugin/infra"
	"net"
)

type Proxy struct {
	Dialer infra.Dialer
	Target string
}

func From(dialer infra.Dialer, target string) (p Proxy) {
	return Proxy{Dialer: dialer, Target: target}
}

// Dial connects to the given address via the infra.
func (p Proxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.Dialer.Dial(network, addr)
	return c, p.Dialer.Addr(), err
}

// DialUDP connects to the given address via the infra.
func (p Proxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.Dialer.DialUDP(network, addr)
}

func (p Proxy) NextDialer(dstAddr string) infra.Dialer {
	n, _ := infra.NewDirect("")
	return n
}
