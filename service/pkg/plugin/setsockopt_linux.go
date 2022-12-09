//go:build linux
// +build linux

package plugin

import (
	"fmt"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"golang.org/x/sys/unix"
	"net"
	"runtime"
	"syscall"
)

var fwmarkIoctl int

func init() {
	switch runtime.GOOS {
	case "linux", "android":
		fwmarkIoctl = 36 /* unix.SO_MARK */
	case "freebsd":
		fwmarkIoctl = 0x1015 /* unix.SO_USER_COOKIE */
	case "openbsd":
		fwmarkIoctl = 0x1021 /* unix.SO_RTABLE */
	}
}

func SoMarkControl(c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		//TODO: force to set 0xff. any chances to customize this value?
		err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, fwmarkIoctl, 0x80)
		if err != nil {
			return
		}
	})
}
func BindControl(c syscall.RawConn, laddr string, lport uint32) error {
	return c.Control(func(fd uintptr) {
		err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, fwmarkIoctl, 0x80)
		if err != nil {
			log.Warn("control: %s", err)
			return
		}
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
			log.Warn("control: failed to set IP_TRANSPARENT")
			return
		}
		ip := net.ParseIP(laddr).To4()
		if err := bindAddr(fd, ip, lport); err != nil {
			log.Warn("control: %s", err)
			return
		}
	})
}

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return fmt.Errorf("failed to set resuse_addr")
	}

	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return fmt.Errorf("failed to set resuse_port")
	}

	var sockaddr syscall.Sockaddr

	switch len(ip) {
	case net.IPv4len:
		a4 := &syscall.SockaddrInet4{
			Port: int(port),
		}
		copy(a4.Addr[:], ip)
		sockaddr = a4
	case net.IPv6len:
		a6 := &syscall.SockaddrInet6{
			Port: int(port),
		}
		copy(a6.Addr[:], ip)
		sockaddr = a6
	default:
		return fmt.Errorf("unexpected length of ip")
	}

	return syscall.Bind(int(fd), sockaddr)
}
