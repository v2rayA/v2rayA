package global

import (
	"os/exec"
)

type SystemServiceControlMode int

const (
	SystemctlMode SystemServiceControlMode = iota
	ServiceMode
	UniversalMode
)

func SetServiceControlMode() (mode SystemServiceControlMode) {
	// The behaviour of this function has changed after v1.1.4

	if _, err := exec.LookPath("systemctl"); err == nil {
		return SystemctlMode
	}
	if _, err := exec.LookPath("service"); err == nil {
		return ServiceMode
	}
	return UniversalMode
}
