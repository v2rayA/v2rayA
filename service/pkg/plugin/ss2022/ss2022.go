package ss2022

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	shadowsocks "github.com/sagernet/sing-shadowsocks"
	"github.com/sagernet/sing-shadowsocks/shadowaead_2022"
	M "github.com/sagernet/sing/common/metadata"

	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type Shadowsocks2022 struct {
	upstream plugin.Dialer
	method   shadowsocks.Method
	server   string // "host:port"
}

func init() {
	log.Trace("[shadowsocks2022] registering dialer")
	plugin.RegisterDialer("ss2022", NewShadowsocks2022Dialer)
}

func NewShadowsocks2022Dialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parse ss2022 url %q: %w", s, err)
	}
	if u.User == nil {
		return nil, fmt.Errorf("ss2022 url %q missing userinfo", s)
	}
	method := u.User.Username()
	password, hasPassword := u.User.Password()
	if method == "" || !hasPassword {
		return nil, fmt.Errorf("ss2022 url %q missing method or psk", s)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("ss2022 url port %q: %w", u.Port(), err)
	}
	m, err := shadowaead_2022.NewWithPassword(method, password, nil)
	if err != nil {
		return nil, fmt.Errorf("ss2022 new method %q: %w", method, err)
	}
	return &Shadowsocks2022{
		upstream: d,
		method:   m,
		server:   net.JoinHostPort(u.Hostname(), strconv.Itoa(port)),
	}, nil
}

func (s *Shadowsocks2022) Addr() string { return s.server }

func (s *Shadowsocks2022) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Shadowsocks2022) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	rc, err := s.upstream.DialContext(ctx, "tcp", s.server)
	if err != nil {
		return nil, fmt.Errorf("[ss2022]: dial server %s: %w", s.server, err)
	}
	wrapped, err := s.method.DialConn(rc, M.ParseSocksaddr(addr))
	if err != nil {
		_ = rc.Close()
		return nil, fmt.Errorf("[ss2022]: wrap conn: %w", err)
	}
	return wrapped, nil
}

func (s *Shadowsocks2022) DialUDP(network string) (plugin.FakeNetPacketConn, error) {
	return nil, fmt.Errorf("[ss2022]: UDP is not yet supported")
}
