// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"fmt"
	"net/netip"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

// configureAddresses assigns the addresses and sets the interface MTU:
// wintun's CreateTUN only records the MTU for its own ring buffers, the IP
// interface keeps reporting 65535 until winipcfg sets it.
func configureAddresses(dev tun.Device, mtu uint32, addr4, addr6 netip.Prefix) error {
	native, ok := dev.(*tun.NativeTun)
	if !ok {
		return fmt.Errorf("tunmips: unexpected device type %T", dev)
	}
	luid := winipcfg.LUID(native.LUID())
	prefixes := []netip.Prefix{addr4}
	if addr6.IsValid() {
		prefixes = append(prefixes, addr6)
	}
	if err := luid.SetIPAddresses(prefixes); err != nil {
		return fmt.Errorf("tunmips: set addresses: %w", err)
	}
	for _, family := range []winipcfg.AddressFamily{windows.AF_INET, windows.AF_INET6} {
		ipif, err := luid.IPInterface(family)
		if err != nil {
			continue
		}
		ipif.NLMTU = mtu
		if err := ipif.Set(); err != nil {
			return fmt.Errorf("tunmips: set MTU: %w", err)
		}
	}
	return nil
}
