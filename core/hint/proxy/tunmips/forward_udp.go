// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"io"
	stdnet "net"
	"net/netip"
	"sync"
	"time"

	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/features/stats"
)

// udpQueue bounds datagrams waiting for the outbound per session. Beyond it
// datagrams are dropped, as xray's tun inbound does.
const udpQueue = 1024

type replyFunc = func(payload []byte, from netip.AddrPort) error

type udpPacket struct {
	data []byte
	dest netip.AddrPort
}

// udpSessions keys sessions by the application's source address only, so
// datagrams from one socket to any number of remotes share one outbound
// session and replies from any remote find their way back: full-cone NAT.
type udpSessions struct {
	f  *forwarder
	mu sync.RWMutex
	m  map[netip.AddrPort]*udpConn
}

func newUDPSessions(f *forwarder) *udpSessions {
	return &udpSessions{f: f, m: make(map[netip.AddrPort]*udpConn)}
}

// handlePacket is the stack's UDP callback; it must not block. payload is
// only valid during the call, so it is copied.
func (u *udpSessions) handlePacket(flow Flow, payload []byte, reply replyFunc) {
	if u.f.dns != nil && u.f.dns.matches(flow.Destination) {
		u.f.dns.handleUDP(flow, payload, reply)
		return
	}
	pkt := &udpPacket{data: append([]byte(nil), payload...), dest: flow.Destination}

	// The send happens under the lock because finished closes egress under
	// the write lock; a send after that close would panic.
	u.mu.RLock()
	if conn, ok := u.m[flow.Source]; ok {
		conn.enqueue(pkt)
		u.mu.RUnlock()
		return
	}
	u.mu.RUnlock()

	u.mu.Lock()
	conn, ok := u.m[flow.Source]
	if !ok {
		conn = &udpConn{sessions: u, src: flow.Source, first: flow.Destination, reply: reply, egress: make(chan *udpPacket, udpQueue)}
		u.m[flow.Source] = conn
		go u.run(conn)
	}
	conn.enqueue(pkt)
	u.mu.Unlock()
}

func (c *udpConn) enqueue(pkt *udpPacket) {
	select {
	case c.egress <- pkt:
	default:
		errors.LogDebug(context.Background(), "tun-mips: drop udp ", len(pkt.data), " bytes to ", pkt.dest, ": queue full")
	}
}

func (u *udpSessions) run(conn *udpConn) {
	defer conn.Close()
	u.f.dispatch(conn, net.Network_UDP, Flow{Source: conn.src, Destination: conn.first})
}

func (u *udpSessions) finished(src netip.AddrPort) {
	u.mu.Lock()
	if conn, ok := u.m[src]; ok {
		delete(u.m, src)
		close(conn.egress)
	}
	u.mu.Unlock()
}

// udpConn presents one application socket's datagrams to the dispatcher as
// a packet connection. Each buffer read carries its destination in
// buf.Buffer.UDP; each buffer written carries the remote it came from,
// which becomes the source address of the datagram sent back.
type udpConn struct {
	sessions *udpSessions
	src      netip.AddrPort
	first    netip.AddrPort
	reply    replyFunc
	egress   chan *udpPacket
	// uplink and downlink, when set, receive the byte counts the inbound
	// statistics want; see forwarder.dispatch.
	uplink, downlink stats.Counter
}

func (c *udpConn) count(counter stats.Counter, n int) {
	if counter != nil {
		counter.Add(int64(n))
	}
}

func (c *udpConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	p, ok := <-c.egress
	if !ok {
		return nil, io.EOF
	}
	b := buf.New()
	if _, err := b.Write(p.data); err != nil {
		b.Release()
		return nil, err
	}
	dest := toDestination(net.Network_UDP, p.dest)
	b.UDP = &dest
	c.count(c.uplink, len(p.data))
	return buf.MultiBuffer{b}, nil
}

func (c *udpConn) Read(p []byte) (int, error) {
	pkt, ok := <-c.egress
	if !ok {
		return 0, io.EOF
	}
	if len(p) < len(pkt.data) {
		return 0, io.ErrShortBuffer
	}
	c.count(c.uplink, len(pkt.data))
	return copy(p, pkt.data), nil
}

func (c *udpConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for i, b := range mb {
		from := c.first
		if b.UDP != nil {
			if ap, ok := toAddrPort(*b.UDP); ok {
				from = ap
			} else {
				errors.LogError(context.Background(), "tun-mips: impossible domain packet ", b.UDP, ", replying from ", from)
			}
		}
		if err := c.reply(b.Bytes(), from); err != nil {
			buf.ReleaseMulti(mb[i:])
			return err
		}
		c.count(c.downlink, int(b.Len()))
		b.Release()
	}
	return nil
}

func (c *udpConn) Write(p []byte) (int, error) {
	if err := c.reply(p, c.first); err != nil {
		return 0, err
	}
	c.count(c.downlink, len(p))
	return len(p), nil
}

func (c *udpConn) Close() error {
	c.sessions.finished(c.src)
	return nil
}

func (c *udpConn) LocalAddr() net.Addr              { return stdnet.UDPAddrFromAddrPort(c.first) }
func (c *udpConn) RemoteAddr() net.Addr             { return stdnet.UDPAddrFromAddrPort(c.src) }
func (c *udpConn) SetDeadline(time.Time) error      { return nil }
func (c *udpConn) SetReadDeadline(time.Time) error  { return nil }
func (c *udpConn) SetWriteDeadline(time.Time) error { return nil }
