//go:build linux

package httpClient

import (
	"context"
	"net"
	"syscall"
	"time"

	"github.com/v2rayA/v2rayA/common/resolv"
	"golang.org/x/sys/unix"
)

// DirectDialer returns a dialer whose sockets bypass v2rayA's transparent
// interception when mark is true. It is intended for service-owned health
// checks, never for forwarded user traffic.
func DirectDialer(timeout time.Duration, mark bool) *net.Dialer {
	dialer := &net.Dialer{Timeout: timeout}
	if !mark {
		return dialer
	}
	dialer.Control = func(_, _ string, c syscall.RawConn) error {
		var markErr error
		if err := c.Control(func(fd uintptr) {
			markErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, 0x80)
		}); err != nil {
			return err
		}
		return markErr
	}
	// DNS must use the same mark, otherwise resolving a node can still enter
	// the transparent proxy that the probe is checking.
	dnsDialer := *dialer
	dialer.Resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			if preferred := resolv.PreferredServers(); len(preferred) > 0 {
				address = preferred[0]
			}
			return dnsDialer.DialContext(ctx, network, address)
		},
	}
	return dialer
}
