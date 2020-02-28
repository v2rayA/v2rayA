package v2ray

import (
	"V2RayA/global"
	"V2RayA/model/v2ray/asset"
	"V2RayA/tools"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func EnableV2rayService() (err error) {
	var out []byte
	switch global.ServiceControlMode {
	case global.DockerMode, global.UniversalMode: //docker, universal中无需enable service
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "update-rc.d v2ray enable").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl enable v2ray").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	}
	return
}

func DisableV2rayService() (err error) {
	var out []byte
	switch global.ServiceControlMode {
	case global.DockerMode, global.UniversalMode: //docker, universal中无需disable service
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "update-rc.d v2ray disable").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl disable v2ray").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	}
	return
}

func LiberalizeProcFile() (err error) {
	if global.ServiceControlMode != global.SystemctlMode && global.ServiceControlMode != global.ServiceMode {
		return
	}
	p, err := asset.GetV2rayServiceFilePath()
	if err != nil {
		return
	}
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return
	}
	s := string(b)
	if strings.Contains(s, "LimitNPROC=500") && strings.Contains(s, "LimitNOFILE=1000000") {
		return
	}
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "LimitNPROC=") || strings.HasPrefix(lines[i], "LimitNOFILE=") {
			lines = append(lines[:i], lines[i+1:]...)
		}
	}
	for i, line := range lines {
		if strings.ToLower(line) == "[service]" {
			s = strings.Join(lines[:i+1], "\n")
			s += "\nLimitNPROC=500\nLimitNOFILE=1000000\n"
			s += strings.Join(lines[i+1:], "\n")
			break
		}
	}
	err = ioutil.WriteFile(p, []byte(s), os.ModeAppend)
	if err != nil {
		return
	}
	if IsV2RayRunning() {
		err = RestartV2rayService()
	}
	return
}

func IsV2rayServiceValid() bool {
	switch global.ServiceControlMode {
	case global.SystemctlMode:
		out, err := exec.Command("sh", "-c", "systemctl list-unit-files|grep v2ray.service").Output()
		return err == nil && len(bytes.TrimSpace(out)) > 0
	case global.ServiceMode:
		out, err := exec.Command("sh", "-c", "service v2ray status|grep not-found").Output()
		return err == nil && len(bytes.TrimSpace(out)) == 0
	case global.DockerMode:
		return asset.IsGeoipExists() && asset.IsGeositeExists()
	case global.UniversalMode:
		if !asset.IsGeoipExists() || !asset.IsGeositeExists() {
			return false
		}
		out, err := exec.Command("sh", "-c", "which v2ray").Output()
		return err == nil && len(bytes.TrimSpace(out)) > 0
	}
	return false
}

func GetV2rayServiceVersion() (ver string, err error) {
	dir, err := asset.GetV2rayWorkingDir()
	if err != nil || len(dir) <= 0 {
		return "", errors.New("cannot find v2ray executable binary")
	}
	out, err := exec.Command("sh", "-c", fmt.Sprintf("%v/v2ray -version|awk '{print $2}'|awk 'NR==1'", dir)).Output()
	return strings.TrimSpace(string(out)), err
}

func IfTProxyModLoaded() bool {
	out, err := exec.Command("sh", "-c", "lsmod|grep xt_TPROXY").Output()
	return err == nil && len(bytes.TrimSpace(out)) > 0
}

func CheckAndProbeTProxy() (err error) {
	ver, err := GetV2rayServiceVersion()
	if err != nil {
		return errors.New("fail in getting the version of v2ray-core: " + err.Error())
	}
	if greaterEqual, err := tools.VersionGreaterEqual(ver, "4.19.1"); err != nil || !greaterEqual {
		return errors.New("the version of v2ray-core is lower than 4.19.1")
	}
	if !IfTProxyModLoaded() && global.ServiceControlMode != global.DockerMode { //docker下无法判断
		var out []byte
		out, err = exec.Command("sh", "-c", "modprobe xt_TPROXY").CombinedOutput()
		if err != nil {
			if !strings.Contains(string(out), "not found") {
				return errors.New("fail in modprobing xt_TPROXY: " + string(out))
			}
			// modprobe失败，不支持xt_TPROXY方案
			return errors.New("not support xt_TPROXY: " + string(out))
		}
	}
	return
}

func CheckDohSupported() (err error) {
	ver, err := GetV2rayServiceVersion()
	if err != nil {
		return errors.New("fail in getting the version of v2ray-core")
	}
	if greaterEqual, err := tools.VersionGreaterEqual(ver, "4.22.0"); err != nil || !greaterEqual {
		return errors.New("the version of v2ray-core is lower than 4.22.0")
	}
	return
}
