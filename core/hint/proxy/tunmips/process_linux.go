// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// platformOwnerOf asks the kernel for the socket via sock_diag, then finds
// the processes whose fd tables reference its inode.
func platformOwnerOf(proto string, src netip.AddrPort) ([]owner, error) {
	inode, err := socketInode(proto, src)
	if err != nil {
		return nil, err
	}
	if inode == 0 {
		return nil, nil
	}
	return ownersOfInode(inode)
}

func socketInode(proto string, src netip.AddrPort) (uint32, error) {
	family := uint8(unix.AF_INET)
	if src.Addr().Unmap().Is6() {
		family = unix.AF_INET6
	}
	var (
		sockets []*netlink.Socket
		err     error
	)
	switch proto {
	case "tcp":
		sockets, err = netlink.SocketDiagTCP(family)
	case "udp":
		sockets, err = netlink.SocketDiagUDP(family)
	default:
		return 0, fmt.Errorf("unknown protocol %q", proto)
	}
	if err != nil {
		return 0, fmt.Errorf("sock_diag: %w", err)
	}
	want := src.Addr().Unmap()
	var wildcard uint32
	for _, s := range sockets {
		if s.ID.SourcePort != src.Port() {
			continue
		}
		local, ok := netip.AddrFromSlice(s.ID.Source)
		if !ok {
			continue
		}
		local = local.Unmap()
		if local == want {
			return s.INode, nil
		}
		if local.IsUnspecified() && wildcard == 0 {
			wildcard = s.INode
		}
	}
	return wildcard, nil
}

// ownersOfInode scans /proc/<pid>/fd for "socket:[inode]". The scan is the
// expensive part; it runs once per TCP connection and once per UDP session.
func ownersOfInode(inode uint32) ([]owner, error) {
	target := "socket:[" + strconv.FormatUint(uint64(inode), 10) + "]"
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var owners []owner
	var lastErr error
	for _, p := range procs {
		pid, err := strconv.ParseUint(p.Name(), 10, 32)
		if err != nil {
			continue
		}
		fdDir := filepath.Join("/proc", p.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			if os.IsPermission(err) {
				lastErr = err
			}
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil || link != target {
				continue
			}
			owners = append(owners, owner{pid: uint32(pid), name: processName(uint32(pid))})
			break
		}
	}
	if len(owners) == 0 && lastErr != nil {
		return nil, fmt.Errorf("scanning /proc: %w", lastErr)
	}
	return owners, nil
}

// processName prefers the executable's basename; comm is only 15 bytes.
func processName(pid uint32) string {
	if exe, err := os.Readlink(filepath.Join("/proc", strconv.FormatUint(uint64(pid), 10), "exe")); err == nil {
		return filepath.Base(strings.TrimSuffix(exe, " (deleted)"))
	}
	if comm, err := os.ReadFile(filepath.Join("/proc", strconv.FormatUint(uint64(pid), 10), "comm")); err == nil {
		return strings.TrimSpace(string(comm))
	}
	return ""
}
