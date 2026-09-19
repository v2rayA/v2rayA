// Package tunmips is the in-core TUN inbound: a wireguard/tun device driven
// by a user-space IP stack whose TCP connections and UDP datagrams are handed
// to xray's dispatcher.
//
// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"net"
	"net/netip"

	"golang.zx2c4.com/wireguard/tun"
)

// packetOffset is the number of bytes reserved in front of every packet
// buffer exchanged with the device. The Linux device needs at least the
// virtio-net header (10 bytes) in front of each packet it writes with GSO
// enabled; 16 keeps the payload 4-byte aligned like wireguard-go does.
const packetOffset = 16

// maxPacketSize bounds one IP packet including any GSO super-packet the
// device may hand us.
const maxPacketSize = 65535

// Flow identifies one intercepted TCP connection or UDP datagram as seen from
// the application: Source is the application's socket, Destination is the
// address it dialed.
type Flow struct {
	Source      netip.AddrPort
	Destination netip.AddrPort
}

// Handlers is the stack-independent contract between the IP stack and the
// inbound. A stack implementation calls these; nothing else in the package
// depends on which stack is behind them.
type Handlers struct {
	// TCP is called once per accepted connection, on its own goroutine. The
	// handshake has already completed; closing conn ends the connection.
	TCP func(flow Flow, conn net.Conn)
	// UDP is called once per datagram, synchronously from the stack's input
	// path, so it must not block. payload is only valid during the call.
	// reply writes one datagram back to flow.Source using from as its source
	// address; it does not depend on any endpoint and may be used after UDP
	// returns.
	UDP func(flow Flow, payload []byte, reply func(payload []byte, from netip.AddrPort) error)
}

// netStack moves packets between a device and Handlers.
type netStack interface {
	Start(dev tun.Device, h Handlers) error
	Close() error
}
