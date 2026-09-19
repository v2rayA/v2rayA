//go:build !linux && !windows && !darwin

// SPDX-License-Identifier: MPL-2.0
package tunmips

import "net/netip"

func platformOwnerOf(string, netip.AddrPort) ([]owner, error) {
	return nil, ErrUnsupportedPlatform
}
