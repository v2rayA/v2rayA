//go:build darwin || dragonfly || freebsd || (js && wasm) || netbsd || openbsd
// +build darwin dragonfly freebsd js,wasm netbsd openbsd

package ipforward

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func IsIpForwardOn() bool {
	out, err := syscall.Sysctl("net.inet.ip.forwarding")
	return err == nil && strings.TrimSpace(out) == "1"
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
	_, err = exec.Command("sysctl -w net.inet.ip.forwarding=" + val).CombinedOutput()
	return err
}
