//go:build linux

package bounddevice

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

func TestProgramAssembly(t *testing.T) {
	for _, attach := range []ebpf.AttachType{ebpf.AttachCGroupInet4Connect, ebpf.AttachCGroupInet6Connect} {
		spec := programSpec(attach, 0x123456789abcdef0)
		for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
			var encoded bytes.Buffer
			if err := spec.Instructions.Marshal(&encoded, order); err != nil {
				t.Fatal(err)
			}
			if encoded.Len()%8 != 0 {
				t.Fatal("invalid BPF instruction alignment")
			}
		}
	}
}

func TestRejectNonCgroupMount(t *testing.T) {
	if _, err := start(t.TempDir(), 1); err == nil {
		t.Fatal("a regular directory must not be accepted as a cgroup mount")
	}
}

// Opt-in privileged test. The guard is attached ONLY to a new, disposable
// cgroup, never to the host root cgroup. Child sockets connect to loopback only.
// No nftables rules, external connections, or production configuration change.
func TestNativeBoundDeviceGuard(t *testing.T) {
	if os.Getenv("V2RAYA_BPF_TEST") != "1" {
		t.Skip("set V2RAYA_BPF_TEST=1 and run as root to verify with the kernel")
	}
	if os.Geteuid() != 0 {
		t.Fatal("native BPF verification needs root")
	}
	cgroup := filepath.Join("/sys/fs/cgroup", fmt.Sprintf("v2raya-binddev-test-%d", os.Getpid()))
	if err := os.Mkdir(cgroup, 0700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(cgroup); err != nil {
			t.Errorf("remove test cgroup: %v", err)
		}
	})
	guard, err := Start(cgroup)
	if err != nil {
		t.Fatalf("BPF load/attach failed: %+v", err)
	}
	t.Cleanup(func() { guard.Close() })
	runChild := func(mode string) {
		t.Helper()
		cmd := exec.Command(os.Args[0], "-test.run=^TestNativeSocketChild$", "-test.v")
		cmd.Env = append(os.Environ(), "V2RAYA_BPF_CHILD="+mode, "V2RAYA_BPF_CHILD_CGROUP="+cgroup)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("socket child (%s): %v\n%s", mode, err, output)
		}
	}
	runChild("guarded")
	if err := guard.Close(); err != nil {
		t.Fatal(err)
	}
	runChild("detached")
	// Model a socket in a different network namespace by attaching a guard
	// expecting a different namespace cookie. Nothing should be marked.
	cookie, err := netnsCookie()
	if err != nil {
		t.Fatal(err)
	}
	otherNS, err := start(cgroup, cookie+1)
	if err != nil {
		t.Fatal(err)
	}
	defer otherNS.Close()
	runChild("other-namespace")
}

func TestNativeSocketChild(t *testing.T) {
	mode := os.Getenv("V2RAYA_BPF_CHILD")
	if mode == "" {
		t.Skip("run by the isolated cgroup test")
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("V2RAYA_BPF_CHILD_CGROUP"), "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		t.Fatal(err)
	}
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		for _, tc := range []struct {
			name       string
			protocol   int
			device     bool
			sourceOnly bool
			mark       int
		}{
			{"ordinary-tcp", unix.IPPROTO_TCP, false, false, 0},
			{"source-bound-tcp", unix.IPPROTO_TCP, false, true, 0},
			{"device-bound-tcp", unix.IPPROTO_TCP, true, false, 0},
			{"preserve-mark", unix.IPPROTO_TCP, true, false, 0x1200},
			{"existing-bypass", unix.IPPROTO_TCP, true, false, 0x1280},
			{"ordinary-mark", unix.IPPROTO_TCP, false, false, 0x1200},
			{"device-bound-udp", unix.IPPROTO_UDP, true, false, 0},
		} {
			t.Run(fmt.Sprintf("%d/%s", family, tc.name), func(t *testing.T) {
				sockType := unix.SOCK_STREAM
				if tc.protocol == unix.IPPROTO_UDP {
					sockType = unix.SOCK_DGRAM
				}
				fd, err := unix.Socket(family, sockType|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, tc.protocol)
				if err != nil {
					t.Fatal(err)
				}
				defer unix.Close(fd)
				if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_MARK, tc.mark); err != nil {
					t.Fatal(err)
				}
				if tc.device {
					if err := unix.SetsockoptString(fd, unix.SOL_SOCKET, unix.SO_BINDTODEVICE, "lo"); err != nil {
						t.Fatal(err)
					}
				}
				var target unix.Sockaddr = &unix.SockaddrInet4{Port: 1, Addr: [4]byte{127, 0, 0, 1}}
				var source unix.Sockaddr = &unix.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}
				if family == unix.AF_INET6 {
					addr := [16]byte{15: 1}
					target = &unix.SockaddrInet6{Port: 1, Addr: addr}
					source = &unix.SockaddrInet6{Addr: addr}
				}
				if tc.sourceOnly {
					if err := unix.Bind(fd, source); err != nil {
						t.Fatal(err)
					}
				}
				if err := unix.Connect(fd, target); err != nil && !errors.Is(err, unix.EINPROGRESS) && !errors.Is(err, unix.ECONNREFUSED) {
					t.Fatalf("connect: %v", err)
				}
				mark, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_MARK)
				if err != nil {
					t.Fatal(err)
				}
				want := tc.mark
				if mode == "guarded" && tc.device && tc.protocol == unix.IPPROTO_TCP {
					want |= BypassMark
				}
				if mark != want {
					t.Fatalf("SO_MARK=%#x, want %#x", mark, want)
				}
			})
		}
	}
}
