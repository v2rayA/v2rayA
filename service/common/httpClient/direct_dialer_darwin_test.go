package httpClient

import (
	"strings"

	"golang.org/x/sys/unix"
)

func boundProbeInterface(fd uintptr, network string) (int, error) {
	if strings.HasSuffix(network, "6") {
		return unix.GetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF)
	}
	return unix.GetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF)
}
