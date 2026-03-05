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
	Dialer   Dialer
	NodeName string
	Protocol string
}

func Dialer2Proxy(dialer Dialer, nodeName string, protocol string) (p Proxy) {
	return DirectProxy{
		Dialer:   dialer,
		NodeName: nodeName,
		Protocol: protocol,
	}
}

// Dial connects to the given address via the infra.
func (p DirectProxy) Dial(network, addr string) (net.Conn, string, error) {
	return p.DialContext(context.Background(), network, addr)
}

func (p DirectProxy) DialContext(ctx context.Context, network, addr string) (net.Conn, string, error) {
	nodeInfo := ""
	if p.NodeName != "" {
		nodeInfo = "[" + p.NodeName + "]"
	}
	protocolInfo := ""
	if p.Protocol != "" {
		protocolInfo = "[" + p.Protocol + "]"
	}
	prefix := nodeInfo + protocolInfo
	if prefix == "" {
		prefix = "[proxy]"
	}

	c, err := p.Dialer.DialContext(ctx, network, addr)
	if err != nil {
		log.Info("%s dial %s failed: %v", prefix, addr, err)
	} else {
		log.Info("%s dial %s success", prefix, addr)
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
