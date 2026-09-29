package configure

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"
)

// A node whose server speaks QUIC (juicity, tuic, hysteria2) listens on UDP
// only, so the TCP dial in Ping fails for a server that is there. A QUIC server
// answers a long header packet carrying a version it does not implement with a
// Version Negotiation packet (RFC 9000 section 6), and that answer is a round
// trip to the node.
const (
	// quicProbeVersion is one of the reserved "grease" versions (0x?a?a?a?a):
	// no implementation implements it, which is what makes a server answer
	// instead of dropping the packet.
	quicProbeVersion = 0x1a2a3a4a
	// quicProbeSize is the smallest datagram a QUIC server accepts a packet
	// with an unknown version to have.
	quicProbeSize = 1200
	// quicProbeConnIDLen is the destination connection ID length the probe
	// carries. Its bytes are left zero: the server only echoes them.
	quicProbeConnIDLen = 8
)

// pingUDP measures a round trip to a QUIC server at host:port. It fails when no
// Version Negotiation comes back within timeout, which is also what a server
// that is not there, or is not QUIC, does. The datagram leaves through the
// probe's dialer: an unmarked, unbound socket would enter the transparent
// interception the probe exists to stay out of, and measure the local stack.
func pingUDP(dialer *net.Dialer, host string, port int, timeout time.Duration) (time.Duration, error) {
	conn, err := dialer.Dial("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	started := time.Now()
	if _, err = conn.Write(quicProbePacket()); err != nil {
		return 0, err
	}
	reply := make([]byte, 512)
	for {
		n, err := conn.Read(reply)
		if err != nil {
			return 0, err
		}
		if isVersionNegotiation(reply[:n]) {
			return time.Since(started), nil
		}
		// The socket is connected, so anything else comes from the server's
		// port too but is not the answer; keep the deadline and read again.
	}
}

// quicProbePacket builds the long header probe. A packet carrying a version we
// do not implement carries only the fields RFC 8999 section 5.1 fixes: the
// header form and the fixed bit, the version, the connection ID lengths and the
// IDs. The rest is padding to quicProbeSize.
func quicProbePacket() []byte {
	packet := make([]byte, quicProbeSize)
	packet[0] = 0xc0 // long header, fixed bit, Initial type
	binary.BigEndian.PutUint32(packet[1:5], quicProbeVersion)
	packet[5] = quicProbeConnIDLen
	packet[6+quicProbeConnIDLen] = 0 // source connection ID length
	return packet
}

// isVersionNegotiation reports a packet in the Version Negotiation format: a
// long header whose version field is zero.
func isVersionNegotiation(packet []byte) bool {
	return len(packet) >= 5 && packet[0]&0x80 != 0 && binary.BigEndian.Uint32(packet[1:5]) == 0
}
