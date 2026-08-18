//go:build !windows
// +build !windows

package privilege

import "os"

// IsRootOrAdmin reports whether the current process runs as root on Unix-like systems.
func IsRootOrAdmin() bool {
	return os.Geteuid() == 0
}
