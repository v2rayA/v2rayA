package global

import (
	"os/exec"
	"strings"
)

type SystemServiceControlMode int

const (
	SystemctlMode = SystemServiceControlMode(iota)
	ServiceMode
	UniversalMode
)

func SetServiceControlMode(modeString string) (mode SystemServiceControlMode) {
	defer func() {
		ServiceControlMode = mode
	}()
	switch modeString {
	case "systemctl":
		return SystemctlMode
	case "service":
		return ServiceMode
	case "universal", "common":
		return UniversalMode
	default:
		//自动检测
		if out, err := exec.Command("sh", "-c", "which systemctl").Output(); err == nil && strings.Contains(string(out), "systemctl") {
			return SystemctlMode
		}
		if out, err := exec.Command("sh", "-c", "which service").Output(); err == nil && strings.Contains(string(out), "service") {
			return ServiceMode
		}
		return UniversalMode
	}
}
