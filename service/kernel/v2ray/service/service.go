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
				return fmt.Errorf("could not load xt_TPROXY: %v", string(out))
			}
			return fmt.Errorf("the kernel has no xt_TPROXY module (modprobe: %v); use redirect mode", string(out))
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
		// The two binaries are released together from one tree, so a mismatch
		// almost always means an xray-core or v2ray-core binary is installed
		// where v2raya_core belongs. Say that instead of two bare numbers.
		return common.Coded("CORE_VERSION_MISMATCH", fmt.Errorf(
			"%w: the core reports version %q but v2rayA is %q; v2rayA needs the v2raya_core binary of the same version, not xray-core or v2ray-core",
			CoreVersionMismatchError, coreVer, serviceVer,
		), map[string]interface{}{
			"core": coreVer,
			"app":  serviceVer,
		})
	}
	return nil
}
