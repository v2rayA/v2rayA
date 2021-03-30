package ipforward

import (
	"os"
	"os/exec"
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
	out, err := exec.Command("sh", "-c", "echo "+val+" > /proc/sys/net/ipv4/ip_forward").CombinedOutput()
	if err != nil {
		err = newError(string(out)).Base(err)
	}
	// ipv6
	_, _ = exec.Command("sh", "-c", "echo "+val+" > /proc/sys/net/ipv6/conf/all/forwarding").CombinedOutput()
	return
}
