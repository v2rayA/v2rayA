package plugin

import (
	"context"
	"fmt"

	"github.com/daeuniverse/outbound/netproxy"
)

type Converter struct {
	Dialer
}

// DialContext implements netproxy.Dialer.
func (c *Converter) DialContext(ctx context.Context, network string, addr string) (netproxy.Conn, error) {
	magic, err := netproxy.ParseMagicNetwork(network)
	if err != nil {
		return nil, err
	}
	switch magic.Network {
	case "tcp":
		rc, err := c.Dialer.DialContext(ctx, magic.Network, addr)
		if err != nil {
			return nil, err
		}
		if nrc, ok := rc.(netproxy.Conn); ok {
			return nrc, nil
		}
		return &netproxy.FakeNetConn{Conn: rc, LAddr: rc.LocalAddr(), RAddr: rc.RemoteAddr()}, nil
	case "udp":
		pc, err := c.Dialer.DialUDP(magic.Network)
		if err != nil {
			return nil, err
		}
		return &PacketConnConverter{
			PacketConn: pc,
			Target:     addr,
		}, nil
	}
	return nil, fmt.Errorf("unexpected network: %v", magic.Network)
}

// Dial implements netproxy.Dialer.
func (c *Converter) Dial(network string, addr string) (netproxy.Conn, error) {
	return c.DialContext(context.TODO(), network, addr)
}

var _ netproxy.Dialer = &Converter{}
