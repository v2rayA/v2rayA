package ssr

import (
	"github.com/mzz2017/v2rayA/extra/proxy"
	"net"
)

type Proxy struct {
	SSR SSR
}

func NewProxy(s string) (p Proxy, err error) {
	d, _ := proxy.NewDirect("")
	ssr, err := NewSSR(s, d)
	if err != nil {
		return
	}
	return Proxy{SSR: *ssr}, nil
}

// Dial connects to the given address via the proxy.
func (p Proxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.SSR.Dial(network, addr)
	return c, p.SSR.Addr(), err
}

// DialUDP connects to the given address via the proxy.
func (p Proxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.SSR.DialUDP(network, addr)
}

func (p Proxy) NextDialer(dstAddr string) proxy.Dialer {
	n, _ := proxy.NewDirect("")
	return n
}
