package v2ray

import (
	"V2RayA/global"
	"V2RayA/tools"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

var v2rayLocationAsset *string

func GetV2rayLocationAsset() (s string) {
	if v2rayLocationAsset != nil {
		return *v2rayLocationAsset
	}
	switch global.ServiceControlMode {
	case global.DockerMode:
		return "/etc/v2ray"
	case global.SystemctlMode, global.ServiceMode:
		p, _ := GetV2rayServiceFilePath()
		out, err := exec.Command("sh", "-c", "cat "+p+"|grep Environment=V2RAY_LOCATION_ASSET").CombinedOutput()
		if err != nil {
			break
		}
		s = strings.TrimSpace(string(out))
		s = s[len("Environment=V2RAY_LOCATION_ASSET="):]
	}
	var err error
	if s == "" {
		//默认为v2ray运行目录
		s, err = getV2rayWorkingDir()
	}
	if err != nil {
		//再不行盲猜一个
		s = "/etc/v2ray"
	}
	v2rayLocationAsset = &s
	return
}

func getV2rayWorkingDir() (string, error) {
	out, err := exec.Command("sh", "-c", "type -p v2ray").CombinedOutput()
	if err != nil {
		out, err = exec.Command("sh", "-c", "which v2ray").CombinedOutput()
	}
	if err != nil {
		return "", errors.New(err.Error() + string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func IsH2yExists() bool {
	_, err := os.Stat(GetV2rayLocationAsset() + "/h2y.dat")
	if err != nil {
		return false
	}
	return true
}
func GetH2yModTime() (time.Time, error) {
	return tools.GetFileModTime(GetV2rayLocationAsset() + "/h2y.dat")
}
func IsCustomExists() bool {
	_, err := os.Stat(GetV2rayLocationAsset() + "/custom.dat")
	if err != nil {
		return false
	}
	return true
}
func GetCustomModTime() (time.Time, error) {
	return tools.GetFileModTime(GetV2rayLocationAsset() + "/custom.dat")
}
