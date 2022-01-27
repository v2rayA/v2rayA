package asset

import (
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os"
	"path"
	"path/filepath"
	"time"
	"github.com/adrg/xdg"
)

func GetV2rayLocationAsset(filename string) string {
	relpath := filepath.Join("v2ray", filename)
	fullpath, err := xdg.SearchDataFile(relpath)
	if err != nil {
		fullpath, err = xdg.DataFile(relpath)
		if err != nil {
			// unlikely, none of the xdg data dirs are writable
			panic(err)
		}
	}
	return fullpath
}

func IsV2rayAssetExists(filename string) bool {
	_, err := os.Stat(GetV2rayLocationAsset(filename))
	if err != nil {
		return false
	}
	return true
}

func IsGFWListExists() bool {
	return IsV2rayAssetExists("LoyalsoldierSite.dat")
}
func IsGeoipExists() bool {
	return IsV2rayAssetExists("geoip.dat")
}
func IsGeoipOnlyCnPrivateExists() bool {
	return IsV2rayAssetExists("geoip-only-cn-private.dat")
}
func IsGeositeExists() bool {
	return IsV2rayAssetExists("geosite.dat")
}
func GetGFWListModTime() (time.Time, error) {
	return files.GetFileModTime(GetV2rayLocationAsset("LoyalsoldierSite.dat"))
}
func IsCustomExists() bool {
	return IsV2rayAssetExists("custom.dat")
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
	return IsV2rayAssetExists("LoyalsoldierSite.dat")
}
