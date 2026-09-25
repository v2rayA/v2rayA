//go:build linux

package httpClient

import (
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
	dialer.Resolver = resolv.DirectResolver(dialer)
	return dialer
}
