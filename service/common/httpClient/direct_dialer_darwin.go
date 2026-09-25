package httpClient

import (
	"strings"

	"golang.org/x/sys/unix"
)

func bindProbeSocket(fd uintptr, network string, index int) error {
	if strings.HasSuffix(network, "6") {
		return unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF, index)
	}
	return unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, index)
}
