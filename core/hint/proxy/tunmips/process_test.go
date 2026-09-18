// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"errors"
	stdnet "net"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/xtls/xray-core/common/session"
)

func TestExclusionDecision(t *testing.T) {
	src := netip.MustParseAddrPort("10.0.85.2:40000")
	cases := []struct {
		name   string
		owners []owner
		err    error
		want   bool
	}{
		{"listed name", []owner{{pid: 7, name: "curl"}}, nil, true},
		{"listed name with path", []owner{{pid: 7, name: "/usr/bin/curl"}}, nil, true},
		{"self pid under another name", []owner{{pid: 4242, name: "renamed"}}, nil, true},
		{"one of several owners listed", []owner{{pid: 1, name: "sh"}, {pid: 2, name: "curl"}}, nil, true},
		{"self pid among unlisted owners", []owner{{pid: 9, name: "firefox"}, {pid: 4242, name: ""}}, nil, true},
		{"owner without a name", []owner{{pid: 9, name: ""}}, nil, false},
		{"windows name case and suffix", []owner{{pid: 9, name: "CURL.EXE"}}, nil, runtime.GOOS == "windows"},
		{"unlisted", []owner{{pid: 9, name: "firefox"}}, nil, false},
		{"no owner", nil, nil, false},
		{"lookup error", nil, errors.New("permission denied"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newExclusion([]string{"curl", " wget "}, []uint32{4242})
			e.lookup = func(string, netip.AddrPort) ([]owner, error) { return tc.owners, tc.err }
			if got := e.excluded("tcp", src); got != tc.want {
				t.Fatalf("excluded = %v, want %v", got, tc.want)
			}
		})
	}
	// An empty list never looks anything up.
	e := newExclusion(nil, nil)
	e.lookup = func(string, netip.AddrPort) ([]owner, error) { t.Fatal("lookup called"); return nil, nil }
	if e.excluded("tcp", src) {
		t.Fatal("empty exclusion excluded a flow")
	}
}

func TestOwnerOfFindsThisProcess(t *testing.T) {
	l, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			defer c.Close()
		}
	}()
	conn, err := stdnet.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	self := uint32(os.Getpid())
	selfName := filepath.Base(os.Args[0])

	owners, err := ownerOf("tcp", conn.LocalAddr().(*stdnet.TCPAddr).AddrPort())
	if err != nil {
		t.Fatalf("ownerOf tcp: %v", err)
	}
	if !hasOwner(owners, self, selfName) {
		t.Fatalf("tcp owners %+v do not include pid %d %q", owners, self, selfName)
	}

	// A UDP socket bound to the wildcard address is found by port alone,
	// looked up under the address the datagram actually left from.
	pc, err := stdnet.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	port := pc.LocalAddr().(*stdnet.UDPAddr).AddrPort().Port()
	owners, err = ownerOf("udp", netip.AddrPortFrom(netip.MustParseAddr("10.0.85.2"), port))
	if err != nil {
		t.Fatalf("ownerOf udp: %v", err)
	}
	if !hasOwner(owners, self, selfName) {
		t.Fatalf("udp owners %+v do not include pid %d", owners, self)
	}

	// A dual-stack socket keeps an IPv4 peer as a v4-mapped address; the
	// lookup is by the IPv4 flow the packet shows.
	if runtime.GOOS != "windows" {
		if c6, err := stdnet.Dial("tcp6", "[::ffff:127.0.0.1]:"+itoa(l.Addr().(*stdnet.TCPAddr).Port)); err == nil {
			defer c6.Close()
			ap := c6.LocalAddr().(*stdnet.TCPAddr).AddrPort()
			owners, err = ownerOf("tcp", netip.AddrPortFrom(ap.Addr().Unmap(), ap.Port()))
			if err != nil || !hasOwner(owners, self, selfName) {
				t.Fatalf("v4-mapped socket: owners %+v err %v", owners, err)
			}
		}
	}

	// An unused port has no owner and is not an error.
	owners, err = ownerOf("tcp", netip.MustParseAddrPort("127.0.0.1:1"))
	if err != nil || len(owners) != 0 {
		t.Fatalf("unused port: owners %+v err %v", owners, err)
	}
}

func hasOwner(owners []owner, pid uint32, name string) bool {
	for _, o := range owners {
		if o.pid == pid && o.name == name {
			return true
		}
	}
	return false
}

func TestExcludedFlowIsForcedToDirect(t *testing.T) {
	d := newFakeDispatcher()
	f := newForwarder(context.Background(), d)
	f.directTag = "direct"
	f.excl = newExclusion([]string{"curl"}, nil)
	calls := 0
	f.excl.lookup = func(proto string, src netip.AddrPort) ([]owner, error) {
		calls++
		if src.Port() == 40000 {
			return []owner{{pid: 1, name: "curl"}}, nil
		}
		return []owner{{pid: 2, name: "firefox"}}, nil
	}
	h := f.handlers()

	for _, tc := range []struct {
		port uint16
		want string
	}{{40000, "direct"}, {40001, ""}} {
		_, stackSide := stdnet.Pipe()
		go h.TCP(Flow{Source: netip.AddrPortFrom(netip.MustParseAddr("10.0.85.2"), tc.port), Destination: netip.MustParseAddrPort("1.1.1.1:443")}, stackSide)
		call := d.next(t)
		if got := session.GetForcedOutboundTagFromContext(call.ctx); got != tc.want {
			t.Fatalf("port %d: forced outbound %q, want %q", tc.port, got, tc.want)
		}
	}
	if calls != 2 {
		t.Fatalf("owner looked up %d times, want once per flow", calls)
	}
	close(d.hold)
}
