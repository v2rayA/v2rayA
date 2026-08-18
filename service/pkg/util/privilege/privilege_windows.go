//go:build windows
// +build windows

package privilege

import (
	"os/user"
	"strings"

	"golang.org/x/sys/windows"
)

// IsRootOrAdmin reports whether the current process runs with SYSTEM or
// Administrators privileges. It prefers token elevation and falls back to
// group membership checks for robustness.
func IsRootOrAdmin() bool {
	token := windows.Token(0)

	if elevated := token.IsElevated(); elevated {
		return true
	}

	if adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid); err == nil {
		if member, err := token.IsMember(adminSID); err == nil && member {
			return true
		}
	}

	if systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid); err == nil {
		if member, err := token.IsMember(systemSID); err == nil && member {
			return true
		}
	}

	if u, err := user.Current(); err == nil {
		name := strings.ToLower(u.Username)
		if strings.Contains(name, "system") || strings.Contains(name, "administrator") {
			return true
		}
	}

	return false
}
