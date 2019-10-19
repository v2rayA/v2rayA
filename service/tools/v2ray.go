package tools

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func RestartV2rayService() (err error) {
	_, err = exec.Command("systemctl", "restart", "v2ray").Output()
	return
}

func WriteV2rayConfig(content string) (err error) {
	return ioutil.WriteFile("/etc/v2ray/config.json", []byte(content), os.ModeAppend)
}
