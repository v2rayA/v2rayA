package httpClient

import (
	"encoding/binary"
	"strings"
	"syscall"
)

func boundProbeInterface(fd uintptr, network string) (int, error) {
	if strings.HasSuffix(network, "6") {
		return syscall.GetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IPV6, 31)
	}
	index, err := syscall.GetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, 31)
	var encoded [4]byte
	binary.NativeEndian.PutUint32(encoded[:], uint32(index))
	return int(binary.BigEndian.Uint32(encoded[:])), err
}
