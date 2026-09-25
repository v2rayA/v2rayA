//go:build darwin || windows

package httpClient

import (
	"context"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/common/resolv"
)

func TestBoundDirectDialerRejectsMissingInterface(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	for _, name := range []string{"", "v2raya-missing-test-interface"} {
		conn, err := boundDirectDialer(time.Second, name).Dial("tcp4", listener.Addr().String())
		if conn != nil {
			conn.Close()
		}
		if err == nil || !strings.Contains(err.Error(), "cannot bind to egress interface") {
			t.Fatalf("interface %q: conn=%v, err=%v", name, conn, err)
		}
	}
}

func TestBoundDirectDialerSocketOptions(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	var loopback *net.Interface
	for i := range interfaces {
		if interfaces[i].Flags&net.FlagLoopback != 0 {
			loopback = &interfaces[i]
			break
		}
	}
	if loopback == nil {
		t.Fatal("no loopback interface")
	}
	for _, tc := range []struct {
		network, address string
	}{
		{"tcp4", "127.0.0.1:0"},
		{"udp4", "127.0.0.1:0"},
		{"tcp6", "[::1]:0"},
		{"udp6", "[::1]:0"},
	} {
		t.Run(tc.network, func(t *testing.T) {
			address := tc.address
			if tc.network[:3] == "tcp" {
				listener, err := net.Listen(tc.network, address)
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				address = listener.Addr().String()
			} else {
				listener, err := net.ListenPacket(tc.network, address)
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				address = listener.LocalAddr().String()
			}
			dialer := boundDirectDialer(time.Second, loopback.Name)
			control := dialer.Control
			checked := false
			dialer.Control = func(network, address string, raw syscall.RawConn) error {
				if err := control(network, address, raw); err != nil {
					return err
				}
				return raw.Control(func(fd uintptr) {
					index, err := boundProbeInterface(fd, network)
					if err != nil || index != loopback.Index {
						t.Errorf("bound index=%d, err=%v, want %d", index, err, loopback.Index)
					}
					checked = true
				})
			}
			conn, err := dialer.DialContext(context.Background(), tc.network, address)
			if err != nil {
				t.Fatal(err)
			}
			conn.Close()
			if !checked {
				t.Fatal("socket binding was not checked")
			}
		})
	}
}

func TestBoundDirectDialerDNSFailsClosed(t *testing.T) {
	_, err := resolv.LookupHostWithDialer("bind-test.invalid", boundDirectDialer(time.Second, ""))
	if err == nil {
		t.Fatal("DNS lookup without an egress interface unexpectedly succeeded")
	}
}
