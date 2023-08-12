package plugin

import (
	"fmt"

	"github.com/daeuniverse/softwind/netproxy"
)

type Converter struct {
	Dialer
}

// Dial implements netproxy.Dialer.
func (c *Converter) Dial(network string, addr string) (netproxy.Conn, error) {
	magic, err := netproxy.ParseMagicNetwork(network)
	if err != nil {
		return nil, err
	}
	switch magic.Network {
	case "tcp":
		rc, err := c.Dialer.Dial(magic.Network, addr)
		if err != nil {
			return nil, err
		}
		return rc, nil
	case "udp":
		pc, _, err := c.Dialer.DialUDP(magic.Network, addr)
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

var _ netproxy.Dialer = &Converter{}
