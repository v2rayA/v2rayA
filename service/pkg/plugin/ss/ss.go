package ss

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/shadowsocks"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/shadowsocks"
)

// Shadowsocks is a base shadowsocks struct
type Shadowsocks struct {
	dialer netproxy.Dialer
}

func init() {
	log.Trace("[shadowsocks] registering dialer")
	plugin.RegisterDialer("ss", NewShadowsocksDialer)
}

func NewShadowsocksDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	dialer, _, err := shadowsocks.NewShadowsocksFromLink(
		&dialer.ExtraOption{},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &Shadowsocks{
		dialer: dialer,
	}, nil
}

// Addr returns forwarder's address.
func (s *Shadowsocks) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *Shadowsocks) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Shadowsocks) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	magicNetwork := netproxy.MagicNetwork{
		Network: "tcp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		return nil, fmt.Errorf("[shadowsocks]: dial to %s: %w", addr, err)
	}
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the infra.
func (s *Shadowsocks) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	log.Info("[%s] dialing udp", "shadowsocks")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, fmt.Errorf("[shadowsocks]: dial udp %w", err)
	}
	return plugin.NewFakeNetPacketConn(rc.(netproxy.PacketConn)), nil
}
