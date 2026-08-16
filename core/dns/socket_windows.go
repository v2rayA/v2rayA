//go:build windows

package dns

import (
	"net"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// setSocketMark is a no-op on Windows since SO_MARK is a Linux-specific
// socket option used for iptables/nftables mark-based filtering.
func setSocketMark(fd uintptr) error {
	return nil
}

// markFd is a no-op Control function on Windows (SO_MARK unsupported).
func markFd(network, address string, c syscall.RawConn) error {
	return nil
}

// markedDialer returns a plain dialer on Windows (no socket marking).
func markedDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
}

// newMarkedDnsClient creates a *dns.Client. On Windows the mark is a no-op;
// on Linux it sets SO_MARK=0x80 to bypass transparent-proxy DNS redirects.
// UDPSize=4096 gives queries without an EDNS0 OPT record a 4 KiB receive
// buffer instead of miekg/dns's 512-byte default.
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
