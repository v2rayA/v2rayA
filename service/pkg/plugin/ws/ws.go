package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Ws is a base Ws struct
type Ws struct {
	dialer   plugin.Dialer
	addr     string
	wsAddr   string
	host     string
	path     string
	header   http.Header
	wsDialer *websocket.Dialer
}

func init() {
	log.Trace("[ws] registering dialer")
	plugin.RegisterDialer("ws", NewWsDialer)
}

// NewWs returns a Ws infra.
func NewWs(s string, d plugin.Dialer) (*Ws, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("NewWs: %w", err)
	}

	t := &Ws{
		dialer: d,
		addr:   u.Host,
	}

	query := u.Query()
	t.host = query.Get("host")
	if t.host == "" {
		t.host = strings.Split(t.addr, ":")[0]
	}
	t.header = http.Header{}
	t.header.Set("Host", t.host)

	t.path = query.Get("path")
	if t.path == "" {
		t.path = "/"
	}
	t.wsAddr = u.Scheme + "://" + t.addr + t.path
	t.wsDialer = &websocket.Dialer{
		NetDial:      d.Dial,
		Subprotocols: []string{"binary"},
	}
	return t, nil
}

func NewWsDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	return NewWs(s, d)
}

// Addr returns forwarder's address.
func (s *Ws) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the infra.
func (s *Ws) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Ws) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[ws] dialing %s via %s (host=%s path=%s)", addr, s.wsAddr, s.host, s.path)
	s.wsDialer.NetDialContext = s.dialer.DialContext
	rc, _, err := s.wsDialer.DialContext(ctx, s.wsAddr, s.header)
	if err != nil {
		return nil, fmt.Errorf("[Ws]: dial to %s: %w", s.wsAddr, err)
	}
	return newConn(rc), err
}

// DialUDP connects to the given address via the infra.
func (s *Ws) DialUDP(network string) (pc plugin.FakeNetPacketConn, err error) {
	//TODO
	return nil, nil
}
