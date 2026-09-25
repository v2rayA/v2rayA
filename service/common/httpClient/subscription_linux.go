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

// Only subscription downloads use this exception. User traffic assigned to
// an empty group remains blocked; refreshing its provider must not depend on it.
func DirectSubscriptionClient() *http.Client {
	if !v2ray.IsTransparentOn(configure.GetSettingNotNil()) {
		return http.DefaultClient
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, Control: func(_, _ string, c syscall.RawConn) error {
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
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = dialer.DialContext
	// Do not accumulate idle transports on a long-running router.
	transport.DisableKeepAlives = true
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}
}
