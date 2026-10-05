//go:build windows

package dns

import (
	"encoding/binary"
	"net"
	"syscall"
	"time"
	"unsafe"

	"github.com/miekg/dns"
)

// From ws2ipdef.h; the two options share the number.
const (
	ipUnicastIf   = 31
	ipv6UnicastIf = 31
)

// setSocketMark is a no-op on Windows: there is no SO_MARK. Self-exclusion
// from the TUN is done by binding to the physical interface instead.
func setSocketMark(fd uintptr) error {
	return nil
}

// markFd binds the socket to the configured egress interface, if any.
// IP_UNICAST_IF wants the index in network byte order; IPV6_UNICAST_IF
// wants host order.
func markFd(network, address string, c syscall.RawConn) error {
	if isLocalDNSAddress(address) {
		return nil
	}
	idx, ok := egressInterfaceIndex()
	if !ok {
		return nil
	}
	var opErr error
	err := c.Control(func(fd uintptr) {
		if isIPv6Address(network, address) {
			opErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IPV6, ipv6UnicastIf, idx)
			return
		}
		var be [4]byte
		binary.BigEndian.PutUint32(be[:], uint32(idx))
		opErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, ipUnicastIf, int(*(*uint32)(unsafe.Pointer(&be[0]))))
	})
	if err != nil {
		return err
	}
	return opErr
}

func markedDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   5 * time.Second,
		Control:   markFd,
		KeepAlive: 30 * time.Second,
	}
}

// newMarkedDnsClient creates a *dns.Client whose sockets are bound to the
// egress interface when one is configured. UDPSize=4096 gives queries
// without an EDNS0 OPT record a 4 KiB receive buffer instead of
// miekg/dns's 512-byte default.
func newMarkedDnsClient(network string) *dns.Client {
	return &dns.Client{
		Net:          network,
		UDPSize:      4096,
		Timeout:      5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Dialer:       markedDialer(),
	}
}
