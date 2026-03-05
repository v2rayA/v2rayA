package ssr

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/shadowsocksr"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/shadowsocks_stream"
)

// SSR struct.
type SSR struct {
	dialer netproxy.Dialer
}

// Addr implements plugin.Dialer.
func (*SSR) Addr() string {
	return ""
}

func init() {
	log.Trace("[shadowsocksr] registering dialer")
	plugin.RegisterDialer("ssr", NewSSRDialer)
}

// NewSSRDialer returns a ssr proxy dialer.
func NewSSRDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	dialer, _, err := shadowsocksr.NewShadowsocksR(
		&dialer.ExtraOption{},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &SSR{
		dialer: dialer,
	}, nil
}

// Dial connects to the address addr on the network net via the proxy.
func (s *SSR) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *SSR) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	magicNetwork := netproxy.MagicNetwork{
		Network: "tcp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		return nil, fmt.Errorf("[ssr]: dial to %s: %w", addr, err)
	}
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the proxy.
func (s *SSR) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	return nil, fmt.Errorf("[ssr] udp not supported now")
}
