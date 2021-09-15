//go:build linux
// +build linux

package ipforward

import (
	"fmt"
	"os"
	"strings"
)

func IsIpForwardOn() bool {
	out, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	return err == nil && strings.TrimSpace(string(out)) == "1"
}

func WriteIpForward(on bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("WriteIpForward: %w", err)
		}
	}()
	val := "0"
	if on {
		val = "1"
	}
	err = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte(val), 0644)
	// ipv6
	_ = os.WriteFile("/proc/sys/net/ipv6/conf/all/forwarding", []byte(val), 0644)
	return
}
