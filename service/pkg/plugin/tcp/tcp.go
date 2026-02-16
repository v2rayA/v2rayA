package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Tcp is a base tcp struct.
type Tcp struct {
	dialer      plugin.Dialer
	proxy       plugin.Proxy
	listenAddr  string
	nodeName    string
	target      string
	TcpListener net.Listener
}

func init() {
	plugin.RegisterServer("tcp", NewTcpServer)
}

func NewTcp(s string, nodeName string, d plugin.Dialer, p plugin.Proxy) (*Tcp, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Warn("parse err: %s", err)
		return nil, err
	}

	addr := u.Host

	h := &Tcp{
		dialer:     d,
		proxy:      p,
		listenAddr: addr,
		nodeName:   nodeName,
		target:     u.Query().Get("target"),
	}

	return h, nil
}

// NewTcpDialer returns a tcp proxy dialer.
func NewTcpDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	return NewTcp(s, "", d, nil)
}

// NewTcpServer returns a tcp proxy server.
func NewTcpServer(s string, nodeName string, p plugin.Proxy) (plugin.Server, error) {
	return NewTcp(s, nodeName, nil, p)
}

// ListenAndServe serves tcp requests.
func (s *Tcp) ListenAndServe() error {
	//go s.ListenAndServeUDP()
	return s.ListenAndServeTCP()
}

func (s *Tcp) ListenAddr() string {
	return s.listenAddr
}

func (s *Tcp) NodeName() string {
	return s.nodeName
}

// ListenAndServeTCP listen and serve on tcp port.
func (s *Tcp) ListenAndServeTCP() error {
	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.Warn("[tcp] failed to listen on %s: %v", s.listenAddr, err)
		return err
	}
	s.TcpListener = l

	log.Trace("[tcp] listening TCP on %s", s.listenAddr)

	for {
		c, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Debug("[tcp] failed to accept: %v", err)
			continue
		}

		go s.Serve(c)
	}
}

func (s *Tcp) Close() error {
	if s.TcpListener == nil {
		return nil
	}
	return s.TcpListener.Close()
}

// Serve serves a connection.
func (s *Tcp) Serve(c net.Conn) {
	defer c.Close()

	if c, ok := c.(*net.TCPConn); ok {
		c.SetKeepAlive(true)
	}

	rc, dialer, err := s.proxy.Dial("tcp", s.target)
	if err != nil {
		log.Debug("[tcp] %s <-> %s, error in dial: %v", c.RemoteAddr(), dialer, err)
		return
	}
	defer rc.Close()

	log.Trace("[tcp] %s <-> %s", c.RemoteAddr(), dialer)

	_, _, err = Relay(c, rc)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return // ignore i/o timeout
		}
		log.Debug("[tcp] relay error: %v", err)
	}
}

// Relay relays between left and right.
func Relay(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		n, err := io.Copy(right, left)
		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
		ch <- res{n, err}
	}()

	n, err := io.Copy(left, right)
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}
	return n, rs.N, err
}

// Addr returns forwarder's address.
func (s *Tcp) Addr() string {
	if s.listenAddr == "" {
		return s.dialer.Addr()
	}
	return s.listenAddr
}

// Dial connects to the address addr on the network net via the TCP proxy.
func (s *Tcp) Dial(network, addr string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, addr)
}

func (s *Tcp) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[tcp] dialing %s via %s", addr, s.listenAddr)
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		return nil, fmt.Errorf("[tcp]: no support for connection type " + network)
	}

	c, err := s.dialer.DialContext(ctx, network, s.listenAddr)
	if err != nil {
		log.Debug("[tcp]: dial to %s error: %s", s.listenAddr, err)
		return nil, err
	}

	return c, nil
}

// DialUDP connects to the given address via the proxy.
func (s *Tcp) DialUDP(network string) (pc plugin.FakeNetPacketConn, err error) {
	//Not support
	return nil, nil
}
