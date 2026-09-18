package dns

import (
	"net"
	"testing"
)

func loopbackInterface(t *testing.T) net.Interface {
	t.Helper()
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagLoopback != 0 {
			return i
		}
	}
	t.Skip("no loopback interface")
	return net.Interface{}
}

func TestEgressInterfaceIndex(t *testing.T) {
	t.Cleanup(func() { SetEgressInterface("") })

	SetEgressInterface("")
	if _, ok := egressInterfaceIndex(); ok {
		t.Fatal("unset interface resolved")
	}
	SetEgressInterface("no-such-interface-0")
	if _, ok := egressInterfaceIndex(); ok {
		t.Fatal("missing interface resolved")
	}
	lo := loopbackInterface(t)
	SetEgressInterface(lo.Name)
	if idx, ok := egressInterfaceIndex(); !ok || idx != lo.Index {
		t.Fatalf("got %d %v, want %d", idx, ok, lo.Index)
	}
}

func TestIsIPv6Address(t *testing.T) {
	for _, tc := range []struct {
		network, address string
		want             bool
	}{
		{"udp", "8.8.8.8:53", false},
		{"udp", "[2001:4860:4860::8888]:53", true},
		{"tcp4", "[::ffff:1.1.1.1]:53", false},
		{"udp6", "[fe80::1%en0]:53", true},
		{"tcp", "dns.google:853", false},
		{"tcp6", "dns.google:853", true},
	} {
		if got := isIPv6Address(tc.network, tc.address); got != tc.want {
			t.Errorf("isIPv6Address(%q, %q) = %v, want %v", tc.network, tc.address, got, tc.want)
		}
	}
}
