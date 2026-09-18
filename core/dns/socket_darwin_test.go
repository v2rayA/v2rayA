//go:build darwin

package dns

import (
	"net"
	"testing"

	"golang.org/x/sys/unix"
)

// The dialer's sockets must carry IP_BOUND_IF for the configured interface
// and nothing when none is configured.
func TestMarkedDialerBindsEgressInterface(t *testing.T) {
	t.Cleanup(func() { SetEgressInterface("") })
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()

	boundIf := func() int {
		conn, err := markedDialer().Dial("udp4", pc.LocalAddr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		raw, err := conn.(*net.UDPConn).SyscallConn()
		if err != nil {
			t.Fatal(err)
		}
		var idx int
		raw.Control(func(fd uintptr) { idx, _ = unix.GetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF) })
		return idx
	}

	SetEgressInterface("")
	if idx := boundIf(); idx != 0 {
		t.Fatalf("unconfigured dialer bound to interface %d", idx)
	}
	lo := loopbackInterface(t)
	SetEgressInterface(lo.Name)
	if idx := boundIf(); idx != lo.Index {
		t.Fatalf("bound to %d, want %d", idx, lo.Index)
	}
}
