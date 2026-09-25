package httpClient

import (
	"context"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"golang.org/x/sys/unix"
)

// Service-owned probes and subscription downloads must not depend on the
// group they are repairing. This dialer is never used for forwarded traffic.
func DirectDialer(timeout time.Duration) *net.Dialer {
	if !v2ray.IsTransparentOn(configure.GetSettingNotNil()) {
		return &net.Dialer{Timeout: timeout}
	}
	dialer := &net.Dialer{Timeout: timeout, Control: func(_, _ string, c syscall.RawConn) error {
		var markErr error
		if err := c.Control(func(fd uintptr) { markErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, 0x80) }); err != nil {
			return err
		}
		return markErr
	}}
	dnsDialer := *dialer
	dialer.Resolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		if preferred := resolv.PreferredServers(); len(preferred) > 0 {
			address = preferred[0]
		}
		return dnsDialer.DialContext(ctx, network, address)
	}}
	return dialer
}

func DirectSubscriptionClient() *http.Client {
	if !v2ray.IsTransparentOn(configure.GetSettingNotNil()) {
		return http.DefaultClient
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = DirectDialer(10 * time.Second).DialContext
	// Do not accumulate idle transports on a long-running router.
	transport.DisableKeepAlives = true
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}
}
