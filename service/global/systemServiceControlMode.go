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
	switch modeString {
	case "systemctl":
		mode = SystemctlMode
	case "service":
		mode = ServiceMode
	case "universal", "common":
		mode = UniversalMode
	default:
		//自动检测
		if out, err := exec.Command("sh", "-c", "which systemctl").Output(); err == nil && strings.Contains(string(out), "systemctl") {
			mode = SystemctlMode
			return
		}
		if out, err := exec.Command("sh", "-c", "which service").Output(); err == nil && strings.Contains(string(out), "service") {
			mode = ServiceMode
			return
		}
		mode = UniversalMode
	}
	ServiceControlMode = mode
	return
}
