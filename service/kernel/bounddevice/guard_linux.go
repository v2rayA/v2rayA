//go:build linux

// Package bounddevice preserves explicitly device-bound TCP connects before
// REDIRECT changes their destination to a local proxy listener.
package bounddevice

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/cilium/ebpf/link"
	"golang.org/x/sys/unix"
)

// BypassMark is the bit already exempted by v2rayA's REDIRECT and DNS rules.
// Other SO_MARK bits are preserved.
const BypassMark = 0x80

// Stable Linux UAPI offsets, NOT offsets into the kernel's struct sock.
// See include/uapi/linux/bpf.h: bpf_sock_addr and bpf_sock.
const (
	addrProtocolOffset = 36
	addrSocketOffset   = 64
	sockBoundDevOffset = 0
	sockMarkOffset     = 16
)

type Guard struct {
	mu    sync.Mutex
	links []io.Closer
}

// Start attaches to the supplied cgroup v2 subtree. Only sockets in this
// process's network namespace are affected, even if the subtree contains
// containers. Real BPF links are required: closing the process's file
// descriptors must also detach the programs after a crash. Nothing is pinned.
func Start(cgroup string) (*Guard, error) {
	cookie, err := netnsCookie()
	if err != nil {
		return nil, fmt.Errorf("read network namespace cookie: %w", err)
	}
	return start(cgroup, cookie)
}

func netnsCookie() (uint64, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return 0, err
	}
	defer unix.Close(fd)
	return unix.GetsockoptUint64(fd, unix.SOL_SOCKET, unix.SO_NETNS_COOKIE)
}

func start(cgroup string, cookie uint64) (_ *Guard, err error) {
	dir, err := os.Open(cgroup)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	var stat unix.Statfs_t
	if err := unix.Fstatfs(int(dir.Fd()), &stat); err != nil {
		return nil, err
	}
	if stat.Type != unix.CGROUP2_SUPER_MAGIC {
		return nil, fmt.Errorf("%s is not on a cgroup v2 mount", cgroup)
	}
	guard := &Guard{}
	defer func() {
		if err != nil {
			guard.Close()
		}
	}()
	for _, attach := range []ebpf.AttachType{ebpf.AttachCGroupInet4Connect, ebpf.AttachCGroupInet6Connect} {
		prog, e := ebpf.NewProgram(programSpec(attach, cookie))
		if e != nil {
			return nil, fmt.Errorf("load bound-device connect guard (%s): %w", attach, e)
		}
		attached, e := link.AttachRawLink(link.RawLinkOptions{
			Target: int(dir.Fd()), Program: prog, Attach: attach,
		})
		prog.Close() // The link owns a kernel reference after a successful attach.
		if e != nil {
			return nil, fmt.Errorf("attach bound-device connect guard (%s): %w", attach, e)
		}
		guard.links = append(guard.links, attached)
	}
	return guard, nil
}

func (g *Guard) Close() error {
	if g == nil {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	var errs []error
	for i := len(g.links) - 1; i >= 0; i-- {
		errs = append(errs, g.links[i].Close())
	}
	g.links = nil
	return errors.Join(errs...)
}

func programSpec(attach ebpf.AttachType, cookie uint64) *ebpf.ProgramSpec {
	return &ebpf.ProgramSpec{
		Name: "v2raya_binddev", Type: ebpf.CGroupSockAddr,
		AttachType: attach, License: "GPL",
		Instructions: asm.Instructions{
			asm.Mov.Reg(asm.R6, asm.R1),
			asm.LoadMem(asm.R2, asm.R6, addrProtocolOffset, asm.Word),
			asm.JNE.Imm(asm.R2, unix.IPPROTO_TCP, "allow"),
			asm.FnGetNetnsCookie.Call(),
			asm.LoadImm(asm.R2, int64(cookie), asm.DWord),
			asm.JNE.Reg(asm.R0, asm.R2, "allow"),
			asm.LoadMem(asm.R2, asm.R6, addrSocketOffset, asm.DWord),
			asm.JEq.Imm(asm.R2, 0, "allow"),
			asm.LoadMem(asm.R3, asm.R2, sockBoundDevOffset, asm.Word),
			asm.JEq.Imm(asm.R3, 0, "allow"),
			asm.LoadMem(asm.R3, asm.R2, sockMarkOffset, asm.Word),
			asm.JSet.Imm(asm.R3, BypassMark, "allow"),
			asm.Or.Imm(asm.R3, BypassMark),
			asm.StoreMem(asm.RFP, -4, asm.R3, asm.Word),
			asm.Mov.Reg(asm.R1, asm.R6),
			asm.Mov.Imm(asm.R2, unix.SOL_SOCKET),
			asm.Mov.Imm(asm.R3, unix.SO_MARK),
			asm.Mov.Reg(asm.R4, asm.RFP),
			asm.Add.Imm(asm.R4, -4),
			asm.Mov.Imm(asm.R5, 4),
			asm.FnSetsockopt.Call(),
			asm.JEq.Imm(asm.R0, 0, "allow"),
			// If the requested exemption cannot be applied, reject connect
			// rather than silently send a bound socket through REDIRECT.
			asm.Mov.Imm(asm.R0, 0),
			asm.Return(),
			asm.Mov.Imm(asm.R0, 1).WithSymbol("allow"),
			asm.Return(),
		},
	}
}
