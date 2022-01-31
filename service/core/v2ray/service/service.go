package service

import (
	"bytes"
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"os/exec"
	"strings"
)

func IsV2rayServiceValid() bool {
	if !asset.DoesV2rayAssetExist("geoip.dat") || !asset.DoesV2rayAssetExist("geosite.dat") {
		return false
	}
	_, ver, err := where.GetV2rayServiceVersion()
	return err == nil && ver != ""
}

func IfTProxyModLoaded() bool {
	out, err := exec.Command("sh", "-c", "lsmod|grep xt_TPROXY").Output()
	return err == nil && len(bytes.TrimSpace(out)) > 0
}

func CheckAndProbeTProxy() (err error) {
	if !IfTProxyModLoaded() && !common.IsDocker() { //docker下无法判断
		var out []byte
		out, err = exec.Command("sh", "-c", "modprobe xt_TPROXY").CombinedOutput()
		if err != nil {
			if !strings.Contains(string(out), "not found") {
				return fmt.Errorf("failed to modprobe xt_TPROXY: %v", string(out))
			}
			// modprobe失败，不支持xt_TPROXY方案
			return fmt.Errorf("not support xt_TPROXY: %v", string(out))
		}
	}
	return
}

func isVersionSatisfied(version string, mustV2rayCore bool) error {
	variant, ver, err := where.GetV2rayServiceVersion()
	if err != nil {
		return fmt.Errorf("failed to get the version of v2ray-core")
	}
	if variant != where.V2ray {
		if mustV2rayCore {
			return fmt.Errorf("v2fly/v2ray-core only feature")
		} else {
			// do not check the version for non-v2ray core
			return nil
		}
	}
	if greaterEqual, err := common.VersionGreaterEqual(ver, version); err != nil || !greaterEqual {
		return fmt.Errorf("the version of v2ray-core is lower than %v", version)
	}
	return nil
}

func CheckV5() (err error) {
	return isVersionSatisfied("5.0.0", true)
}
