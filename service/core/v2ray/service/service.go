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
	ver, err := where.GetV2rayServiceVersion()
	if err != nil {
		return fmt.Errorf("failed to get the version of v2ray-core")
	}
	if ver == "UnknownClient" && mustV2rayCore {
		return fmt.Errorf("v2fly/v2ray-core only feature")
	}
	if greaterEqual, err := common.VersionGreaterEqual(ver, version); err != nil || !greaterEqual {
		return fmt.Errorf("the version of v2ray-core is lower than %v", version)
	}
	return nil
}

func CheckDohSupported() (err error) {
	return isVersionSatisfied("4.22.0", false)
}

func CheckLogNoneSupported() (err error) {
	return isVersionSatisfied("4.20.0", false)
}

func CheckTcpDnsSupported() (err error) {
	return isVersionSatisfied("4.40.0", true)
}

func CheckQuicLocalDnsSupported() (err error) {
	return isVersionSatisfied("4.34.0", true)
}

func CheckFakednsOthersSupported() (err error) {
	return isVersionSatisfied("4.38.0", true)
}

func CheckFakednsAutoConfigureSupported() (err error) {
	return isVersionSatisfied("4.38.1", true)
}

func CheckBalancerSupported() (err error) {
	return isVersionSatisfied("4.4", false)
}

func CheckObservatorySupported() (err error) {
	return isVersionSatisfied("4.38.0", true)
}

func CheckHostsListSupported() (err error) {
	return isVersionSatisfied("4.37.3", true)
}

func CheckQueryStrategySupported() (err error) {
	return isVersionSatisfied("4.37.0", true)
}

func CheckMemconservativeSupported() (err error) {
	return isVersionSatisfied("4.39.0", true)
}

func CheckGrpcSupported() (err error) {
	return isVersionSatisfied("4.36.0", false)
}
