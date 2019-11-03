package global

import (
	"V2RayA/models/v2ray"
	"os/exec"
)

var ServiceControlMode v2ray.ServiceControlMode
var V2RAY_LOCATION_ASSET string

func getV2rayLocationAsset() (s string) {
	switch ServiceControlMode {
	case v2ray.Docker:
		return "/etc/v2ray"
	case v2ray.Systemctl, v2ray.Service:
		var (
			p   string
			out []byte
			err error
		)
		if ServiceControlMode == v2ray.Systemctl {
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
		sout := string(out)
		p = sout[1 : len(sout)-1]
		out, err = exec.Command("sh", "-c", "cat "+p+"|grep Environment=V2RAY_LOCATION_ASSET").Output()
		if err != nil {
			return
		}
		s = string(out)
		s = s[len("Environment=V2RAY_LOCATION_ASSET="):]
	}
	if s == "" {
		s = "/etc/v2ray"
	}
	return
}

func init() {
	ServiceControlMode = v2ray.NewServiceControlMode()
	V2RAY_LOCATION_ASSET = getV2rayLocationAsset()
}
