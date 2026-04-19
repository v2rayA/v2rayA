//go:build !windows
// +build !windows

package conf

import (
	"os"
	"path/filepath"
)

func loadPlatformEnv() error {
	return nil
}

func expandPlatformConfigPaths(p *Params) {}

func sanitizeConfigDirForPlatform(config string, _ bool) string {
	return config
}

func defaultConfigDir(isLite bool) string {
	if isLite {
		if userConfigDir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(userConfigDir, "v2raya")
		}
		return filepath.Join(os.TempDir(), "v2raya")
	}

	return "/etc/v2raya"
}
