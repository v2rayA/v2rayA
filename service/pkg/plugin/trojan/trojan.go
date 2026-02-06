package trojan

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/trojan"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/trojanc"
)

// Trojan is a base trojan struct
type Trojan struct {
	dialer netproxy.Dialer
}

func init() {
	println("[DEBUG] trojan.init called")
	log.Trace("[trojan] registering dialer")
	plugin.RegisterDialer("trojan", NewTrojanDialer)
	plugin.RegisterDialer("trojan-go", NewTrojanDialer)
}

func NewTrojanDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {

	dialer, _, err := trojan.NewTrojan(
		&dialer.ExtraOption{
			TlsImplementation: "tls",
		},
		&plugin.Converter{
			Dialer: d,
		},
		s,
	)
	if err != nil {
		return nil, err
	}
	return &Trojan{
		dialer: dialer,
	}, nil
}

// Addr returns forwarder's address.
func (s *Trojan) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *Trojan) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Trojan) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[%s] dialing %s", "trojan", addr)
	magicNetwork := netproxy.MagicNetwork{
		Network: "tcp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr)
	if err != nil {
		log.Info("[%s] dial %s failed: %v", "trojan", addr, err)
		return nil, fmt.Errorf("[trojan]: dial to %s: %w", addr, err)
	}
	log.Info("[%s] dial %s success", "trojan", addr)
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the infra.
func (s *Trojan) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	log.Info("[%s] dialing udp", "trojan")
	magicNetwork := netproxy.MagicNetwork{
		Network: "udp",
		Mark:    plugin.ShouldSetMark(),
	}
	rc, err := s.dialer.DialContext(context.TODO(), magicNetwork.Encode(), "")
	if err != nil {
		return nil, fmt.Errorf("[trojan]: dial udp %w", err)
	}
	return plugin.NewFakeNetPacketConn(rc.(netproxy.PacketConn)), nil
}
