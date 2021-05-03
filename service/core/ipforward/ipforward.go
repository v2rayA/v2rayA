package ipforward

import (
	"os"
	"strings"
)

func IsIpForwardOn() bool {
	out, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	return err == nil && strings.TrimSpace(string(out)) == "1"
}

//返回ipv4.ip_forward的开启状态。该命令写的ip_forward状态重启将失效。
func WriteIpForward(on bool) (err error) {
	val := "0"
	if on {
		val = "1"
	}
	err = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte(val), 0644)
	// ipv6
	_ = os.WriteFile("/proc/sys/net/ipv6/conf/all/forwarding", []byte(val), 0644)
	return
}
