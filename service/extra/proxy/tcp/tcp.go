package tcp

import (
	"github.com/v2rayA/v2rayA/extra/proxy"
	"github.com/v2rayA/v2rayA/global"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

// Tcp is a base tcp struct.
type Tcp struct {
	dialer      proxy.Dialer
	proxy       proxy.Proxy
	listenAddr  string
	target      string
	TcpListener net.Listener
}

func NewTcp(s string, d proxy.Dialer, p proxy.Proxy) (*Tcp, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Printf("parse err: %s\n", err)
		return nil, err
	}

	addr := u.Host

	h := &Tcp{
		dialer:     d,
		proxy:      p,
		listenAddr: addr,
		target:     u.Query().Get("target"),
	}

	return h, nil
}

// NewTcpDialer returns a tcp proxy dialer.
func NewTcpDialer(s string, d proxy.Dialer) (proxy.Dialer, error) {
	return NewTcp(s, d, nil)
}

// NewTcpServer returns a tcp proxy server.
func NewTcpServer(s string, p proxy.Proxy) (proxy.Server, error) {
	return NewTcp(s, nil, p)
}

// ListenAndServe serves tcp requests.
func (s *Tcp) ListenAndServe() error {
	//go s.ListenAndServeUDP()
	return s.ListenAndServeTCP()
}

// ListenAndServeTCP listen and serve on tcp port.
func (s *Tcp) ListenAndServeTCP() error {
	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		if global.IsDebug() {
			log.Printf("[tcp] failed to listen on %s: %v\n", s.listenAddr, err)
		}
		return err
	}
	s.TcpListener = l

	if global.IsDebug() {
		log.Printf("[tcp] listening TCP on %s\n", s.listenAddr)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			if global.IsDebug() {
				log.Printf("[tcp] failed to accept: %v\n", err)
			}
			continue
		}

		go s.Serve(c)
	}
}

func (s *Tcp) Close() error {
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
		if global.IsDebug() {
			log.Printf("[tcp] %s <-> %s, error in dial: %v", c.RemoteAddr(), dialer, err)
		}
		return
	}
	defer rc.Close()

	if global.IsDebug() {
		log.Printf("[tcp] %s <-> %s", c.RemoteAddr(), dialer)
	}

	_, _, err = Relay(c, rc)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return // ignore i/o timeout
		}
		if global.IsDebug() {
			log.Printf("[tcp] relay error: %v", err)
		}
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
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		return nil, newError("[tcp]: no support for connection type " + network)
	}

	c, err := s.dialer.Dial(network, s.listenAddr)
	if err != nil {
		if global.IsDebug() {
			log.Printf("[tcp]: dial to %s error: %s\n", s.listenAddr, err)
		}
		return nil, err
	}

	return c, nil
}

// DialUDP connects to the given address via the proxy.
func (s *Tcp) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	//Not support
	return nil, nil, nil
}
