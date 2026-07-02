//go:build !windows

package dns

import (
	"syscall"
)

// setSocketMark sets SO_MARK on a socket identified by its file descriptor.
// SO_MARK (option 36) is used for iptables/nftables mark-based filtering
// to prevent DNS query loops (mark 0x80 → iptables RETURN).
// This is Linux-specific; on other Unix platforms it compiles but is a no-op
// at the syscall level (option 36 has a different meaning).
func setSocketMark(fd uintptr) error {
	return syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, 36, 0x80)
}
