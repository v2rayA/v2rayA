package tuic

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/tuic"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/tuic"
)

// Tuic is a base tuic struct
type Tuic struct {
	dialer netproxy.Dialer
}

func init() {
	log.Info("[tuic] registering dialer")
	plugin.RegisterDialer("tuic", NewTuicDialer)
}

func NewTuicDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {

	dialer, _, err := tuic.NewTuic(
		&dialer.ExtraOption{},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &Tuic{
		dialer: dialer,
	}, nil
}

// Addr returns forwarder's address.
func (s *Tuic) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *Tuic) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Tuic) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	magicNetwork := netproxy.MagicNetwork{
		Network: "tcp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		return nil, fmt.Errorf("[tuic]: dial to %s: %w", addr, err)
	}
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the infra.
func (s *Tuic) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	log.Info("[%s] dialing udp", "tuic")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, fmt.Errorf("[tuic]: dial udp %w", err)
	}
	return plugin.NewFakeNetPacketConn(rc.(netproxy.PacketConn)), nil
}
