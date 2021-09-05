package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	return s.dial(network, addr)
}

func (s *Ws) dial(network, addr string) (net.Conn, error) {
	rc, _, err := s.wsDialer.Dial(s.wsAddr, s.header)
	if err != nil {
		return nil, fmt.Errorf("[Ws]: dial to %s: %w", s.wsAddr, err)
	}
	return newConn(rc), err
}

// DialUDP connects to the given address via the infra.
func (s *Ws) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	//TODO
	return nil, nil, nil
}
