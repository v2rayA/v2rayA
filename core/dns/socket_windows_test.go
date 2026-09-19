//go:build windows

package dns

import (
	"encoding/binary"
	"net"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The dialer's sockets must carry IP_UNICAST_IF for the configured
// interface and nothing when none is configured.
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
		// The option is set in network byte order, but getsockopt hands
		// the index back in host order.
		var buf [4]byte
		raw.Control(func(fd uintptr) {
			n := int32(len(buf))
			_ = windows.Getsockopt(windows.Handle(fd), syscall.IPPROTO_IP, ipUnicastIf, (*byte)(unsafe.Pointer(&buf[0])), &n)
		})
		return int(binary.LittleEndian.Uint32(buf[:]))
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
