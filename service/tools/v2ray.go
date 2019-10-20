package tools

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func RestartV2rayService() (err error) {
	_, err = exec.Command("service", "v2ray", "restart").Output()
	if err != nil && strings.Index(err.Error(), "command not found") > -1 {
		_, err = exec.Command("systemctl", "restart", "v2ray").Output()
	}
	return
}

func WriteV2rayConfig(content string) (err error) {
	return ioutil.WriteFile("/etc/v2ray/config.json", []byte(content), os.ModeAppend)
}

func IsV2RayRunning() bool {
	out, err := exec.Command("sh", "-c", "service v2ray status|head -n 5|grep running").Output()
	if err != nil && strings.Index(err.Error(), "command not found") > -1 {
		out, err = exec.Command("sh", "-c", "systemctl status v2ray|head -n 5|grep running").Output()
	}
	return err == nil && len(out) > 0
}
