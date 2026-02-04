package plugin

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strings"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type FakeNetPacketConn interface {
	net.PacketConn
}

type fakeNetPacketConn struct {
	netproxy.PacketConn
}

type fakeNetConn struct {
	netproxy.Conn
}

func (f *fakeNetConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 30000}
}

func (f *fakeNetConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 80}
}

func (f *fakeNetConn) Read(b []byte) (n int, err error) {
	n, err = f.Conn.Read(b)
	if n > 0 {
		log.Trace("[conn] read %d bytes from %v", n, f.RemoteAddr())
	}
	return
}

func (f *fakeNetConn) Write(b []byte) (n int, err error) {
	n, err = f.Conn.Write(b)
	if n > 0 {
		log.Trace("[conn] write %d bytes to %v", n, f.RemoteAddr())
	}
	return
}

func NewFakeNetConn(c netproxy.Conn) net.Conn {
	if c == nil {
		return nil
	}
	return &fakeNetConn{Conn: c}
}

func (f *fakeNetPacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{}
}

func (f *fakeNetPacketConn) RemoteAddr() net.Addr {
	return &net.UDPAddr{}
}

func (f *fakeNetPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, ap, err := f.PacketConn.ReadFrom(p)
	if err != nil {
		return 0, nil, err
	}
	return n, net.UDPAddrFromAddrPort(ap), nil
}

func (f *fakeNetPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return f.PacketConn.WriteTo(p, addr.String())
}

func NewFakeNetPacketConn(pc netproxy.PacketConn) FakeNetPacketConn {
	return &fakeNetPacketConn{PacketConn: pc}
}

// Dialer is used to create connection.
type Dialer interface {
	// Addr is the dialer's addr
	Addr() string

	// Dial connects to the given address
	Dial(network, addr string) (c net.Conn, err error)

	DialContext(ctx context.Context, network, addr string) (c net.Conn, err error)

	// DialUDP connects to the given address
	DialUDP(network string) (pc FakeNetPacketConn, err error)
}

// DialerCreator is a function to create dialers.
type DialerCreator func(s string, dialer Dialer) (Dialer, error)

var (
	dialerMap = make(map[string]DialerCreator)
)

// RegisterDialer is used to register a dialer.
func RegisterDialer(name string, c DialerCreator) {
	dialerMap[name] = c
}

// getAvailableSchemes returns a list of registered dialer schemes
func getAvailableSchemes() []string {
	schemes := make([]string, 0, len(dialerMap))
	for k := range dialerMap {
		schemes = append(schemes, k)
	}
	return schemes
}

// DialerFromURL calls the registered creator to create dialers.
// dialer is the default upstream dialer so cannot be nil, we can use Default when calling this function.
func DialerFromURL(s string, dialer Dialer) (Dialer, error) {
	if dialer == nil {
		return nil, fmt.Errorf("DialerFromURL: dialer cannot be nil")
	}

	u, err := url.Parse(s)
	if err != nil {
		log.Warn("[plugin] parse URL '%s' failed: %v", s, err)
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	c, ok := dialerMap[scheme]
	if !ok {
		log.Warn("[plugin] unknown scheme '%s' in URL: %s. Available schemes: %v", u.Scheme, s, getAvailableSchemes())
		return nil, fmt.Errorf("unknown scheme '%s'", u.Scheme)
	}

	log.Trace("[plugin] creating dialer for scheme '%s' from URL: %s", scheme, s)
	result, err := c(s, dialer)
	if err != nil {
		log.Warn("[plugin] failed to create dialer for scheme '%s': %v", scheme, err)
		return nil, fmt.Errorf("create %s dialer: %w", scheme, err)
	}

	log.Trace("[plugin] successfully created dialer for scheme '%s'", scheme)
	return result, nil
}

// ShouldSetMark determines if SO_MARK should be set on sockets.
// SO_MARK should only be set on Linux with root privileges.
// Returns 0x80 if mark should be set, 0 otherwise.
func ShouldSetMark() uint32 {
	// Don't set mark on Windows or macOS (they don't support redirect/tproxy)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return 0
	}

	// Don't set mark in lite mode (no root privileges)
	if conf.GetEnvironmentConfig().Lite {
		return 0
	}

	// Set mark on Linux with root privileges
	return 0x80
}
