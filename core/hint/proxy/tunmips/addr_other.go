//go:build !linux && !windows && !darwin

// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"net/netip"

	"golang.zx2c4.com/wireguard/tun"
)

func configureAddresses(tun.Device, uint32, netip.Prefix, netip.Prefix) error {
	return ErrUnsupportedPlatform
}
