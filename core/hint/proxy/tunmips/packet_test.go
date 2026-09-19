// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"encoding/binary"
	"net/netip"
)

// Test-only packet builders. They produce minimal, checksum-correct IPv4/IPv6
// TCP and UDP packets so tests can drive the stack through a channel device.

func csum(sum uint32, b []byte) uint32 {
	for i := 0; i+1 < len(b); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(b[i:]))
	}
	if len(b)%2 == 1 {
		sum += uint32(b[len(b)-1]) << 8
	}
	return sum
}

func fold(sum uint32) uint16 {
	for sum>>16 != 0 {
		sum = sum&0xffff + sum>>16
	}
	return ^uint16(sum)
}

func pseudoHeader(src, dst netip.Addr, proto uint8, length int) []byte {
	if src.Is4() {
		b := make([]byte, 12)
		s, d := src.As4(), dst.As4()
		copy(b[0:], s[:])
		copy(b[4:], d[:])
		b[9] = proto
		binary.BigEndian.PutUint16(b[10:], uint16(length))
		return b
	}
	b := make([]byte, 40)
	s, d := src.As16(), dst.As16()
	copy(b[0:], s[:])
	copy(b[16:], d[:])
	binary.BigEndian.PutUint32(b[32:], uint32(length))
	b[39] = proto
	return b
}

func ipHeader(src, dst netip.Addr, proto uint8, payloadLen int) []byte {
	if src.Is4() {
		h := make([]byte, 20)
		h[0] = 0x45
		binary.BigEndian.PutUint16(h[2:], uint16(20+payloadLen))
		h[8] = 64
		h[9] = proto
		s, d := src.As4(), dst.As4()
		copy(h[12:], s[:])
		copy(h[16:], d[:])
		binary.BigEndian.PutUint16(h[10:], fold(csum(0, h)))
		return h
	}
	h := make([]byte, 40)
	h[0] = 0x60
	binary.BigEndian.PutUint16(h[4:], uint16(payloadLen))
	h[6] = proto
	h[7] = 64
	s, d := src.As16(), dst.As16()
	copy(h[8:], s[:])
	copy(h[24:], d[:])
	return h
}

// tcpSyn builds a SYN from src to dst with initial sequence number 1000.
func tcpSyn(src, dst netip.AddrPort) []byte {
	return tcpSegment(src, dst, 1000, 0, 0x02)
}

// tcpAck builds the third handshake segment acknowledging the peer's ISN.
func tcpAck(src, dst netip.AddrPort, peerSeq uint32) []byte {
	return tcpSegment(src, dst, 1001, peerSeq+1, 0x10)
}

func tcpSegment(src, dst netip.AddrPort, seq, ack uint32, flags uint8) []byte {
	t := make([]byte, 20)
	binary.BigEndian.PutUint16(t[0:], src.Port())
	binary.BigEndian.PutUint16(t[2:], dst.Port())
	binary.BigEndian.PutUint32(t[4:], seq)
	binary.BigEndian.PutUint32(t[8:], ack)
	t[12] = 5 << 4
	t[13] = flags
	binary.BigEndian.PutUint16(t[14:], 65535)
	binary.BigEndian.PutUint16(t[16:], fold(csum(csum(0, pseudoHeader(src.Addr(), dst.Addr(), 6, len(t))), t)))
	return append(ipHeader(src.Addr(), dst.Addr(), 6, len(t)), t...)
}

// udpDatagram builds one UDP datagram from src to dst.
func udpDatagram(src, dst netip.AddrPort, payload []byte) []byte {
	u := make([]byte, 8+len(payload))
	binary.BigEndian.PutUint16(u[0:], src.Port())
	binary.BigEndian.PutUint16(u[2:], dst.Port())
	binary.BigEndian.PutUint16(u[4:], uint16(len(u)))
	copy(u[8:], payload)
	c := fold(csum(csum(0, pseudoHeader(src.Addr(), dst.Addr(), 17, len(u))), u))
	if c == 0 {
		c = 0xffff
	}
	binary.BigEndian.PutUint16(u[6:], c)
	return append(ipHeader(src.Addr(), dst.Addr(), 17, len(u)), u...)
}

// parsed is the little a test needs to know about a packet the stack wrote.
type parsed struct {
	proto    uint8
	src, dst netip.AddrPort
	tcpFlags uint8
	tcpSeq   uint32
	payload  []byte
}

func parsePacket(p []byte) parsed {
	var r parsed
	var l4 []byte
	switch p[0] >> 4 {
	case 4:
		hl := int(p[0]&0x0f) * 4
		r.proto = p[9]
		src := netip.AddrFrom4([4]byte(p[12:16]))
		dst := netip.AddrFrom4([4]byte(p[16:20]))
		l4 = p[hl:]
		r.src, r.dst = netip.AddrPortFrom(src, binary.BigEndian.Uint16(l4[0:])), netip.AddrPortFrom(dst, binary.BigEndian.Uint16(l4[2:]))
	case 6:
		r.proto = p[6]
		src := netip.AddrFrom16([16]byte(p[8:24]))
		dst := netip.AddrFrom16([16]byte(p[24:40]))
		l4 = p[40:]
		r.src, r.dst = netip.AddrPortFrom(src, binary.BigEndian.Uint16(l4[0:])), netip.AddrPortFrom(dst, binary.BigEndian.Uint16(l4[2:]))
	}
	switch r.proto {
	case 6:
		r.tcpFlags = l4[13]
		r.tcpSeq = binary.BigEndian.Uint32(l4[4:])
	case 17:
		r.payload = l4[8:]
	}
	return r
}
