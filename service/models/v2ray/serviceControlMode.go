package v2ray

import (
	"os/exec"
	"strings"
)

type ServiceControlMode int

const (
	Systemctl = ServiceControlMode(iota)
	Service
	Common
	Docker
)

func NewServiceControlMode() (mode ServiceControlMode) {
	if _, err := exec.Command("ls", "/.dockerenv").Output(); err == nil {
		mode = Docker
		return
	}
	if out, err := exec.Command("sh", "-c", "type -p systemctl").Output(); strings.Contains(string(out), "systemctl") && err == nil {
		mode = Systemctl
		return
	}
	if out, err := exec.Command("sh", "-c", "type -p service").Output(); strings.Contains(string(out), "service") && err == nil {
		mode = Service
		return
	}
	mode = Common
	return
}
