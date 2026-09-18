// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/tun"
)

// From netinet6/in6_var.h and netinet6/nd6.h.
const (
	siocAIFADDR6        = 2155899162
	in6IFFNoDAD         = 0x0020
	nd6InfiniteLifetime = 0xFFFFFFFF
)

type ifAliasReq4 struct {
	Name    [unix.IFNAMSIZ]byte
	Addr    unix.RawSockaddrInet4
	Dstaddr unix.RawSockaddrInet4
	Mask    unix.RawSockaddrInet4
}

type addrLifetime6 struct {
	Expire    float64
	Preferred float64
	Vltime    uint32
	Pltime    uint32
}

type ifAliasReq6 struct {
	Name     [unix.IFNAMSIZ]byte
	Addr     unix.RawSockaddrInet6
	Dstaddr  unix.RawSockaddrInet6
	Mask     unix.RawSockaddrInet6
	Flags    uint32
	Lifetime addrLifetime6
}

// configureAddresses assigns the point-to-point IPv4 pair and, when given,
// the IPv6 address through the same ioctls ifconfig uses. utun is a
// point-to-point interface, so IPv4 needs the peer (gateway) address too.
func configureAddresses(dev tun.Device, _ uint32, addr4, addr6 netip.Prefix) error {
	name, err := dev.Name()
	if err != nil {
		return err
	}
	if err := setAddr4(name, addr4); err != nil {
		return err
	}
	if addr6.IsValid() {
		if err := setAddr6(name, addr6); err != nil {
			return err
		}
	}
	return nil
}

func setAddr4(name string, p netip.Prefix) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	req := ifAliasReq4{
		Addr:    unix.RawSockaddrInet4{Len: unix.SizeofSockaddrInet4, Family: unix.AF_INET, Addr: p.Addr().As4()},
		Dstaddr: unix.RawSockaddrInet4{Len: unix.SizeofSockaddrInet4, Family: unix.AF_INET, Addr: peerOf(p).As4()},
		Mask:    unix.RawSockaddrInet4{Len: unix.SizeofSockaddrInet4, Family: unix.AF_INET, Addr: [4]byte(net.CIDRMask(p.Bits(), 32))},
	}
	copy(req.Name[:], name)
	if err := ioctlPtr(fd, unix.SIOCAIFADDR, unsafe.Pointer(&req)); err != nil {
		return os.NewSyscallError("SIOCAIFADDR", err)
	}
	return nil
}

func setAddr6(name string, p netip.Prefix) error {
	fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	req := ifAliasReq6{
		Addr:     unix.RawSockaddrInet6{Len: unix.SizeofSockaddrInet6, Family: unix.AF_INET6, Addr: p.Addr().As16()},
		Mask:     unix.RawSockaddrInet6{Len: unix.SizeofSockaddrInet6, Family: unix.AF_INET6, Addr: [16]byte(net.CIDRMask(p.Bits(), 128))},
		Flags:    in6IFFNoDAD,
		Lifetime: addrLifetime6{Vltime: nd6InfiniteLifetime, Pltime: nd6InfiniteLifetime},
	}
	copy(req.Name[:], name)
	if err := ioctlPtr(fd, siocAIFADDR6, unsafe.Pointer(&req)); err != nil {
		return os.NewSyscallError("SIOCAIFADDR6", err)
	}
	return nil
}

func ioctlPtr(fd int, req uint, arg unsafe.Pointer) error {
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg)); errno != 0 {
		return fmt.Errorf("ioctl %#x: %w", req, errno)
	}
	return nil
}
