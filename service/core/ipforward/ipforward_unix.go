//go:build darwin || dragonfly || freebsd || (js && wasm) || netbsd || openbsd
// +build darwin dragonfly freebsd js,wasm netbsd openbsd

package ipforward

import (
	"fmt"
	"github.com/v2rayA/v2rayA/db/configure"
	"os/exec"
	"strings"
	"syscall"
)

func IsIpForwardOn() bool {
	if setting := configure.GetSettingNotNil(); setting.Transparent == configure.TransparentClose {
		return setting.IntranetSharing
	}
	out, err := syscall.Sysctl("net.inet.ip.forwarding")
	return err == nil && strings.TrimSpace(out) == "1"
}

func WriteIpForward(on bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("WriteIpForward: %w", err)
		}
	}()
	if setting := configure.GetSettingNotNil(); setting.Transparent == configure.TransparentClose {
		setting.IntranetSharing = on
	}
	val := "0"
	if on {
		val = "1"
	}
	_, err = exec.Command("sysctl -w net.inet.ip.forwarding=" + val).CombinedOutput()
	return err
}
