// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	stdnet "net"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/transport"
)

// fakeDispatcher records every DispatchLink and lets the test drive the
// link from the outbound side.
type fakeDispatcher struct {
	mu    sync.Mutex
	calls []dispatchCall
	ready chan *dispatchCall
	// hold keeps DispatchLink blocked until closed, like a live outbound.
	hold chan struct{}
}

type dispatchCall struct {
	ctx     context.Context
	dest    net.Destination
	inbound *session.Inbound
	link    *transport.Link
}

func newFakeDispatcher() *fakeDispatcher {
	return &fakeDispatcher{ready: make(chan *dispatchCall, 8), hold: make(chan struct{})}
}

func (d *fakeDispatcher) Type() interface{} { return (*fakeDispatcher)(nil) }
func (d *fakeDispatcher) Start() error      { return nil }
func (d *fakeDispatcher) Close() error      { return nil }
func (d *fakeDispatcher) Dispatch(context.Context, net.Destination) (*transport.Link, error) {
	panic("not used")
}

func (d *fakeDispatcher) DispatchLink(ctx context.Context, dest net.Destination, link *transport.Link) error {
	call := &dispatchCall{ctx: ctx, dest: dest, inbound: session.InboundFromContext(ctx), link: link}
	d.mu.Lock()
	d.calls = append(d.calls, *call)
	d.mu.Unlock()
	d.ready <- call
	<-d.hold
	return nil
}

func (d *fakeDispatcher) next(t *testing.T) *dispatchCall {
	t.Helper()
	select {
	case c := <-d.ready:
		return c
	case <-time.After(2 * time.Second):
		t.Fatal("DispatchLink not called")
	}
	return nil
}

func (d *fakeDispatcher) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.calls)
}

func TestTCPFlowIsDispatchedOnce(t *testing.T) {
	d := newFakeDispatcher()
	f := newForwarder(context.Background(), d)
	f.tag = "transparent"
	h := f.handlers()

	app, stackSide := stdnet.Pipe()
	flow := Flow{Source: netip.MustParseAddrPort("10.0.85.2:40000"), Destination: netip.MustParseAddrPort("1.1.1.1:443")}
	go h.TCP(flow, stackSide)

	call := d.next(t)
	if call.dest.Network != net.Network_TCP || call.dest.Address.IP().String() != "1.1.1.1" || call.dest.Port != 443 {
		t.Fatalf("dispatched to %v", call.dest)
	}
	if call.inbound == nil || call.inbound.Tag != "transparent" || call.inbound.Source.Port != 40000 {
		t.Fatalf("inbound session %+v", call.inbound)
	}

	// Application bytes reach the link; link bytes reach the application.
	go app.Write([]byte("hello"))
	mb, err := call.link.Reader.ReadMultiBuffer()
	if err != nil || mb.String() != "hello" {
		t.Fatalf("link read %q, %v", mb.String(), err)
	}
	go call.link.Writer.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes([]byte("world"))})
	got := make([]byte, 5)
	if _, err := app.Read(got); err != nil || string(got) != "world" {
		t.Fatalf("app read %q, %v", got, err)
	}
	close(d.hold)
	if d.count() != 1 {
		t.Fatalf("DispatchLink called %d times", d.count())
	}
}

