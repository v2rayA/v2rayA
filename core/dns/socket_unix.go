//go:build !windows

package dns

import (
	"net"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// markFd returns a net.Dialer.Control function that sets SO_MARK=0x80 on
// every socket created by the dialer. 0x80 is the fwmark checked by the
// v2raya nftables/iptables DNS redirect chains (mark & 0x80 == 0x80 → RETURN),
// which prevents the DNS module's own upstream queries from being hijacked
// back into itself (redirect loop → unbounded memory growth).
func markFd(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		_ = setSocketMark(fd) // SO_MARK=36
	})
}

// markedDialer returns a *net.Dialer that sets SO_MARK=0x80 on all sockets.
func markedDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   5 * time.Second,
		Control:   markFd,
		KeepAlive: 30 * time.Second,
	}
}

// newMarkedDnsClient creates a *dns.Client whose every socket carries the
// SO_MARK=0x80 mark, so upstream DNS queries bypass the transparent-proxy
// redirect rules. Use this for ALL DNS upstream clients; creating a client
// without the mark (e.g. a bare &dns.Client{}) makes its queries loop back
// into the DNS module and exhaust memory.
//
// Note: miekg/dns dials through net.Dialer for UDP as well, so the marked
// Dialer covers every transport this client supports — no manual socket
// path is needed. UDPSize is set to 4096 so queries without an EDNS0 OPT
// record still get a 4 KiB receive buffer instead of miekg/dns's 512-byte
// default (queries built with SetEdns0 override this via their OPT record).
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

// setSocketMark sets SO_MARK on a socket identified by its file descriptor.
// SO_MARK (option 36) is used for iptables/nftables mark-based filtering
// to prevent DNS query loops (mark 0x80 → iptables RETURN).
// This is Linux-specific; on other Unix platforms it compiles but is a no-op
// at the syscall level (option 36 has a different meaning).
func setSocketMark(fd uintptr) error {
	return syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, 36, 0x80)
}
