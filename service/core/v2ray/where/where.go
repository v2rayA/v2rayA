package where

import (
	"bytes"
	"fmt"
	"github.com/v2rayA/v2rayA/global"
	"os/exec"
	"strings"
)

var NotFoundErr = fmt.Errorf("not found")
var ServiceNameList = []string{"v2ray"}

func GetV2rayServiceFileContent() (content string, err error) {
	switch global.ServiceControlMode {
	case global.SystemctlMode:
		content, err = getV2raySystemdServiceFileContent()
		if err != nil {
			break
		}
	case global.ServiceMode:
		p, _ := getV2rayServiceFilePath()
		var out []byte
		out, err = exec.Command("sh", "-c", "cat "+p+"|grep Environment=V2RAY_LOCATION_ASSET").CombinedOutput()
		if err != nil {
			break
		}
		content = string(bytes.TrimSpace(out))
	}
	return
}
func getV2rayServiceFilePath() (path string, err error) {
	for _, target := range ServiceNameList {
		if path, err = _getV2rayServiceFilePath(target); err == nil {
			return
		}
	}
	return
}
func getV2raySystemdServiceFileContent() (content string, err error) {
	for _, target := range ServiceNameList {
		if out, err := exec.Command("sh", "-c", "systemctl cat "+target).CombinedOutput(); err == nil {
			return string(bytes.TrimSpace(out)), nil
		}
	}
	return
}
func _getV2rayServiceFilePath(target string) (path string, err error) {
	var out []byte
	if global.ServiceControlMode == global.SystemctlMode {
		out, err = exec.Command("sh", "-c", "systemctl status "+target+"|grep /"+target+".service").CombinedOutput()
		if err != nil {
			err = newError(strings.TrimSpace(string(out)))
			if !strings.Contains(string(out), "not be found") {
				path = `/usr/lib/systemd/system/` + target + `.service`
				return
			}
		}
	} else if global.ServiceControlMode == global.ServiceMode {
		out, err = exec.Command("sh", "-c", "service "+target+" status|grep /"+target+".service").CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) == "(Reason:" {
			if !strings.Contains(string(out), "not be found") {
				path = `/lib/systemd/system/` + target + `.service`
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
	r := strings.Index(sout, "/"+target+".service")
	if l < 0 || r < 0 {
		err = newError("failure: _getV2rayServiceFilePath")
		return
	}
	path = sout[l : r+len("/"+target+".service")]
	return
}

/* get the version of v2ray-core without 'v' like 4.23.1 */
func GetV2rayServiceVersion() (ver string, err error) {
	v2rayPath, err := GetV2rayBinPath()
	if err != nil || len(v2rayPath) <= 0 {
		return "", newError("cannot find v2ray executable binary")
	}
	out, err := exec.Command("sh", "-c", fmt.Sprintf("%v -version", v2rayPath)).Output()
	var fields []string
	if fields = strings.Fields(strings.TrimSpace(string(out))); len(fields) < 2 {
		return "", newError("cannot parse version of v2ray")
	}
	ver = fields[1]
	if strings.ToUpper(fields[0]) != "V2RAY" {
		ver = "UnknownClient"
	}
	return
}

func GetV2rayBinPath() (string, error) {
	v2rayBinPath := global.GetEnvironmentConfig().V2rayBin
	if v2rayBinPath == "" {
		return getV2rayBinPathAnyway()
	}
	return v2rayBinPath, nil
}

func getV2rayBinPathAnyway() (path string, err error) {
	for _, target := range ServiceNameList {
		if path, err = getV2rayBinPath(target); err == nil {
			return
		}
	}
	return
}

func getV2rayBinPath(target string) (string, error) {
	var pa string
	//从环境变量里找
	out, err := exec.Command("sh", "-c", "which "+target).CombinedOutput()
	if err != nil {
		return "", NotFoundErr
	}
	pa = strings.TrimSpace(string(out))
	if pa == "" {
		return "", NotFoundErr
	}
	return pa, nil
}
