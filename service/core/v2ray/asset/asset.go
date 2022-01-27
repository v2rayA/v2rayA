package asset

import (
	"errors"
	"github.com/adrg/xdg"
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

func GetV2rayLocationAssetOverride() string {
	return filepath.Join(xdg.RuntimeDir, "v2raya")
}

func GetV2rayLocationAsset(filename string) (string, error) {
	variant, _, err := where.GetV2rayServiceVersion()
	if err != nil {
		variant = where.Unknown
	}
	var location string
	var folder string
	switch variant {
	case where.V2ray:
		location = "V2RAY_LOCATION_ASSET"
		folder = "v2ray"
	case where.Xray:
		location = "XRAY_LOCATION_ASSET"
		folder = "xray"
	default:
		location = "V2RAY_LOCATION_ASSET"
		folder = "v2ray"
	}
	location = os.Getenv(location)
	searchPaths := make([]string, 0)
	if location != "" {
		searchPaths = append(
			searchPaths,
			filepath.Join(location, filename),
		)
	}
	if runtime.GOOS != "windows" {
		searchPaths = append(
			searchPaths,
			filepath.Join("/usr/local/share", folder, filename),
			filepath.Join("/usr/share", folder, filename),
		)
	}
	if location != "" {
		for _, searchPath := range searchPaths {
			if _, err = os.Stat(searchPath); err != nil && errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return searchPath, nil
		}
		return searchPaths[0], nil
	} else {
		relpath := filepath.Join(folder, filename)
		fullpath, err := xdg.SearchDataFile(relpath)
		if err != nil {
			fullpath, err = xdg.DataFile(relpath)
			if err != nil {
				return "", err
			}
		}
		runtimepath, err := xdg.RuntimeFile(filepath.Join("v2raya", filename))
		if err != nil {
			return "", err
		}
		os.Remove(runtimepath)
		err = os.Symlink(fullpath, runtimepath)
		if err != nil {
			return "", err
		}
		return fullpath, err
	}
}

func IsV2rayAssetExists(filename string) bool {
	fullpath, err := GetV2rayLocationAsset(filename)
	if err != nil {
		return false
	}
	_, err = os.Stat(fullpath)
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
	fullpath, err := GetV2rayLocationAsset("LoyalsoldierSite.dat")
	if err != nil {
		return time.Now(), err
	}
	return files.GetFileModTime(fullpath)
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
