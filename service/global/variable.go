package global

import (
	"os/exec"
	"strings"
)

var ServiceControlMode SystemServiceControlMode
var V2RAY_LOCATION_ASSET string

func getV2rayLocationAsset() (s string) {
	switch ServiceControlMode {
	case DockerMode:
		return "/etc/v2ray"
	case SystemctlMode, ServiceMode:
		var (
			p   string
			out []byte
			err error
		)
		if ServiceControlMode == SystemctlMode {
			out, err = exec.Command("sh", "-c", "systemctl status v2ray|grep Loaded|awk '{print $3}'").Output()
			if err != nil {
				p = `/usr/lib/systemd/system/v2ray.service`
			}
		} else {
			out, err = exec.Command("sh", "-c", "systemctl v2ray status|grep Loaded|awk '{print $3}'").Output()
			if err != nil {
				p = `/lib/systemd/system/v2ray.service`
			}
		}
		sout := strings.TrimSpace(string(out))
		p = sout[1 : len(sout)-1]
		out, err = exec.Command("sh", "-c", "cat "+p+"|grep Environment=V2RAY_LOCATION_ASSET").Output()
		if err != nil {
			return
		}
		s = strings.TrimSpace(string(out))
		s = s[len("Environment=V2RAY_LOCATION_ASSET="):]
	}
	var err error
	if s == "" {
		//默认为v2ray运行目录
		s, err = getV2rayWorkingDir()
	}
	if err != nil {
		//再不行盲猜一个
		s = "/etc/v2ray"
	}
	return
}

func getV2rayWorkingDir() (string, error) {
	out, err := exec.Command("sh", "-c", "type -p v2ray").Output()
	if err != nil {
		out, err = exec.Command("sh", "-c", "which v2ray").Output()
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func init() {
	ServiceControlMode = GetServiceControlMode()
	V2RAY_LOCATION_ASSET = getV2rayLocationAsset()
}
