// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"bytes"
	"net"
	"net/netip"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/tun/tuntest"
)

// startTestStack wires a mipsStack to a channel device. Packets pushed into
// dev.Outbound are what the operating system sent into the TUN; packets
// received from dev.Inbound are what the stack wrote back.
func startTestStack(t *testing.T, h Handlers) *tuntest.ChannelTUN {
	t.Helper()
	dev := tuntest.NewChannelTUN()
	s := newMipsStack(1420)
	if err := s.Start(dev.TUN(), h); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		s.Close()
		dev.TUN().Close()
	})
	return dev
}

func recvPacket(t *testing.T, dev *tuntest.ChannelTUN) parsed {
	t.Helper()
	select {
	case p := <-dev.Inbound:
		return parsePacket(p)
	case <-time.After(2 * time.Second):
		t.Fatal("no packet written to the device")
	}
	return parsed{}
}

func TestTCPSynReachesHandlerAndIsAnswered(t *testing.T) {
	for _, tc := range []struct{ name, src, dst string }{
		{"v4", "10.0.85.2:40000", "1.1.1.1:443"},
		{"v6", "[fdfe:dcba:9876::2]:40000", "[2606:4700:4700::1111]:443"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, dst := netip.MustParseAddrPort(tc.src), netip.MustParseAddrPort(tc.dst)
			got := make(chan Flow, 1)
			conns := make(chan net.Conn, 1)
			dev := startTestStack(t, Handlers{TCP: func(f Flow, c net.Conn) { got <- f; conns <- c }})

			dev.Outbound <- tcpSyn(src, dst)

			p := recvPacket(t, dev)
			if p.proto != 6 || p.tcpFlags&0x12 != 0x12 {
				t.Fatalf("expected SYN-ACK on the device, got proto %d flags %#x", p.proto, p.tcpFlags)
			}
			if p.src != dst || p.dst != src {
				t.Fatalf("SYN-ACK addressed %v -> %v, want %v -> %v", p.src, p.dst, dst, src)
			}
			// Accept returns, and the handler runs, once the handshake
			// completes with the application's ACK.
			dev.Outbound <- tcpAck(src, dst, p.tcpSeq)
			select {
			case f := <-got:
				if f.Source != src || f.Destination != dst {
					t.Fatalf("flow %+v, want %v -> %v", f, src, dst)
				}
				(<-conns).Close()
			case <-time.After(2 * time.Second):
				t.Fatal("TCP handler not called after the handshake")
			}
		})
	}
}

func TestUDPDatagramReachesHandlerAndReplyFromSpoofsSource(t *testing.T) {
	for _, tc := range []struct{ name, src, dst, third string }{
		{"v4", "10.0.85.2:40000", "8.8.8.8:53", "9.9.9.9:53"},
		{"v6", "[fdfe:dcba:9876::2]:40000", "[2001:4860:4860::8888]:53", "[2620:fe::fe]:53"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, dst, third := netip.MustParseAddrPort(tc.src), netip.MustParseAddrPort(tc.dst), netip.MustParseAddrPort(tc.third)
			type seen struct {
				flow    Flow
				payload []byte
				reply   func([]byte, netip.AddrPort) error
			}
			got := make(chan seen, 1)
			dev := startTestStack(t, Handlers{UDP: func(f Flow, payload []byte, reply func([]byte, netip.AddrPort) error) {
				got <- seen{f, bytes.Clone(payload), reply}
			}})

			dev.Outbound <- udpDatagram(src, dst, []byte("query"))

			var s seen
			select {
			case s = <-got:
			case <-time.After(2 * time.Second):
				t.Fatal("UDP handler not called")
			}
			if s.flow.Source != src || s.flow.Destination != dst || string(s.payload) != "query" {
				t.Fatalf("got flow %+v payload %q", s.flow, s.payload)
			}

			// A reply from a different remote must leave the device with
			// that remote as its source: this is what full-cone NAT needs.
			if err := s.reply([]byte("answer"), third); err != nil {
				t.Fatalf("reply: %v", err)
			}
			p := recvPacket(t, dev)
			if p.proto != 17 || p.src != third || p.dst != src || string(p.payload) != "answer" {
				t.Fatalf("reply packet %+v payload %q, want %v -> %v", p, p.payload, third, src)
			}
		})
	}
}

// TestNoStackNoDelivery is the failure-first check: without Start nothing
// consumes the device, so the packet is never read.
func TestNoStackNoDelivery(t *testing.T) {
	dev := tuntest.NewChannelTUN()
	select {
	case dev.Outbound <- tcpSyn(netip.MustParseAddrPort("10.0.85.2:1"), netip.MustParseAddrPort("1.1.1.1:443")):
		t.Fatal("packet consumed without a stack")
	case <-time.After(100 * time.Millisecond):
	}
}
