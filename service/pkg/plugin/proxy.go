package plugin

import (
	"net"
)

// Proxy is a dialer manager
type Proxy interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, proxy string, err error)

	// DialUDP connects to the given address via the proxy.
	DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error)

	// Get the dialer by dstAddr
	NextDialer(dstAddr string) Dialer
}

type DirectProxy struct {
	Dialer Dialer
}

func Dialer2Proxy(dialer Dialer) (p Proxy) {
	return DirectProxy{Dialer: dialer}
}

// Dial connects to the given address via the infra.
func (p DirectProxy) Dial(network, addr string) (net.Conn, string, error) {
	c, err := p.Dialer.Dial(network, addr)
	return c, p.Dialer.Addr(), err
}

// DialUDP connects to the given address via the infra.
func (p DirectProxy) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	return p.Dialer.DialUDP(network, addr)
}

func (p DirectProxy) NextDialer(dstAddr string) Dialer {
	return &Direct{}
}
