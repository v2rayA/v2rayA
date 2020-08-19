package v2ray

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/core/v2ray/asset"
	"github.com/mzz2017/v2rayA/global"
)

func EnableV2rayService() (err error) {
	var out []byte
	switch global.ServiceControlMode {
	case global.UniversalMode: //docker, universal中无需enable service
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "update-rc.d v2ray enable").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl enable v2ray").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	}
	return
}

func DisableV2rayService() (err error) {
	var out []byte
	switch global.ServiceControlMode {
	case global.UniversalMode: //docker, universal中无需disable service
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "update-rc.d v2ray disable").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl disable v2ray").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	}
	return
}

func OptimizeServiceFile() (err error) {
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
	if strings.Contains(s, "LimitNPROC=500") &&
		strings.Contains(s, "LimitNOFILE=1000000") &&
		strings.Contains(s, "CapabilityBoundingSet=CAP_NET_BIND_SERVICE CAP_NET_RAW CAP_NET_ADMIN") {
		return
	}
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "LimitNPROC=") ||
			strings.HasPrefix(lines[i], "LimitNOFILE=") ||
			strings.HasPrefix(lines[i], "CapabilityBoundingSet=") {
			lines = append(lines[:i], lines[i+1:]...)
		}
	}
	for i, line := range lines {
		if strings.ToLower(line) == "[service]" {
			s = strings.Join(lines[:i+1], "\n")
			s += "\n"
			s += "LimitNPROC=500\n"
			s += "LimitNOFILE=1000000\n"
			s += "CapabilityBoundingSet=CAP_NET_BIND_SERVICE CAP_NET_RAW CAP_NET_ADMIN\n"
			s += strings.Join(lines[i+1:], "\n")
			break
		}
	}
	err = ioutil.WriteFile(p, []byte(s), os.ModeAppend)
	if err != nil {
		return
	}
	if global.ServiceControlMode == global.SystemctlMode {
		_, _ = exec.Command("sh", "-c", "systemctl daemon-reload").Output()
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
	case global.UniversalMode:
		if !asset.IsGeoipExists() || !asset.IsGeositeExists() {
			return false
		}
		out, err := exec.Command("sh", "-c", "which v2ray").Output()
		return err == nil && len(bytes.TrimSpace(out)) > 0
	}
	return false
}

/* get the version of v2ray-core without 'v' like 4.23.1 */
func GetV2rayServiceVersion() (ver string, err error) {
	dir, err := asset.GetV2rayWorkingDir()
	if err != nil || len(dir) <= 0 {
		return "", newError("cannot find v2ray executable binary")
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
		return newError("failed to get the version of v2ray-core").Base(err)
	}
	if greaterEqual, err := common.VersionGreaterEqual(ver, "4.19.1"); err != nil || !greaterEqual {
		return newError("the version of v2ray-core (" + ver + ") is lower than 4.19.1")
	}
	if !IfTProxyModLoaded() && !common.IsInDocker() { //docker下无法判断
		var out []byte
		out, err = exec.Command("sh", "-c", "modprobe xt_TPROXY").CombinedOutput()
		if err != nil {
			if !strings.Contains(string(out), "not found") {
				return newError("failed to modprobe xt_TPROXY: " + string(out))
			}
			// modprobe失败，不支持xt_TPROXY方案
			return newError("not support xt_TPROXY: " + string(out))
		}
	}
	return
}

func CheckDohSupported(ver string) (err error) {
	if ver == "" {
		ver, err = GetV2rayServiceVersion()
		if err != nil {
			return newError("failed to get the version of v2ray-core")
		}
	}
	if greaterEqual, err := common.VersionGreaterEqual(ver, "4.22.0"); err != nil || !greaterEqual {
		return newError("the version of v2ray-core is lower than 4.22.0")
	}
	return
}
