// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"syscall"

	"golang.zx2c4.com/wireguard/tun"
)

// ErrUnsupportedPlatform is returned where no address configuration exists;
// the device can still be created, but nothing would route into it.
var ErrUnsupportedPlatform = errors.New("tunmips: TUN is not supported on this platform")

// openDevice creates the TUN device and puts the operating-system side of
// the link into service: interface addresses assigned, link up. On macOS
// name is a prefix ("utun") and the kernel picks the unit; Name() reports
// the real one.
func openDevice(name string, mtu uint32, addr4, addr6 netip.Prefix) (tun.Device, error) {
	if !addr4.IsValid() || !addr4.Addr().Is4() {
		return nil, fmt.Errorf("tunmips: address4 %q is not an IPv4 prefix", addr4)
	}
	if addr6.IsValid() && !addr6.Addr().Is6() {
		return nil, fmt.Errorf("tunmips: address6 %q is not an IPv6 prefix", addr6)
	}
	dev, err := tun.CreateTUN(name, int(mtu))
	if err != nil {
		if errors.Is(err, os.ErrExist) || errors.Is(err, syscall.EBUSY) || errors.Is(err, syscall.EEXIST) {
			// A previous core that was killed rather than stopped leaves its
			// device behind; nothing here deletes it, because the routes
			// that came with it are not ours to guess at.
			return nil, fmt.Errorf("tunmips: device %q already exists, left by an earlier run: delete it (Linux: ip link del %s) and remove its routes, or reboot: %w", name, name, err)
		}
		return nil, fmt.Errorf("tunmips: create device %q: %w", name, err)
	}
	if err := configureAddresses(dev, mtu, addr4, addr6); err != nil {
		dev.Close()
		return nil, err
	}
	return dev, nil
}

// peerOf returns the other usable address of a point-to-point prefix: the
// gateway the operating system will send to. For 10.0.85.2/30 it is
// 10.0.85.1.
func peerOf(p netip.Prefix) netip.Addr {
	first := p.Masked().Addr().Next()
	if first == p.Addr() {
		return first.Next()
	}
	return first
}
