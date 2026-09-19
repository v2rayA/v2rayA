// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"io"
	stdnet "net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	"github.com/xtls/xray-core/common/errors"
)

const (
	dnsPort       = 53
	dnsTimeout    = 10 * time.Second
	dnsUDPIdle    = 10 * time.Second
	dnsMaxMessage = 4096
	// dnsMaxSessions bounds the UDP sessions and, separately, the TCP
	// relays. Every process on the host can create flows at will; beyond
	// this the hijack answers nothing new until sessions expire, which
	// costs the attacker's own resolution and nothing else.
	dnsMaxSessions = 4096
	// dnsLocalRefresh is how long the snapshot of the host's addresses is
	// trusted before it is taken again.
	dnsLocalRefresh = 10 * time.Second
)

// dnsHijack answers every port-53 flow whose destination is not an address
// of this host by relaying it to the core's own DNS module instead of
// routing it. The application keeps talking to the resolver it chose;
// replies are sourced from that resolver's address.
type dnsHijack struct {
	target string
	// local is the host's own addresses; localAt is when it was taken.
	local   atomic.Pointer[map[netip.Addr]struct{}]
	localAt atomic.Int64

	mu       sync.Mutex
	sessions map[Flow]*dnsUDPSession
	tcpLive  int
}

func newDNSHijack(target string) *dnsHijack {
	h := &dnsHijack{target: target, sessions: make(map[Flow]*dnsUDPSession)}
	h.refreshLocal()
	return h
}

// refreshLocal snapshots the host's addresses: a query to one of them (the
// core's own listener on 127.2.0.17, a LAN resolver on this box) never
// enters the TUN in the first place, and if it somehow does it is not
// ours to redirect. Addresses come and go with DHCP, so the snapshot is
// retaken every dnsLocalRefresh by whichever query notices it is stale.
func (h *dnsHijack) refreshLocal() {
	local := make(map[netip.Addr]struct{})
	if addrs, err := stdnet.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			if ipn, ok := a.(*stdnet.IPNet); ok {
				if ip, ok := netip.AddrFromSlice(ipn.IP); ok {
					local[ip.Unmap()] = struct{}{}
				}
			}
		}
	}
	h.local.Store(&local)
	h.localAt.Store(time.Now().UnixNano())
}

// matches reports whether a flow is a DNS query this hijack should answer.
func (h *dnsHijack) matches(dest netip.AddrPort) bool {
	if dest.Port() != dnsPort {
		return false
	}
	addr := dest.Addr().Unmap()
	if addr.IsLoopback() {
		return false
	}
	if time.Since(time.Unix(0, h.localAt.Load())) > dnsLocalRefresh {
		h.refreshLocal()
	}
	_, local := (*h.local.Load())[addr]
	return !local
}

// handleTCP relays one TCP DNS stream to the module.
func (h *dnsHijack) handleTCP(flow Flow, conn stdnet.Conn) {
	defer conn.Close()
	h.mu.Lock()
	if h.tcpLive >= dnsMaxSessions {
		h.mu.Unlock()
		return
	}
	h.tcpLive++
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		h.tcpLive--
		h.mu.Unlock()
	}()
	// A stream nobody writes to is not kept open: the client has one
	// timeout to send a query, and the whole relay one to finish.
	conn.SetDeadline(time.Now().Add(2 * dnsTimeout))
	up, err := stdnet.DialTimeout("tcp", h.target, dnsTimeout)
	if err != nil {
		errors.LogWarning(context.Background(), "tun-mips: dns ", flow.Destination, " -> ", h.target, ": ", err)
		return
	}
	defer up.Close()
	done := make(chan struct{}, 2)
	relay := func(dst, src stdnet.Conn) {
		io.Copy(dst, src)
		if c, ok := dst.(interface{ CloseWrite() error }); ok {
			c.CloseWrite()
		}
		done <- struct{}{}
	}
	go relay(up, conn)
	go relay(conn, up)
	<-done
	// Give the other direction a moment to drain the final answer.
	select {
	case <-done:
	case <-time.After(dnsTimeout):
	}
}

// handleUDP forwards one datagram; the reply comes back through reply with
// the queried resolver as its source. Sessions are per four-tuple so each
// answer returns to the socket and resolver address it belongs to.
func (h *dnsHijack) handleUDP(flow Flow, payload []byte, reply replyFunc) {
	h.mu.Lock()
	s, ok := h.sessions[flow]
	if !ok {
		if len(h.sessions) >= dnsMaxSessions {
			h.mu.Unlock()
			errors.LogDebug(context.Background(), "tun-mips: dns session limit reached, dropping query from ", flow.Source)
			return
		}
		up, err := stdnet.Dial("udp", h.target)
		if err != nil {
			h.mu.Unlock()
			errors.LogWarning(context.Background(), "tun-mips: dns ", flow.Destination, " -> ", h.target, ": ", err)
			return
		}
		s = &dnsUDPSession{flow: flow, conn: up, reply: reply}
		h.sessions[flow] = s
		go h.readReplies(s)
	}
	s.touch()
	h.mu.Unlock()

	if _, err := s.conn.Write(clampUDPSize(payload)); err != nil {
		errors.LogDebug(context.Background(), "tun-mips: dns write to ", h.target, ": ", err)
	}
}

// clampUDPSize caps the EDNS buffer size a query advertises at
// dnsMaxMessage. The reply is read into a buffer of that size; a larger
// answer would arrive truncated without the TC bit and the client would
// wait out its timeout on a broken message, whereas an answer the module
// truncates at the advertised size carries TC and the client retries over
// TCP. Queries without EDNS, or within the cap, pass through untouched.
func clampUDPSize(payload []byte) []byte {
	msg := new(dns.Msg)
	if err := msg.Unpack(payload); err != nil {
		return payload
	}
	opt := msg.IsEdns0()
	if opt == nil || opt.UDPSize() <= dnsMaxMessage {
		return payload
	}
	opt.SetUDPSize(dnsMaxMessage)
	packed, err := msg.Pack()
	if err != nil {
		return payload
	}
	return packed
}

func (h *dnsHijack) readReplies(s *dnsUDPSession) {
	defer func() {
		h.mu.Lock()
		if h.sessions[s.flow] == s {
			delete(h.sessions, s.flow)
		}
		h.mu.Unlock()
		s.conn.Close()
	}()
	buf := make([]byte, dnsMaxMessage)
	for {
		s.conn.SetReadDeadline(s.deadline())
		n, err := s.conn.Read(buf)
		if err != nil {
			return
		}
		if err := s.reply(buf[:n], s.flow.Destination); err != nil {
			return
		}
	}
}

func (h *dnsHijack) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for flow, s := range h.sessions {
		s.conn.Close()
		delete(h.sessions, flow)
	}
}

type dnsUDPSession struct {
	flow  Flow
	conn  stdnet.Conn
	reply replyFunc

	mu   sync.Mutex
	last time.Time
}

func (s *dnsUDPSession) touch() {
	s.mu.Lock()
	s.last = time.Now()
	s.mu.Unlock()
}

func (s *dnsUDPSession) deadline() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last.Add(dnsUDPIdle)
}
