package plugin

import (
	"context"
	"net"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Proxy is a dialer manager
type Proxy interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, proxy string, err error)

	DialContext(ctx context.Context, network, addr string) (c net.Conn, proxy string, err error)

	// DialUDP connects to the given address via the proxy.
	DialUDP(network string) (pc FakeNetPacketConn, proxy string, err error)

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
	return p.DialContext(context.Background(), network, addr)
}

func (p DirectProxy) DialContext(ctx context.Context, network, addr string) (net.Conn, string, error) {
	log.Info("[proxy] dialing %s via %s", addr, p.Dialer.Addr())
	c, err := p.Dialer.DialContext(ctx, network, addr)
	if err != nil {
		log.Info("[proxy] dial %s via %s failed: %v", addr, p.Dialer.Addr(), err)
	} else {
		log.Info("[proxy] dial %s via %s success", addr, p.Dialer.Addr())
	}
	return c, p.Dialer.Addr(), err
}

// DialUDP connects to the given address via the infra.
func (p DirectProxy) DialUDP(network string) (pc FakeNetPacketConn, proxy string, err error) {
	pc, err = p.Dialer.DialUDP(network)
	return pc, p.Dialer.Addr(), err
}

func (p DirectProxy) NextDialer(dstAddr string) Dialer {
	return &Direct{}
}
