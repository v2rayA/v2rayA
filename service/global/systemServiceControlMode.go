package global

import (
	"os/exec"
	"strings"
)

type SystemServiceControlMode int

const (
	SystemctlMode SystemServiceControlMode = iota
	ServiceMode
	UniversalMode
)

func SetServiceControlMode() (mode SystemServiceControlMode) {
	// The behaviour of this function has changed after v1.1.4
	if out, err := exec.Command("sh", "-c", "which systemctl").Output(); err == nil && strings.Contains(string(out), "systemctl") {
		return SystemctlMode
	}
	if out, err := exec.Command("sh", "-c", "which service").Output(); err == nil && strings.Contains(string(out), "service") {
		return ServiceMode
	}
	return UniversalMode
}
