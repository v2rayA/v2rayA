//go:build darwin

package dns

import (
	"net"
	"syscall"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/sys/unix"
)

// setSocketMark is a no-op on macOS: there is no SO_MARK. Self-exclusion
// from the TUN is done by binding to the physical interface instead.
func setSocketMark(fd uintptr) error {
	return nil
}

// markFd binds the socket to the configured egress interface, if any.
func markFd(network, address string, c syscall.RawConn) error {
	idx, ok := egressInterfaceIndex()
	if !ok {
		return nil
	}
	var opErr error
	err := c.Control(func(fd uintptr) {
		if isIPv6Address(network, address) {
			opErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF, idx)
		} else {
			opErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, idx)
		}
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
