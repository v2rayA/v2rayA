package asset

import (
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

func GetV2rayLocationAssetOverride() string {
	if runtime.GOOS != "windows" {
		return filepath.Join(xdg.RuntimeDir, "v2raya")
	} else {
		return conf.GetEnvironmentConfig().Config
	}
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
	// check if V2RAY_LOCATION_ASSET is set
	if location != "" {
		// add V2RAY_LOCATION_ASSET to search path
		searchPaths := []string{
			filepath.Join(location, filename),
		}
		// additional paths for non windows platforms
		if runtime.GOOS != "windows" {
			searchPaths = append(
				searchPaths,
				filepath.Join("/usr/local/share", folder, filename),
				filepath.Join("/usr/share", folder, filename),
			)
		}
		for _, searchPath := range searchPaths {
			if _, err = os.Stat(searchPath); err != nil && errors.Is(err, fs.ErrNotExist) {
				continue
			}
			// return the first path that exists
			return searchPath, nil
		}
		// or download asset into V2RAY_LOCATION_ASSET
		return searchPaths[0], nil
	} else {
		if runtime.GOOS != "windows" {
			// search XDG data directories on non windows platform
			// symlink all assets into XDG_RUNTIME_DIR
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
		} else {
			// fallback to the old behavior of using only config dir on windows
			return filepath.Join(conf.GetEnvironmentConfig().Config, filename), nil
		}
	}
}

func DoesV2rayAssetExist(filename string) bool {
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

func GetGFWListModTime() (time.Time, error) {
	fullpath, err := GetV2rayLocationAsset("LoyalsoldierSite.dat")
	if err != nil {
		return time.Now(), err
	}
	return files.GetFileModTime(fullpath)
}
func IsCustomExists() bool {
	return DoesV2rayAssetExist("custom.dat")
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

func Download(url string, to string) (err error) {
	log.Info("Downloading %v to %v", url, to)
	c := http.Client{Timeout: 90 * time.Second}
	resp, err := c.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			defer resp.Body.Close()
			err = fmt.Errorf("code: %v %v", resp.StatusCode, resp.Status)
		}
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(to, b, 0644)
}
