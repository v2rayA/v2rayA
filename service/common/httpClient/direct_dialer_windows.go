package httpClient

import (
	"encoding/binary"
	"strings"
	"syscall"
)

func bindProbeSocket(fd uintptr, network string, index int) error {
	// IP_UNICAST_IF and IPV6_UNICAST_IF share the option number, but
	// IPv4 takes the interface index in network order, IPv6 in host order.
	const unicastIf = 31
	if strings.HasSuffix(network, "6") {
		return syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IPV6, unicastIf, index)
	}
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], uint32(index))
	return syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, unicastIf, int(binary.NativeEndian.Uint32(encoded[:])))
}
