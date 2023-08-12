// https://tools.ietf.org/html/rfc1928

// socks5 client:
// https://github.com/golang/net/tree/master/proxy
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package socks5 implements a socks5 proxy.
package socks5

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/v2rayA/shadowsocksR/tools/leakybuf"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/plugin/socks"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Version is socks5 version number.
const Version = 5

// Socks5 is a base socks5 struct.
type Socks5 struct {
	dialer      plugin.Dialer
	proxy       plugin.Proxy
	addr        string
	user        string
	password    string
	TcpListener net.Listener
}

func init() {
	plugin.RegisterServer("socks5", NewSocks5Server)
	plugin.RegisterServer("socks", NewSocks5Server)
	plugin.RegisterDialer("socks5", NewSocks5Dialer)
	plugin.RegisterDialer("socks", NewSocks5Dialer)
}

// NewSocks5 returns a Proxy that makes SOCKS v5 connections to the given address
// with an optional username and password. (RFC 1928)
func NewSocks5(s string, d plugin.Dialer, p plugin.Proxy) (*Socks5, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Warn("parse err: %s", err)
		return nil, err
	}

	addr := u.Host
	user := u.User.Username()
	pass, _ := u.User.Password()

	h := &Socks5{
		dialer:   d,
		proxy:    p,
		addr:     addr,
		user:     user,
		password: pass,
	}

	return h, nil
}

// NewSocks5Dialer returns a socks5 proxy dialer.
func NewSocks5Dialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	return NewSocks5(s, d, nil)
}

// NewSocks5Server returns a socks5 proxy server.
func NewSocks5Server(s string, p plugin.Proxy) (plugin.Server, error) {
	return NewSocks5(s, nil, p)
}

// ListenAndServe serves socks5 requests.
func (s *Socks5) ListenAndServe() error {
	//go s.ListenAndServeUDP()
	return s.ListenAndServeTCP()
}

func (s *Socks5) ListenAddr() string {
	return s.addr
}

// ListenAndServeTCP listen and serve on tcp port.
func (s *Socks5) ListenAndServeTCP() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Warn("[socks5] failed to listen on %s: %v", s.addr, err)
		return err
	}
	s.TcpListener = l

	log.Trace("[socks5] listening TCP on %s", s.addr)

	for {
		c, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Debug("[socks5] failed to accept: %v", err)
			continue
		}

		go s.Serve(c)
	}
}

func (s *Socks5) Close() error {
	if s.TcpListener == nil {
		return nil
	}
	return s.TcpListener.Close()
}

// Serve serves a connection.
func (s *Socks5) Serve(c net.Conn) {
	defer c.Close()

	tgt, err := s.handshake(c)
	if err != nil {
		// UDP: keep the connection until disconnect then free the UDP socket
		if err == socks.Errors[9] {
			buf := leakybuf.GlobalLeakyBuf.Get()
			defer leakybuf.GlobalLeakyBuf.Put(buf)
			// block here
			for {
				_, err := c.Read(buf)
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				// log.Println("[socks5] servetcp udp associate end")
				return
			}
		}

		log.Debug("[socks5] failed in handshake with %s: %v", c.RemoteAddr(), err)
		return
	}

	rc, dialer, err := s.proxy.Dial("tcp", tgt.String())
	if err != nil {
		log.Debug("[socks5] %s <-> %s via %s, error in dial: %v", c.RemoteAddr(), tgt, dialer, err)
		return
	}
	defer rc.Close()

	log.Trace("[socks5] %s <-> %s via %s", c.RemoteAddr(), tgt, dialer)

	_, _, err = Relay(c, rc)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return // ignore i/o timeout
		}
		log.Debug("[socks5] relay error: %v", err)
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
func (s *Socks5) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the SOCKS5 proxy.
func (s *Socks5) Dial(network, addr string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		return nil, fmt.Errorf("[socks5]: no support for connection type %v", network)
	}

	c, err := s.dialer.Dial(network, s.addr)
	if err != nil {
		log.Debug("[socks5]: dial to %s error: %s", s.addr, err)
		return nil, err
	}

	if err := s.connect(c, addr); err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}

