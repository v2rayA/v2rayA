//go:build darwin || windows

package httpClient

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

func DirectDialer(timeout time.Duration, bypass bool) *net.Dialer {
	setting := configure.GetSettingNotNil()
	if !bypass || setting.TransparentType != configure.TransparentTun || !v2ray.IsTransparentOn(setting) {
		return &net.Dialer{Timeout: timeout}
	}
	return boundDirectDialer(timeout, v2ray.TunEgressInterfaceIfTun(setting))
}

func boundDirectDialer(timeout time.Duration, name string) *net.Dialer {
	iface, err := net.InterfaceByName(name)
	dialer := &net.Dialer{Timeout: timeout, Control: func(network, _ string, conn syscall.RawConn) error {
		// An unbound fallback would probe the local TUN stack, which can
		// accept TCP connections even when the remote server is down.
		if err != nil {
			return fmt.Errorf("direct probe: cannot bind to egress interface %q: %w", name, err)
		}
		var bindErr error
		if err := conn.Control(func(fd uintptr) {
			bindErr = bindProbeSocket(fd, network, iface.Index)
		}); err != nil {
			return err
		}
		return bindErr
	}}
	dialer.Resolver = resolv.DirectResolver(dialer)
	return dialer
}
