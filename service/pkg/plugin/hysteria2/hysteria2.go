package hysteria2

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/hysteria2"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Hysteria2 is a base hysteria2 struct
type Hysteria2 struct {
	dialer netproxy.Dialer
}

func init() {
	log.Trace("[hysteria2] registering dialer")
	plugin.RegisterDialer("hysteria2", NewHysteria2Dialer)
}

func NewHysteria2Dialer(s string, d plugin.Dialer) (plugin.Dialer, error) {

	dialer, _, err := hysteria2.NewHysteria2(
		&dialer.ExtraOption{},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &Hysteria2{
		dialer: dialer,
	}, nil
}

// Addr returns forwarder's address.
func (s *Hysteria2) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *Hysteria2) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Hysteria2) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[%s] dialing %s", "hysteria2", addr)
	magicNetwork := netproxy.MagicNetwork{
		Network: "tcp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		log.Info("[%s] dial %s failed: %v", "hysteria2", addr, err)
		return nil, err
	}
	log.Info("[%s] dial %s success", "hysteria2", addr)
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the infra.
func (s *Hysteria2) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	log.Info("[%s] dialing udp", "hysteria2")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, fmt.Errorf("[hysteria2]: dial udp %w", err)
	}
	return plugin.NewFakeNetPacketConn(rc.(netproxy.PacketConn)), nil
}
