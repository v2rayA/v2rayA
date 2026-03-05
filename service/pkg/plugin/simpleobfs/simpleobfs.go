package simpleobfs

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type ObfsType int

const (
	HTTP ObfsType = iota
	TLS
)

func init() {
	log.Trace("[simpleobfs] registering dialer")
	plugin.RegisterDialer("simpleobfs", NewSimpleObfsDialer)
	plugin.RegisterDialer("simple-obfs", NewSimpleObfsDialer)
	plugin.RegisterDialer("obfs-local", NewSimpleObfsDialer)
}

// NewSimpleObfsDialer returns a simple-obfs proxy dialer.
func NewSimpleObfsDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	return NewSimpleObfs(s, d)
}

// SimpleObfs is a base http-obfs struct
type SimpleObfs struct {
	dialer   plugin.Dialer
	obfstype ObfsType
	addr     string
	path     string
	host     string
}

// NewSimpleobfs returns a simpleobfs proxy.
func NewSimpleObfs(s string, d plugin.Dialer) (*SimpleObfs, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("simpleobfs: %w", err)
	}

	t := &SimpleObfs{
		dialer: d,
		addr:   u.Host,
	}
	query := u.Query()
	obfstype := query.Get("type")
	if obfstype == "" {
		obfstype = query.Get("obfs")
	}
	switch strings.ToLower(obfstype) {
	case "http":
		t.obfstype = HTTP
	case "tls":
		t.obfstype = TLS
	default:
		return nil, fmt.Errorf("unsupported obfs type %v", obfstype)
	}
	t.host = query.Get("host")
	t.path = query.Get("path")
	if t.path == "" {
		t.path = query.Get("uri")
	}
	return t, nil
}

// Addr returns forwarder's address.
func (s *SimpleObfs) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the proxy.
func (s *SimpleObfs) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *SimpleObfs) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[simpleobfs] dialing %s via %s (type=%v host=%v path=%v)", addr, s.addr, s.obfstype, s.host, s.path)
	rc, err := s.dialer.DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return nil, fmt.Errorf("[simpleobfs]: dial to %s: %w", s.addr, err)
	}
	var c net.Conn
	switch s.obfstype {
	case HTTP:
		_, port, _ := net.SplitHostPort(s.addr)
		if port == "" {
			port = "80"
		}
		c = NewHTTPObfs(rc, s.host, port, s.path)
	case TLS:
		c = NewTLSObfs(rc, s.host)
	}
	return c, nil
}

// DialUDP connects to the given address via the proxy.
func (s *SimpleObfs) DialUDP(network string) (pc plugin.FakeNetPacketConn, err error) {
	//TODO
	return nil, nil
}
