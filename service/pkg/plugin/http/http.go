package http

import (
	"context"
	"fmt"
	"net"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/outbound/dialer/http"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"

	_ "github.com/daeuniverse/outbound/protocol/http"
)

// HTTP is a base http struct
type HTTP struct {
	dialer netproxy.Dialer
}

func init() {
	log.Trace("[http] registering dialer")
	plugin.RegisterDialer("http", NewHTTPDialer)
	plugin.RegisterDialer("https", NewHTTPDialer)
}

func NewHTTPDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {

	dialer, _, err := http.NewHTTP(
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
	return &HTTP{
		dialer: dialer,
	}, nil
}

// Addr returns forwarder's address.
func (s *HTTP) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *HTTP) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *HTTP) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	rc, err := s.dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[http]: dial to %s: %w", addr, err)
	}
	return plugin.NewFakeNetConn(rc), nil
}

// DialUDP connects to the given address via the infra.
func (s *HTTP) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	return nil, fmt.Errorf("[http] udp not supported")
}
