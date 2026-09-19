// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"net/netip"

	"github.com/xtls/xray-core/common/buf"
	c "github.com/xtls/xray-core/common/ctx"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/log"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/transport"
	"github.com/xtls/xray-core/transport/internet/stat"
)

// forwarder turns stack callbacks into dispatcher sessions. It mirrors
// xray's proxy/tun handler: one DispatchLink per TCP connection, and one
// per UDP source with every datagram carrying its own destination so the
// outbound can do full-cone NAT.
type forwarder struct {
	ctx        context.Context
	dispatcher routing.Dispatcher
	tag        string
	sniffing   session.SniffingRequest
	userLevel  uint32
	uplink     stats.Counter
	downlink   stats.Counter

	udp *udpSessions
	// dns, when set, answers port-53 flows to non-local resolvers itself.
	dns *dnsHijack
	// excl, when set, sends flows owned by listed processes to directTag.
	excl      *exclusion
	directTag string
}

func newForwarder(ctx context.Context, dispatcher routing.Dispatcher) *forwarder {
	f := &forwarder{ctx: ctx, dispatcher: dispatcher}
	f.udp = newUDPSessions(f)
	return f
}

func (f *forwarder) handlers() Handlers {
	return Handlers{TCP: f.handleTCP, UDP: f.udp.handlePacket}
}

func (f *forwarder) handleTCP(flow Flow, conn net.Conn) {
	defer conn.Close()
	if f.dns != nil && f.dns.matches(flow.Destination) {
		f.dns.handleTCP(flow, conn)
		return
	}
	f.dispatch(conn, net.Network_TCP, flow)
}

// dispatch runs one session to completion. DispatchLink returns when the
// outbound is done with the link, so the caller may close conn afterwards.
func (f *forwarder) dispatch(conn net.Conn, network net.Network, flow Flow) {
	source := toDestination(network, flow.Source)
	destination := toDestination(network, flow.Destination)

	ctx, cancel := context.WithCancel(f.ctx)
	defer cancel()
	ctx = c.ContextWithID(ctx, session.NewID())

	if f.uplink != nil || f.downlink != nil {
		// A stat.CounterConnection only forwards Read and Write, so wrapping
		// a udpConn would hide its ReadMultiBuffer and lose the per-datagram
		// destination that full-cone NAT relies on. The udpConn counts its
		// own bytes instead.
		if uc, ok := conn.(*udpConn); ok {
			uc.uplink, uc.downlink = f.uplink, f.downlink
		} else {
			conn = &stat.CounterConnection{Connection: conn, ReadCounter: f.uplink, WriteCounter: f.downlink}
		}
	}

	inbound := session.Inbound{
		Name:          "tun-mips",
		Tag:           f.tag,
		CanSpliceCopy: 3,
		Source:        source,
		User:          &protocol.MemoryUser{Level: f.userLevel},
	}
	ctx = session.ContextWithInbound(ctx, &inbound)
	ctx = session.ContextWithContent(ctx, &session.Content{SniffingRequest: f.sniffing})
	ctx = session.SubContextFromMuxInbound(ctx)
	// Process exclusion is a routing decision: the owner is looked up once
	// per session and, when listed, the dispatcher is told to skip the
	// router and use the direct outbound. The tag is an attribute of the
	// session Content, which must exist and must not yet have attributes
	// when SubContextFromMuxInbound runs, hence the order.
	if f.excl != nil && f.excl.excluded(network.SystemString(), flow.Source) {
		ctx = session.SetForcedOutboundTagToContext(ctx, f.directTag)
	}
	ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
		From:   source,
		To:     destination,
		Status: log.AccessAccepted,
	})
	errors.LogInfo(ctx, "processing from ", source, " to ", destination)

	link := &transport.Link{
		Reader: &buf.TimeoutWrapperReader{Reader: buf.NewReader(conn)},
		Writer: buf.NewWriter(conn),
	}
	if err := f.dispatcher.DispatchLink(ctx, destination, link); err != nil {
		errors.LogError(ctx, errors.New("connection closed").Base(err))
	}
}

func toDestination(network net.Network, ap netip.AddrPort) net.Destination {
	return net.Destination{Network: network, Address: net.IPAddress(ap.Addr().AsSlice()), Port: net.Port(ap.Port())}
}

// toAddrPort converts back; ok is false for domain destinations, which a
// reply can never be sourced from.
func toAddrPort(d net.Destination) (netip.AddrPort, bool) {
	if d.Address == nil || d.Address.Family().IsDomain() {
		return netip.AddrPort{}, false
	}
	addr, ok := netip.AddrFromSlice(d.Address.IP())
	if !ok {
		return netip.AddrPort{}, false
	}
	return netip.AddrPortFrom(addr.Unmap(), uint16(d.Port)), true
}