// DialUDP connects to the given address via the proxy.
func (s *Socks5) DialUDP(network, addr string) (pc net.PacketConn, writeTo net.Addr, err error) {
	c, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		log.Debug("[socks5] dialudp dial tcp to %s error: %s", s.addr, err)
		return nil, nil, err
	}

	// send VER, NMETHODS, METHODS
	c.Write([]byte{Version, 1, 0})

	buf := make([]byte, socks.MaxAddrLen)
	// read VER METHOD
	if _, err := io.ReadFull(c, buf[:2]); err != nil {
		return nil, nil, err
	}

	dstAddr := socks.ParseAddr(addr)
	// write VER CMD RSV ATYP DST.ADDR DST.PORT
	c.Write(append([]byte{Version, socks.CmdUDPAssociate, 0}, dstAddr...))

	// read VER REP RSV ATYP BND.ADDR BND.PORT
	if _, err := io.ReadFull(c, buf[:3]); err != nil {
		return nil, nil, err
	}

	rep := buf[1]
	if rep != 0 {
		log.Debug("[socks5] server reply: %d, not succeeded", rep)
		return nil, nil, fmt.Errorf("server connect failed")
	}

	uAddr, err := socks.ReadAddrBuf(c, buf)
	if err != nil {
		return nil, nil, err
	}

	pc, nextHop, err := s.dialer.DialUDP(network, uAddr.String())
	if err != nil {
		log.Debug("[socks5] dialudp to %s error: %s", uAddr.String(), err)
		return nil, nil, err
	}

	pkc := NewPktConn(pc, nextHop, dstAddr, true, c)
	return pkc, nextHop, err
}

// connect takes an existing connection to a socks5 proxy server,
// and commands the server to extend that connection to target,
// which must be a canonical address with a host and port.
func (s *Socks5) connect(conn net.Conn, target string) error {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("proxy: failed to parse port number: %v", portStr)
	}
	if port < 1 || port > 0xffff {
		return fmt.Errorf("proxy: port number out of range: %v", portStr)
	}

	// the size here is just an estimate
	buf := make([]byte, 0, 6+len(host))

	buf = append(buf, Version)
	if len(s.user) > 0 && len(s.user) < 256 && len(s.password) < 256 {
		buf = append(buf, 2 /* num auth methods */, socks.AuthNone, socks.AuthPassword)
	} else {
		buf = append(buf, 1 /* num auth methods */, socks.AuthNone)
	}

	if _, err := conn.Write(buf); err != nil {
		return fmt.Errorf("proxy: failed to write greeting to SOCKS5 proxy at %v: %w", s.addr, err)
	}

	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return fmt.Errorf("proxy: failed to read greeting from SOCKS5 proxy at %v: %w", s.addr, err)
	}
	if buf[0] != Version {
		return fmt.Errorf("proxy: SOCKS5 proxy at " + s.addr + " has unexpected version " + strconv.Itoa(int(buf[0])))
	}
	if buf[1] == 0xff {
		return fmt.Errorf("proxy: SOCKS5 proxy at " + s.addr + " requires authentication")
	}

	if buf[1] == socks.AuthPassword {
		buf = buf[:0]
		buf = append(buf, 1 /* password protocol version */)
		buf = append(buf, uint8(len(s.user)))
		buf = append(buf, s.user...)
		buf = append(buf, uint8(len(s.password)))
		buf = append(buf, s.password...)

		if _, err := conn.Write(buf); err != nil {
			return fmt.Errorf("proxy: failed to write authentication request to SOCKS5 proxy at %v: %w", s.addr, err)
		}

		if _, err := io.ReadFull(conn, buf[:2]); err != nil {
			return fmt.Errorf("proxy: failed to read authentication reply from SOCKS5 proxy at %v: %w", s.addr, err)
		}

		if buf[1] != 0 {
			return fmt.Errorf("proxy: SOCKS5 proxy at " + s.addr + " rejected username/password")
		}
	}

	buf = buf[:0]
	buf = append(buf, Version, socks.CmdConnect, 0 /* reserved */)

	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buf = append(buf, socks.ATypIP4)
			ip = ip4
		} else {
			buf = append(buf, socks.ATypIP6)
		}
		buf = append(buf, ip...)
	} else {
		if len(host) > 255 {
			return fmt.Errorf("proxy: destination hostname too long: " + host)
		}
		buf = append(buf, socks.ATypDomain)
		buf = append(buf, byte(len(host)))
		buf = append(buf, host...)
	}
	buf = append(buf, byte(port>>8), byte(port))

	if _, err := conn.Write(buf); err != nil {
		return fmt.Errorf("proxy: failed to write connect request to SOCKS5 proxy at %v: %w", s.addr, err)
	}

	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return fmt.Errorf("proxy: failed to read connect reply from SOCKS5 proxy at %v: %w", s.addr, err)
	}

	failure := "unknown error"
	if int(buf[1]) < len(socks.Errors) {
		failure = socks.Errors[buf[1]].Error()
	}

	if len(failure) > 0 {
		return fmt.Errorf("proxy: SOCKS5 proxy at " + s.addr + " failed to connect: " + failure)
	}

	bytesToDiscard := 0
	switch buf[3] {
	case socks.ATypIP4:
		bytesToDiscard = net.IPv4len
	case socks.ATypIP6:
		bytesToDiscard = net.IPv6len
	case socks.ATypDomain:
		_, err := io.ReadFull(conn, buf[:1])
		if err != nil {
			return fmt.Errorf("proxy: failed to read domain length from SOCKS5 proxy at %v: %w", s.addr, err)
		}
		bytesToDiscard = int(buf[0])
	default:
		return fmt.Errorf("proxy: got unknown address type %v", strconv.Itoa(int(buf[3]))+" from SOCKS5 proxy at "+s.addr)
	}

	if cap(buf) < bytesToDiscard {
		buf = make([]byte, bytesToDiscard)
	} else {
		buf = buf[:bytesToDiscard]
	}
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("proxy: failed to read address from SOCKS5 proxy at %v: %w", s.addr, err)
	}

	// Also need to discard the port number
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return fmt.Errorf("proxy: failed to read port from SOCKS5 proxy at %v, %w", s.addr, err)
	}

	return nil
}

