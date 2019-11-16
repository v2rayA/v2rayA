package global

import (
	"os/exec"
	"strings"
)

type SystemServiceControlMode int

const (
	SystemctlMode = SystemServiceControlMode(iota)
	ServiceMode
	CommonMode
	DockerMode
)

func GetServiceControlMode() (mode SystemServiceControlMode) {
	if _, err := exec.Command("ls", "/.dockerenv").Output(); err == nil {
		mode = DockerMode
		return
	}
	if out, err := exec.Command("sh", "-c", "type -p systemctl").Output(); strings.Contains(string(out), "systemctl") && err == nil {
		mode = SystemctlMode
		return
	}
	if out, err := exec.Command("sh", "-c", "type -p service").Output(); strings.Contains(string(out), "service") && err == nil {
		mode = ServiceMode
		return
	}
	mode = CommonMode
	return
}
