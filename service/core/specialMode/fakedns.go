package specialMode

import (
	"github.com/v2rayA/v2rayA/db/configure"
)

func CouldUseFakeDns() bool {
	return configure.GetSettingNotNil().AntiPollution != configure.AntipollutionClosed
}

func ShouldUseFakeDns() bool {
	return configure.GetSettingNotNil().SpecialMode == configure.SpecialModeFakeDns
}
