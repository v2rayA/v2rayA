package asset

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	url2 "net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/adrg/xdg"
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func GetV2rayLocationAssetOverride() string {
	if assetDir := conf.GetEnvironmentConfig().V2rayAssetsDirectory; assetDir != "" {
		return assetDir
	}
	if assetDir := os.Getenv("V2RAY_LOCATION_ASSET"); assetDir != "" {
		return assetDir
	}
	if runtime.GOOS != "windows" {
		return filepath.Join(xdg.RuntimeDir, "v2raya")
	} else {
		return conf.GetEnvironmentConfig().Config
	}
}

func GetV2rayLocationAsset(filename string) (string, error) {
	// All variants use XRAY_LOCATION_ASSET; dat files are stored under
	// v2raya's own XDG data subdirectory ("v2raya/"), not under "xray/".
	const envKey = "XRAY_LOCATION_ASSET"
	const folder = "v2raya"

	location := os.Getenv(envKey)
	// check if XRAY_LOCATION_ASSET is set
	if location != "" {
		// add XRAY_LOCATION_ASSET to search path
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
			if _, err := os.Stat(searchPath); err != nil && errors.Is(err, fs.ErrNotExist) {
				continue
			}
			// return the first path that exists
			return searchPath, nil
		}
		// or download asset into XRAY_LOCATION_ASSET
		return searchPaths[0], nil
	} else {
		if runtime.GOOS != "windows" {
			// search XDG data directories on non windows platform
			// symlink all assets into XDG_RUNTIME_DIR so xray-core can find them
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

// coreAssets are the files v2raya_core itself loads by name; it is given
// XRAY_LOCATION_ASSET and looks nowhere else.
var coreAssets = []string{"geoip.dat", "geosite.dat", "LoyalsoldierSite.dat", "geoip-only-cn-private.dat"}

// EnsureCoreAssets links the dat files the core needs into assetDir when they
// live somewhere else. v2rayA only ever created those links as a side effect
// of looking a file up for itself, and on a system whose XDG runtime directory
// is cleared between sessions the core then started with an empty asset
// directory and failed with "failed to open geosite.dat" while the files sat
// in /usr/share/v2raya all along.
func EnsureCoreAssets(assetDir string) {
	if runtime.GOOS == "windows" || assetDir == "" {
		return
	}
	for _, name := range coreAssets {
		target := filepath.Join(assetDir, name)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		source := findAssetOutsideDir(name, assetDir)
		if source == "" {
			continue
		}
		if err := os.MkdirAll(assetDir, 0755); err != nil {
			log.Warn("cannot create the asset directory %v: %v", assetDir, err)
			return
		}
		_ = os.Remove(target)
		if err := os.Symlink(source, target); err != nil {
			log.Warn("cannot link %v into %v: %v", source, assetDir, err)
			continue
		}
		log.Info("linked %v into the core asset directory %v", source, assetDir)
	}
}

// findAssetOutsideDir returns the first readable copy of name that is not
// already in assetDir, searching the XDG data directories and the two system
// directories a distribution package installs into.
func findAssetOutsideDir(name string, assetDir string) string {
	var candidates []string
	if p, err := xdg.SearchDataFile(filepath.Join("v2raya", name)); err == nil {
		candidates = append(candidates, p)
	}
	candidates = append(candidates,
		filepath.Join("/usr/local/share", "v2raya", name),
		filepath.Join("/usr/share", "v2raya", name),
	)
	for _, c := range candidates {
		if filepath.Dir(c) == filepath.Clean(assetDir) {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
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

func GFWListMissingError() error {
	dir := GetV2rayLocationAssetOverride()
	return common.Coded("GFWLIST_MISSING", fmt.Errorf("GFWList mode needs LoyalsoldierSite.dat, which is missing from %s; update GFWList first", dir), map[string]interface{}{"dir": dir})
}

func GetGFWListModTime() (time.Time, error) {
	fullpath, err := GetV2rayLocationAsset("LoyalsoldierSite.dat")
	if err != nil {
		return time.Now(), err
	}
	return files.GetFileModTime(fullpath)
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

func GetNftablesConfigPath() (p string) {
	return path.Join(conf.GetEnvironmentConfig().Config, "v2raya.nft")
}

func Download(url string, to string) (err error) {
	log.Info("Downloading %v to %v", url, to)
	host := "unknown host"
	if u, parseErr := url2.Parse(url); parseErr == nil && u.Hostname() != "" {
		host = u.Hostname()
	}
	status := ""
	c := http.Client{Timeout: 90 * time.Second}
	resp, err := c.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			defer resp.Body.Close()
			status = resp.Status
			err = fmt.Errorf("download from %s failed: HTTP %s", host, status)
		} else {
			// The request never got a reply, so there is no status to show;
			// a message that ends in "HTTP )" tells the user nothing.
			reason := err
			for {
				inner := errors.Unwrap(reason)
				if inner == nil {
					break
				}
				reason = inner
			}
			return common.Coded("ASSET_UNREACHABLE", err, map[string]interface{}{
				"host":   host,
				"detail": reason.Error(),
			})
		}
		return common.Coded("ASSET_DOWNLOAD_FAILED", err, map[string]interface{}{
			"host":   host,
			"status": status,
		})
	}
	defer resp.Body.Close()
	status = resp.Status
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.Coded("ASSET_DOWNLOAD_FAILED", err, map[string]interface{}{
			"host":   host,
			"status": status,
		})
	}
	if err = os.WriteFile(to, b, 0644); err != nil {
		return common.Coded("ASSET_DOWNLOAD_FAILED", err, map[string]interface{}{
			"host":   host,
			"status": status,
		})
	}
	return nil
}
