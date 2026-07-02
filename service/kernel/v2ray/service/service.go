package service

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
)

var CoreVersionMismatchError = fmt.Errorf("core version mismatch")

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
	if !IfTProxyModLoaded() && !common.IsDocker() && !iptables.IsNft() { //docker下无法判断，nft不需要
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

// CheckCoreVersionMatch checks whether the v2raya_core binary version matches
// the v2rayA service version. Development builds ("debug", "unstable") skip
// strict matching. In release builds, the two binaries are built from the same
// source tree with the same version string, so they must match exactly.
func CheckCoreVersionMatch() error {
	_, coreVer, err := where.GetV2rayServiceVersion()
	if err != nil {
		return fmt.Errorf("failed to get v2raya_core version: %v", err)
	}
	serviceVer := conf.Version

	// Development builds: skip strict matching
	var DevVersions = []string{"debug", "unstable"}
	if common.PrefixListSatisfyString(DevVersions, coreVer) != -1 ||
		common.PrefixListSatisfyString(DevVersions, serviceVer) != -1 {
		return nil
	}

	if coreVer != serviceVer {
		return fmt.Errorf(
			"%w: v2raya_core version %q does not match v2rayA version %q",
			CoreVersionMismatchError, coreVer, serviceVer,
		)
	}
	return nil
}
