// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/tun"
)

// configureAddresses assigns the addresses and brings the link up. The
// addresses vanish with the device, so there is no counterpart.
func configureAddresses(dev tun.Device, _ uint32, addr4, addr6 netip.Prefix) error {
	name, err := dev.Name()
	if err != nil {
		return err
	}
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("tunmips: %s: %w", name, err)
	}
	for _, p := range []netip.Prefix{addr4, addr6} {
		if !p.IsValid() {
			continue
		}
		addr := &netlink.Addr{IPNet: &net.IPNet{IP: p.Addr().AsSlice(), Mask: net.CIDRMask(p.Bits(), p.Addr().BitLen())}}
		if err := netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("tunmips: add %s to %s: %w", p, name, err)
		}
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("tunmips: %s up: %w", name, err)
	}
	return nil
}
