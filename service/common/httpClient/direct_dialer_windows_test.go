package httpClient

import (
	"strings"
	"syscall"
)

// boundProbeInterface reports the interface the socket is bound to. Windows
// wants the index in network byte order on the way in, but reports it back in
// host order, so the value read here needs no swap.
func boundProbeInterface(fd uintptr, network string) (int, error) {
	if strings.HasSuffix(network, "6") {
		return syscall.GetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IPV6, 31)
	}
	return syscall.GetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, 31)
}
