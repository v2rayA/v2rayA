package v2ray

import (
	"V2RayA/global"
	"V2RayA/tools"
	"os"
	"time"
)

func IsH2yExists() bool {
	_, err := os.Stat(global.V2RAY_LOCATION_ASSET + "/h2y.dat")
	if err != nil {
		return false
	}
	return true
}
func GetH2yModTime() (time.Time,error) {
	return tools.GetFileModTime(global.V2RAY_LOCATION_ASSET + "/h2y.dat")
}
func IsCustomExists() bool {
	_, err := os.Stat(global.V2RAY_LOCATION_ASSET + "/custom.dat")
	if err != nil {
		return false
	}
	return true
}
func GetCustomModTime() (time.Time,error) {
	return tools.GetFileModTime(global.V2RAY_LOCATION_ASSET + "/custom.dat")
}