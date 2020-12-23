package v2ray

import (
	"bytes"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"os/exec"
	"strings"
)

func IsV2rayServiceValid() bool {
	if !asset.IsGeoipExists() || !asset.IsGeositeExists() {
		return false
	}
	ver, err := where.GetV2rayServiceVersion()
	return err == nil && ver != ""
}

func IfTProxyModLoaded() bool {
	out, err := exec.Command("sh", "-c", "lsmod|grep xt_TPROXY").Output()
	return err == nil && len(bytes.TrimSpace(out)) > 0
}

func CheckAndProbeTProxy() (err error) {
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
		ver, err = where.GetV2rayServiceVersion()
		if err != nil {
			return newError("failed to get the version of v2ray-core")
		}
	}
	if greaterEqual, err := common.VersionGreaterEqual(ver, "4.22.0"); err != nil || !greaterEqual {
		return newError("the version of v2ray-core is lower than 4.22.0")
	}
	return
}
