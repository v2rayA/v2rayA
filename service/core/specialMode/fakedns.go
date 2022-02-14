package specialMode

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
)

func CouldUseFakeDns() bool {
	variant, ver, err := where.GetV2rayServiceVersion()
	if err != nil {
		ver = "0.0.0"
	}
	if variant == where.V2ray {
		if fakeDnsValid, _ := common.VersionGreaterEqual(ver, "4.35.0"); !fakeDnsValid {
			return false
		}
	}
	return configure.GetSettingNotNil().AntiPollution != configure.AntipollutionClosed
}

func ShouldUseFakeDns() bool {
	return configure.GetSettingNotNil().SpecialMode == configure.SpecialModeFakeDns
}
