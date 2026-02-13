//go:build windows
// +build windows

package conf

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sys/windows"

	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/pkg/util/privilege"
)

var platformEnvOnce sync.Once

func loadPlatformEnv() error {
	var loadErr error
	platformEnvOnce.Do(func() {
		envFilePath := os.Getenv("V2RAYA_WIN_ENVFILE")
		if envFilePath == "" {
			for i, arg := range os.Args {
				if arg == "--win-envfile" && i+1 < len(os.Args) {
					envFilePath = os.Args[i+1]
					break
				}
			}
		}

		if envFilePath == "" {
			return
		}

		absPath, err := filepath.Abs(envFilePath)
		if err != nil {
			loadErr = fmt.Errorf("failed to get absolute path: %w", err)
			return
		}

		file, err := os.Open(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn("Windows env file not found: %s (skipping)", absPath)
				return
			}
			loadErr = fmt.Errorf("failed to open env file: %w", err)
			return
		}
		defer file.Close()

		log.Info("Loading Windows env file: %s", absPath)

		scanner := bufio.NewScanner(file)
		lineNum := 0
		loadedCount := 0

		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				log.Warn("Invalid line %d in env file: %s", lineNum, line)
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if len(value) >= 2 {
				if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
					value = value[1 : len(value)-1]
				}
			}

			if err := os.Setenv(key, value); err != nil {
				log.Warn("Failed to set env var %s: %v", key, err)
			} else {
				loadedCount++
			}
		}

		if err := scanner.Err(); err != nil {
			loadErr = fmt.Errorf("failed to read env file: %w", err)
			return
		}

		log.Info("Loaded %d environment variables from %s", loadedCount, absPath)
	})
	return loadErr
}

func isSystemAccount() bool {
	token := windows.Token(0)

	if systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid); err == nil {
		if member, err := token.IsMember(systemSID); err == nil && member {
			return true
		}
	}

	if u, err := user.Current(); err == nil {
		uname := strings.ToUpper(u.Username)
		if strings.Contains(uname, "SYSTEM") {
			return true
		}
	}

	if envUser := strings.ToUpper(os.Getenv("USERNAME")); strings.Contains(envUser, "SYSTEM") {
		return true
	}

	return false
}

func defaultConfigDir(isLite bool) string {
	if isLite {
		if appData := os.Getenv("AppData"); appData != "" {
			return filepath.Join(appData, "v2rayA")
		}
		if userConfigDir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(userConfigDir, "v2rayA")
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "AppData", "Roaming", "v2rayA")
		}
		return filepath.Join(os.TempDir(), "v2rayA")
	}

	base := os.Getenv("ProgramData")
	account := ""

	if isSystemAccount() {
		account = "SYSTEM"
	} else if privilege.IsRootOrAdmin() {
		account = "Administrator"
	}

	if account != "" && base != "" {
		return filepath.Join(base, account, "v2rayA")
	}

	if userConfigDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(userConfigDir, "v2rayA")
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "AppData", "Roaming", "v2rayA")
	}

	return filepath.Join(os.TempDir(), "v2rayA")
}
