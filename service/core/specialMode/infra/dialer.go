package infra

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"syscall"
	"time"
)

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return newError("failed to set resuse_addr")
	}

	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return newError("failed to set resuse_port")
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
		return newError("unexpected length of ip")
	}

	return syscall.Bind(int(fd), sockaddr)
}

func newDialer(laddr string, lport uint32, timeout time.Duration) (dialer *net.Dialer) {
	return &net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, 0xff)
				if err != nil {
					log.Printf("control: %s", err)
					return
				}
				if err := syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
					log.Println("control: failed to set IP_TRANSPARENT")
					return
				}
				ip := net.ParseIP(laddr).To4()
				if err := bindAddr(fd, ip, lport); err != nil {
					log.Printf("control: %s", err)
					return
				}
			})
		},
	}
}
