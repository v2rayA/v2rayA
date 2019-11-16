package v2ray

import (
	"V2RayA/global"
	"os/exec"
)

func EnableV2rayService() (err error) {
	switch global.ServiceControlMode {
	case global.DockerMode, global.CommonMode: //docker, common中无需enable service
	case global.ServiceMode:
		_, err = exec.Command("sh", "-c", "update-rc.d v2ray enable").CombinedOutput()
	case global.SystemctlMode:
		_, err = exec.Command("sh", "-c", "systemctl enable v2ray").Output()
	}
	return
}

func DisableV2rayService() (err error) {
	switch global.ServiceControlMode {
	case global.DockerMode, global.CommonMode: //docker, common中无需disable service
	case global.ServiceMode:
		_, err = exec.Command("sh", "-c", "update-rc.d v2ray disable").CombinedOutput()
	case global.SystemctlMode:
		_, err = exec.Command("sh", "-c", "systemctl disable v2ray").Output()
	}
	return
}
