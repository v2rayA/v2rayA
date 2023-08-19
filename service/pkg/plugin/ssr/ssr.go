package ssr

import (
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/shadowsocksr"
	"github.com/daeuniverse/softwind/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"

	_ "github.com/daeuniverse/softwind/protocol/shadowsocks_stream"
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
	rc, err := s.dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[ssr]: dial to %s: %w", addr, err)
	}
	return &netproxy.FakeNetConn{
		Conn:  rc,
		LAddr: nil,
		RAddr: nil,
	}, err
}

// DialUDP connects to the given address via the proxy.
func (s *SSR) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	return nil, nil, fmt.Errorf("[ssr] udp not supported now")
}
