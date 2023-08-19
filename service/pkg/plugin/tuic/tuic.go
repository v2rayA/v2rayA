package tuic

import (
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/tuic"
	"github.com/daeuniverse/softwind/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"

	_ "github.com/daeuniverse/softwind/protocol/tuic"
)

// Tuic is a base tuic struct
type Tuic struct {
	dialer netproxy.Dialer
}

func init() {
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
	return s.dial(network, addr)
}

func (s *Tuic) dial(network, addr string) (net.Conn, error) {
	rc, err := s.dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[tuic]: dial to %s: %w", addr, err)
	}
	return &netproxy.FakeNetConn{
		Conn:  rc,
		LAddr: nil,
		RAddr: nil,
	}, err
}

// DialUDP connects to the given address via the infra.
func (s *Tuic) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	rc, err := s.dialer.Dial("udp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("[tuic]: dial to %s: %w", addr, err)
	}
	return &netproxy.FakeNetPacketConn{
		PacketConn: rc.(netproxy.PacketConn),
		LAddr:      nil,
		RAddr:      nil,
	}, nil, err
}
