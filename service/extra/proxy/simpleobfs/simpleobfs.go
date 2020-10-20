package simpleobfs

import "C"
import (
	"fmt"
	"github.com/v2rayA/v2rayA/extra/proxy"
	"log"
	"net"
	"net/url"
	"strings"
)

type ObfsType int

const (
	HTTP ObfsType = iota
	TLS
)

func init() {
	proxy.RegisterDialer("simpleobfs", NewSimpleObfsDialer)
}

// NewSimpleObfsDialer returns a simple-obfs proxy dialer.
func NewSimpleObfsDialer(s string, d proxy.Dialer) (proxy.Dialer, error) {
	return NewSimpleObfs(s, d)
}

// SimpleObfs is a base http-obfs struct
type SimpleObfs struct {
	dialer   proxy.Dialer
	obfstype ObfsType
	addr     string
	path     string
	host     string
}

// NewSimpleobfs returns a simpleobfs proxy.
func NewSimpleObfs(s string, d proxy.Dialer) (*SimpleObfs, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, newError("[simpleobfs]").Base(err)
	}

	t := &SimpleObfs{
		dialer: d,
		addr:   u.Host,
	}
	query := u.Query()
	obfstype := query.Get("type")
	switch strings.ToLower(obfstype) {
	case "http":
		t.obfstype = HTTP
	case "tls":
		t.obfstype = TLS
	default:
		return nil, newError("unsupported obfs type ", obfstype)
	}
	t.host = query.Get("host")
	t.path = query.Get("path")
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
	return s.dial(network, addr)
}

func (s *SimpleObfs) dial(network, addr string) (c net.Conn, err error) {
	rc, err := s.dialer.Dial("tcp", s.addr)
	log.Println("dial", s.addr, s.obfstype, s.host, s.path)
	if err != nil {
		return nil, newError(fmt.Sprintf("[simpleobfs]: dial to %s", s.addr)).Base(err)
	}
	switch s.obfstype {
	case HTTP:
		rs := strings.Split(s.addr, ":")
		var port string
		if len(rs) == 1 {
			port = "80"
		} else {
			port = rs[1]
		}
		c = NewHTTPObfs(rc, rs[0], port, s.path)
	case TLS:
		c = NewTLSObfs(rc, s.host)
	}
	return c, err
}

// DialUDP connects to the given address via the proxy.
func (s *SimpleObfs) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	//TODO
	return nil, nil, nil
}
