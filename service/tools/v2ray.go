package tools

import (
	"os/exec"
)

func RestartV2rayService() (err error) {
	_, err = exec.Command("systemctl", "restart", "v2ray").Output()
	return
}
