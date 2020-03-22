package dnsPoison

import (
	"log"
	"net"
	"syscall"
	"time"
)

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
