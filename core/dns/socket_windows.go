//go:build windows

package dns

// setSocketMark is a no-op on Windows since SO_MARK is a Linux-specific
// socket option used for iptables/nftables mark-based filtering.
func setSocketMark(fd uintptr) error {
	return nil
}
