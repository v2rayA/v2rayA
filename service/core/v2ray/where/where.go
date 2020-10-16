package where

import (
	"fmt"
	"github.com/v2rayA/v2rayA/global"
	"os/exec"
	"path"
	"strings"
)

func GetV2rayServiceFilePath() (path string, err error) {
	var out []byte

	if global.ServiceControlMode == global.SystemctlMode {
		out, err = exec.Command("sh", "-c", "systemctl status v2ray|grep /v2ray.service").CombinedOutput()
		if err != nil {
			err = newError(strings.TrimSpace(string(out)))
			if !strings.Contains(string(out), "not be found") {
				path = `/usr/lib/systemd/system/v2ray.service`
				return
			}
		}
	} else if global.ServiceControlMode == global.ServiceMode {
		out, err = exec.Command("sh", "-c", "service v2ray status|grep /v2ray.service").CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) == "(Reason:" {
			if !strings.Contains(string(out), "not be found") {
				path = `/lib/systemd/system/v2ray.service`
				return
			}
			if err != nil {
				err = newError(strings.TrimSpace(string(out)))
			}
		}
	} else {
		err = newError("commands systemctl and service not found")
		return
	}
	if err != nil {
		return
	}
	sout := string(out)
	l := strings.Index(sout, "/")
	r := strings.Index(sout, "/v2ray.service")
	if l < 0 || r < 0 {
		err = newError("fail: getV2rayServiceFilePath")
		return
	}
	path = sout[l : r+len("/v2ray.service")]
	return
}
/* get the version of v2ray-core without 'v' like 4.23.1 */
func GetV2rayServiceVersion() (ver string, err error) {
	dir, err := GetV2rayWorkingDir()
	if err != nil || len(dir) <= 0 {
		return "", newError("cannot find v2ray executable binary")
	}
	out, err := exec.Command("sh", "-c", fmt.Sprintf("%v/v2ray -version|awk '{print $2}'|awk 'NR==1'", dir)).Output()
	return strings.TrimSpace(string(out)), err
}

func GetV2rayWorkingDir() (string, error) {
	switch global.ServiceControlMode {
	case global.SystemctlMode, global.ServiceMode:
		//从systemd的启动参数里找
		p, _ := GetV2rayServiceFilePath()
		out, err := exec.Command("sh", "-c", "cat "+p+"|grep ExecStart=").CombinedOutput()
		if err != nil {
			return "", newError(string(out)).Base(err)
		}
		arr := strings.SplitN(strings.TrimSpace(string(out)), " ", 2)
		return path.Dir(strings.TrimPrefix(arr[0], "ExecStart=")), nil
	case global.UniversalMode:
		//从环境变量里找
		out, err := exec.Command("sh", "-c", "which v2ray").CombinedOutput()
		if err == nil {
			return path.Dir(strings.TrimSpace(string(out))), nil
		}
	}
	return "", newError("not found")
}