// Handshake fast-tracks SOCKS initialization to get target address to connect.
func (s *Socks5) handshake(rw io.ReadWriter) (socks.Addr, error) {
	// Read RFC 1928 for request and reply structure and sizes
	buf := make([]byte, socks.MaxAddrLen)
	// read VER, NMETHODS, METHODS
	if _, err := io.ReadFull(rw, buf[:2]); err != nil {
		return nil, err
	}

	nmethods := buf[1]
	if _, err := io.ReadFull(rw, buf[:nmethods]); err != nil {
		return nil, err
	}

	// write VER METHOD
	if s.user != "" && s.password != "" {
		_, err := rw.Write([]byte{Version, socks.AuthPassword})
		if err != nil {
			return nil, err
		}

		_, err = io.ReadFull(rw, buf[:2])
		if err != nil {
			return nil, err
		}

		// Get username
		userLen := int(buf[1])
		if userLen <= 0 {
			rw.Write([]byte{1, 1})
			return nil, fmt.Errorf("auth failed: wrong username length")
		}

		if _, err := io.ReadFull(rw, buf[:userLen]); err != nil {
			return nil, fmt.Errorf("auth failed: cannot get username")
		}
		user := string(buf[:userLen])

		// Get password
		_, err = rw.Read(buf[:1])
		if err != nil {
			return nil, fmt.Errorf("auth failed: cannot get password len")
		}

		passLen := int(buf[0])
		if passLen <= 0 {
			rw.Write([]byte{1, 1})
			return nil, fmt.Errorf("auth failed: wrong password length")
		}

		_, err = io.ReadFull(rw, buf[:passLen])
		if err != nil {
			return nil, fmt.Errorf("auth failed: cannot get password")
		}
		pass := string(buf[:passLen])

		// Verify
		if user != s.user || pass != s.password {
			_, err = rw.Write([]byte{1, 1})
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("auth failed, authinfo: " + user + ":" + pass)
		}

		// Response auth state
		_, err = rw.Write([]byte{1, 0})
		if err != nil {
			return nil, err
		}

	} else if _, err := rw.Write([]byte{Version, socks.AuthNone}); err != nil {
		return nil, err
	}

	// read VER CMD RSV ATYP DST.ADDR DST.PORT
	if _, err := io.ReadFull(rw, buf[:3]); err != nil {
		return nil, err
	}
	cmd := buf[1]
	addr, err := socks.ReadAddrBuf(rw, buf)
	if err != nil {
		return nil, err
	}
	switch cmd {
	case socks.CmdConnect:
		_, err = rw.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}) // SOCKS v5, reply succeeded
	case socks.CmdUDPAssociate:
		listenAddr := socks.ParseAddr(rw.(net.Conn).LocalAddr().String())
		_, err = rw.Write(append([]byte{5, 0, 0}, listenAddr...)) // SOCKS v5, reply succeeded
		if err != nil {
			return nil, socks.Errors[7]
		}
		err = socks.Errors[9]
	default:
		return nil, socks.Errors[7]
	}

	return addr, err // skip VER, CMD, RSV fields
}
