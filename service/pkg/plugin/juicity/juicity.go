package juicity

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/juicity"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/juicity"
)

// Juicity is a base juicity struct
type Juicity struct {
	dialer netproxy.Dialer
}

func init() {
	log.Trace("[juicity] registering dialer")
	plugin.RegisterDialer("juicity", NewJuicityDialer)
}

func NewJuicityDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {

	dialer, _, err := juicity.NewJuicity(
		&dialer.ExtraOption{},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &Juicity{
		dialer: dialer,
	}, nil
}

// Addr returns forwarder's address.
func (s *Juicity) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *Juicity) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Juicity) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[%s] dialing %s", "juicity", addr)
	magicNetwork := netproxy.MagicNetwork{
		Network: "tcp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		log.Info("[%s] dial %s failed: %v", "juicity", addr, err)
		return nil, fmt.Errorf("[juicity]: dial to %s: %w", addr, err)
	}
	log.Info("[%s] dial %s success", "juicity", addr)
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the infra.
func (s *Juicity) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	log.Info("[%s] dialing udp", "juicity")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, fmt.Errorf("[juicity]: dial udp %w", err)
	}
	return plugin.NewFakeNetPacketConn(rc.(netproxy.PacketConn)), nil
}