func TestUDPSessionIsPerSourceWithPerPacketDestinationAndReplySource(t *testing.T) {
	d := newFakeDispatcher()
	f := newForwarder(context.Background(), d)
	h := f.handlers()

	src := netip.MustParseAddrPort("10.0.85.2:40000")
	a := netip.MustParseAddrPort("1.1.1.1:3478")
	b := netip.MustParseAddrPort("2.2.2.2:3478")
	c := netip.MustParseAddrPort("3.3.3.3:3478")

	type sent struct {
		payload []byte
		from    netip.AddrPort
	}
	replies := make(chan sent, 4)
	reply := func(p []byte, from netip.AddrPort) error {
		replies <- sent{append([]byte(nil), p...), from}
		return nil
	}

	h.UDP(Flow{Source: src, Destination: a}, []byte("to-a"), reply)
	h.UDP(Flow{Source: src, Destination: b}, []byte("to-b"), reply)

	call := d.next(t)
	if call.dest.Network != net.Network_UDP || call.dest.Port != 3478 || call.dest.Address.IP().String() != "1.1.1.1" {
		t.Fatalf("session dispatched to %v", call.dest)
	}
	for _, want := range []struct {
		payload string
		dest    netip.AddrPort
	}{{"to-a", a}, {"to-b", b}} {
		mb, err := call.link.Reader.ReadMultiBuffer()
		if err != nil {
			t.Fatal(err)
		}
		if mb.String() != want.payload || mb[0].UDP == nil {
			t.Fatalf("read %q UDP=%v", mb.String(), mb[0].UDP)
		}
		if got, _ := toAddrPort(*mb[0].UDP); got != want.dest {
			t.Fatalf("buffer destination %v, want %v", got, want.dest)
		}
	}
	// Two destinations, one source: still one session.
	if d.count() != 1 {
		t.Fatalf("DispatchLink called %d times, want 1", d.count())
	}

	// A packet from a third remote is written back with that remote as
	// its source.
	from := toDestination(net.Network_UDP, c)
	bb := buf.FromBytes([]byte("from-c"))
	bb.UDP = &from
	if err := call.link.Writer.WriteMultiBuffer(buf.MultiBuffer{bb}); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-replies:
		if string(r.payload) != "from-c" || r.from != c {
			t.Fatalf("reply %q from %v, want from %v", r.payload, r.from, c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no reply written")
	}

	// When the outbound finishes, the session goes away and the next
	// datagram from the same source starts a new one.
	close(d.hold)
	deadline := time.Now().Add(2 * time.Second)
	for {
		f.udp.mu.RLock()
		_, alive := f.udp.m[src]
		f.udp.mu.RUnlock()
		if !alive {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("session not removed after the outbound finished")
		}
		time.Sleep(10 * time.Millisecond)
	}
	d.hold = make(chan struct{})
	h.UDP(Flow{Source: src, Destination: a}, []byte("again"), reply)
	d.next(t)
	if d.count() != 2 {
		t.Fatalf("DispatchLink called %d times, want 2", d.count())
	}
	close(d.hold)
}

// countingCounter is a stats.Counter that only adds.
type countingCounter struct{ n int64 }

func (c *countingCounter) Value() int64      { return c.n }
func (c *countingCounter) Set(v int64) int64 { c.n = v; return v }
func (c *countingCounter) Add(v int64) int64 { c.n += v; return c.n }

// With inbound statistics on, xray wraps connections in a
// stat.CounterConnection that only forwards Read and Write. A UDP session
// read through that path loses the destination on every buffer and every
// datagram goes to the first destination; the counters must not cost the
// per-packet addressing.
func TestUDPStatisticsKeepPerPacketDestination(t *testing.T) {
	d := newFakeDispatcher()
	f := newForwarder(context.Background(), d)
	up, down := &countingCounter{}, &countingCounter{}
	f.uplink, f.downlink = up, down
	h := f.handlers()

	src := netip.MustParseAddrPort("10.0.85.2:40001")
	a := netip.MustParseAddrPort("1.1.1.1:3478")
	b := netip.MustParseAddrPort("2.2.2.2:3478")
	replies := make(chan netip.AddrPort, 2)
	reply := func(p []byte, from netip.AddrPort) error { replies <- from; return nil }

	h.UDP(Flow{Source: src, Destination: a}, []byte("to-a"), reply)
	h.UDP(Flow{Source: src, Destination: b}, []byte("to-bb"), reply)
	call := d.next(t)
	for _, want := range []netip.AddrPort{a, b} {
		mb, err := call.link.Reader.ReadMultiBuffer()
		if err != nil {
			t.Fatal(err)
		}
		if mb[0].UDP == nil {
			t.Fatalf("buffer %q lost its destination", mb.String())
		}
		if got, _ := toAddrPort(*mb[0].UDP); got != want {
			t.Fatalf("buffer destination %v, want %v", got, want)
		}
	}
	if up.Value() != 9 {
		t.Fatalf("uplink counted %d bytes, want 9", up.Value())
	}
	from := toDestination(net.Network_UDP, b)
	bb := buf.FromBytes([]byte("from-b"))
	bb.UDP = &from
	if err := call.link.Writer.WriteMultiBuffer(buf.MultiBuffer{bb}); err != nil {
		t.Fatal(err)
	}
	if got := <-replies; got != b {
		t.Fatalf("reply from %v, want %v", got, b)
	}
	if down.Value() != 6 {
		t.Fatalf("downlink counted %d bytes, want 6", down.Value())
	}
	close(d.hold)
}

func TestToAddrPortRejectsDomains(t *testing.T) {
	if _, ok := toAddrPort(net.UDPDestination(net.DomainAddress("example.com"), 53)); ok {
		t.Fatal("domain destination converted to an address")
	}
	ap, ok := toAddrPort(net.UDPDestination(net.IPAddress([]byte{8, 8, 8, 8}), 53))
	if !ok || ap != netip.MustParseAddrPort("8.8.8.8:53") {
		t.Fatalf("got %v %v", ap, ok)
	}
}
