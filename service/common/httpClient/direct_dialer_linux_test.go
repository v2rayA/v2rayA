//go:build linux

package httpClient

import (
	"errors"
	"net"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/common/resolv"
	"golang.org/x/sys/unix"
)

func TestDirectDialerMarksOnlyRequestedConnections(t *testing.T) {
	plain := DirectDialer(time.Second, false)
	if plain.Control != nil || plain.Resolver != nil {
		t.Fatal("plain dialer unexpectedly changes the socket or resolver")
	}
	marked := DirectDialer(2*time.Second, true)
	if marked.Control == nil || marked.Resolver == nil {
		t.Fatal("marked dialer must set a socket policy for TCP and DNS")
	}
	if marked.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", marked.Timeout)
	}
	for _, tc := range []struct {
		name   string
		dialer *net.Dialer
		want   int
	}{{"plain", plain, 0}, {"marked", marked, 0x80}} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			control := tc.dialer.Control
			tc.dialer.Control = func(network, address string, conn syscall.RawConn) error {
				if control != nil {
					if err := control(network, address, conn); err != nil {
						return err
					}
				}
				var got int
				var opErr error
				if err := conn.Control(func(fd uintptr) {
					got, opErr = unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK)
				}); err != nil {
					return err
				}
				if opErr != nil {
					return opErr
				}
				if got != tc.want {
					t.Errorf("SO_MARK = %#x, want %#x", got, tc.want)
				}
				return nil
			}
			conn, err := tc.dialer.Dial("tcp4", listener.Addr().String())
			if errors.Is(err, unix.EPERM) {
				t.Skip("SO_MARK requires CAP_NET_ADMIN or CAP_NET_RAW")
			}
			if err != nil {
				t.Fatal(err)
			}
			conn.Close()
		})
	}
}

func TestDirectDialerMarksDNSConnections(t *testing.T) {
	dialer := DirectDialer(time.Second, true)
	control := dialer.Control
	var checked, denied atomic.Bool
	stop := errors.New("stop before sending DNS")
	dialer.Control = func(network, address string, conn syscall.RawConn) error {
		if err := control(network, address, conn); err != nil {
			if errors.Is(err, unix.EPERM) {
				denied.Store(true)
			}
			return err
		}
		var got int
		var opErr error
		if err := conn.Control(func(fd uintptr) {
			got, opErr = unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK)
		}); err != nil {
			return err
		}
		if opErr != nil {
			return opErr
		}
		if got != 0x80 {
			t.Errorf("DNS SO_MARK = %#x, want 0x80", got)
		}
		checked.Store(true)
		return stop
	}
	_, err := resolv.LookupHostWithDialer("mark-test.invalid", dialer)
	if denied.Load() {
		t.Skip("SO_MARK requires CAP_NET_ADMIN or CAP_NET_RAW")
	}
	if err == nil || !checked.Load() {
		t.Fatalf("DNS socket was not checked: checked=%v, err=%v", checked.Load(), err)
	}
}
