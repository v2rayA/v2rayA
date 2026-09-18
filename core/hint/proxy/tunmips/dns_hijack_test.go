// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	stdnet "net"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// fakeDNSTarget answers every UDP datagram and every TCP length-prefixed
// message with "ans:" + the query, on one loopback port for both.
func fakeDNSTarget(t *testing.T) string {
	t.Helper()
	pc, err := stdnet.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := pc.LocalAddr().(*stdnet.UDPAddr).Port
	l, err := stdnet.Listen("tcp", stdnet.JoinHostPort("127.0.0.1", itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close(); l.Close() })
	go func() {
		buf := make([]byte, 4096)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			pc.WriteTo(append([]byte("ans:"), buf[:n]...), from)
		}
	}()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				var hdr [2]byte
				if _, err := io.ReadFull(c, hdr[:]); err != nil {
					return
				}
				q := make([]byte, binary.BigEndian.Uint16(hdr[:]))
				if _, err := io.ReadFull(c, q); err != nil {
					return
				}
				a := append([]byte("ans:"), q...)
				binary.BigEndian.PutUint16(hdr[:], uint16(len(a)))
				c.Write(append(hdr[:], a...))
			}()
		}
	}()
	return pc.LocalAddr().String()
}

func itoa(i int) string { return strconv.Itoa(i) }

func TestDNSHijackUDPRepliesFromQueriedResolver(t *testing.T) {
	d := newFakeDispatcher()
	f := newForwarder(context.Background(), d)
	f.dns = newDNSHijack(fakeDNSTarget(t))
	h := f.handlers()

	src := netip.MustParseAddrPort("10.0.85.2:40000")
	resolver := netip.MustParseAddrPort("8.8.8.8:53")
	type sent struct {
		payload []byte
		from    netip.AddrPort
	}
	replies := make(chan sent, 1)
	h.UDP(Flow{Source: src, Destination: resolver}, []byte("q1"), func(p []byte, from netip.AddrPort) error {
		replies <- sent{append([]byte(nil), p...), from}
		return nil
	})
	select {
	case r := <-replies:
		if string(r.payload) != "ans:q1" || r.from != resolver {
			t.Fatalf("reply %q from %v", r.payload, r.from)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no DNS reply")
	}
	if d.count() != 0 {
		t.Fatalf("port-53 datagram reached the dispatcher (%d calls)", d.count())
	}

	// Port 443 to the same host is ordinary traffic.
	h.UDP(Flow{Source: src, Destination: netip.MustParseAddrPort("8.8.8.8:443")}, []byte("x"), func([]byte, netip.AddrPort) error { return nil })
	d.next(t)
	close(d.hold)
}

func TestDNSHijackTCPRelaysToTarget(t *testing.T) {
	d := newFakeDispatcher()
	f := newForwarder(context.Background(), d)
	f.dns = newDNSHijack(fakeDNSTarget(t))
	h := f.handlers()

	app, stackSide := stdnet.Pipe()
	go h.TCP(Flow{Source: netip.MustParseAddrPort("10.0.85.2:40001"), Destination: netip.MustParseAddrPort("1.1.1.1:53")}, stackSide)

	msg := []byte{0, 2, 'q', '2'}
	if _, err := app.Write(msg); err != nil {
		t.Fatal(err)
	}
	app.SetReadDeadline(time.Now().Add(2 * time.Second))
	var hdr [2]byte
	if _, err := io.ReadFull(app, hdr[:]); err != nil {
		t.Fatal(err)
	}
	a := make([]byte, binary.BigEndian.Uint16(hdr[:]))
	if _, err := io.ReadFull(app, a); err != nil || string(a) != "ans:q2" {
		t.Fatalf("answer %q, %v", a, err)
	}
	if d.count() != 0 {
		t.Fatal("TCP DNS reached the dispatcher")
	}
}

func TestDNSHijackLeavesLocalResolversAlone(t *testing.T) {
	h := newDNSHijack("127.0.0.1:1")
	if h.matches(netip.MustParseAddrPort("127.2.0.17:53")) {
		t.Fatal("loopback resolver hijacked")
	}
	if h.matches(netip.MustParseAddrPort("8.8.8.8:853")) {
		t.Fatal("non-53 port hijacked")
	}
	if !h.matches(netip.MustParseAddrPort("10.0.85.1:53")) {
		t.Fatal("gateway address not hijacked")
	}
	if !h.matches(netip.MustParseAddrPort("[2001:4860:4860::8888]:53")) {
		t.Fatal("IPv6 resolver not hijacked")
	}
	addrs, _ := stdnet.InterfaceAddrs()
	for _, a := range addrs {
		if ipn, ok := a.(*stdnet.IPNet); ok {
			ip, _ := netip.AddrFromSlice(ipn.IP)
			if h.matches(netip.AddrPortFrom(ip.Unmap(), 53)) {
				t.Fatalf("host address %v hijacked", ip)
			}
		}
	}
}

// A client advertising a 64 KiB EDNS buffer must not be answered with a
// datagram larger than the hijack's reply buffer; the advertised size is
// capped on the way to the module, everything else is passed as is.
func TestDNSHijackClampsAdvertisedUDPSize(t *testing.T) {
	q := new(dns.Msg)
	q.SetQuestion("example.com.", dns.TypeA)
	q.SetEdns0(65535, false)
	raw, _ := q.Pack()
	out := new(dns.Msg)
	if err := out.Unpack(clampUDPSize(raw)); err != nil {
		t.Fatal(err)
	}
	if got := out.IsEdns0().UDPSize(); got != dnsMaxMessage {
		t.Fatalf("advertised size %d, want %d", got, dnsMaxMessage)
	}
	if out.Id != q.Id || out.Question[0].Name != "example.com." {
		t.Fatalf("query altered: %v", out)
	}
	q.SetEdns0(1232, false)
	raw, _ = q.Pack()
	if got := clampUDPSize(raw); !bytes.Equal(got, raw) {
		t.Fatal("a query within the cap was rewritten")
	}
	plain := new(dns.Msg)
	plain.SetQuestion("example.com.", dns.TypeA)
	raw, _ = plain.Pack()
	if got := clampUDPSize(raw); !bytes.Equal(got, raw) {
		t.Fatal("a query without EDNS was rewritten")
	}
	if got := clampUDPSize([]byte{1, 2, 3}); !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Fatal("garbage was rewritten")
	}
}
