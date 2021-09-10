package asset

import (
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os"
	"path"
	"path/filepath"
	"time"
)

func GetV2rayLocationAsset() (s string) {
	var candidates = []string{
		"/usr/local/share/v2ray",
		"/usr/share/v2ray",
		"/opt/share/v2ray",
		"/usr/local/share/xray",
		"/usr/share/xray",
		"/opt/share/xray",
	}
	var is bool
	if ver, err := where.GetV2rayServiceVersion(); err == nil {
		if is, err = common.VersionGreaterEqual(ver, "4.27.1"); is {
			for _, c := range candidates {
				if _, err := os.Stat(c); os.IsNotExist(err) {
					continue
				}
				if _, err := os.Stat(path.Join(c, "geoip.dat")); os.IsNotExist(err) {
					continue
				}
				s = c
				break
			}
		}
	}
	if s == "" {
		// set as v2rayA config directory
		s = conf.GetEnvironmentConfig().Config
	}
	return
}

func IsGFWListExists() bool {
	_, err := os.Stat(path.Join(GetV2rayLocationAsset(), "LoyalsoldierSite.dat"))
	if err != nil {
		return false
	}
	return true
}
func IsGeoipExists() bool {
	_, err := os.Stat(path.Join(GetV2rayLocationAsset(), "geoip.dat"))
	if err != nil {
		return false
	}
	return true
}
func IsGeoipOnlyCnPrivateExists() bool {
	_, err := os.Stat(path.Join(GetV2rayLocationAsset(), "geoip-only-cn-private.dat"))
	if err != nil {
		return false
	}
	return true
}
func IsGeositeExists() bool {
	_, err := os.Stat(path.Join(GetV2rayLocationAsset(), "geosite.dat"))
	if err != nil {
		return false
	}
	return true
}
func GetGFWListModTime() (time.Time, error) {
	return files.GetFileModTime(path.Join(GetV2rayLocationAsset(), "LoyalsoldierSite.dat"))
}
func IsCustomExists() bool {
	_, err := os.Stat(path.Join(GetV2rayLocationAsset(), "custom.dat"))
	if err != nil {
		return false
	}
	return true
}

func GetConfigBytes() (b []byte, err error) {
	b, err = os.ReadFile(GetV2rayConfigPath())
	if err != nil {
		log.Warn("failed to get config: %v", err)
		return
	}
	b = jsonc.ToJSON(b)
	return
}

func GetV2rayConfigPath() (p string) {
	return path.Join(conf.GetEnvironmentConfig().Config, "config.json")
}

func GetV2rayConfigDirPath() (p string) {
	return conf.GetEnvironmentConfig().V2rayConfigDirectory
}

func LoyalsoldierSiteDatExists() bool {
	if info, err := os.Stat(filepath.Join(GetV2rayLocationAsset(), "LoyalsoldierSite.dat")); err == nil && !info.IsDir() {
		return true
	}
	return false
}
