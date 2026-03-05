package anytls

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/anytls"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/anytls"
)

type AnyTLS struct {
	dialer netproxy.Dialer
}

func init() {
	log.Trace("[anytls] registering dialer")
	plugin.RegisterDialer("anytls", NewAnytlsDialer)
}

func NewAnytlsDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	dialer, _, err := anytls.NewAnytls(
		&dialer.ExtraOption{},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &AnyTLS{
		dialer: dialer,
	}, nil
}

func (a *AnyTLS) Addr() string {
	return ""
}

func (a *AnyTLS) Dial(network, address string) (net.Conn, error) {
	return a.DialContext(context.Background(), network, address)
}

func (a *AnyTLS) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	magicNetwork := netproxy.MagicNetwork{
		Network: network,
		Mark:    plugin.ShouldSetMark(),
	}
	conn, err := a.dialer.DialContext(ctx, magicNetwork.Encode(), address)
	if err != nil {
		return nil, err
	}
	return plugin.NewFakeNetConn(conn), nil
}

func (a *AnyTLS) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	log.Info("[%s] dialing udp", "anytls")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := a.dialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, fmt.Errorf("[anytls]: dial udp %w", err)
	}
	return plugin.NewFakeNetPacketConn(rc.(netproxy.PacketConn)), nil
}
