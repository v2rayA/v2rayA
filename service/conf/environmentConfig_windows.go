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
	"golang.org/x/sys/windows/svc"

	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/pkg/util/privilege"
)

var platformEnvOnce sync.Once

// expandWindowsEnv 使用 Windows API 展开 %VAR% 风格的环境变量引用
func expandWindowsEnv(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	utf16Src, err := windows.UTF16FromString(s)
	if err != nil {
		return s
	}
	// 先探测所需缓冲区大小
	n, _ := windows.ExpandEnvironmentStrings(&utf16Src[0], nil, 0)
	if n == 0 {
		return s
	}
	buf := make([]uint16, n)
	n, err = windows.ExpandEnvironmentStrings(&utf16Src[0], &buf[0], n)
	if err != nil || n == 0 {
		return s
	}
	return windows.UTF16ToString(buf[:n])
}

// expandPlatformConfigPaths 展开 Params 中所有路径字段里的 Windows %VAR% 环境变量
func expandPlatformConfigPaths(p *Params) {
	p.Config = expandWindowsEnv(p.Config)
	p.V2rayBin = expandWindowsEnv(p.V2rayBin)
	p.V2rayConfigDirectory = expandWindowsEnv(p.V2rayConfigDirectory)
	p.V2rayAssetsDirectory = expandWindowsEnv(p.V2rayAssetsDirectory)
	p.LogFile = expandWindowsEnv(p.LogFile)
	p.TransparentHook = expandWindowsEnv(p.TransparentHook)
	p.CoreHook = expandWindowsEnv(p.CoreHook)
	p.PluginManager = expandWindowsEnv(p.PluginManager)
	p.WinEnvFile = expandWindowsEnv(p.WinEnvFile)
	p.WebDir = expandWindowsEnv(p.WebDir)
}

func loadPlatformEnv() error {
	var loadErr error
	platformEnvOnce.Do(func() {
		// 检查是否作为 Windows 服务运行
		isService, err := svc.IsWindowsService()
		if err != nil {
			log.Warn("Failed to check if running as Windows service: %v", err)
			return
		}

		// 只在作为服务运行时才加载环境变量文件
		if !isService {
			return
		}

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

			// 展开值中 %VAR% 风格的 Windows 环境变量引用
			value = expandWindowsEnv(value)

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

// sanitizeConfigDirForPlatform avoids creating Linux-style paths like C:\\etc\\v2raya on Windows
func sanitizeConfigDirForPlatform(config string, isLite bool) string {
	cleaned := filepath.Clean(config)
	normalized := strings.ToLower(filepath.ToSlash(cleaned))

	// Detect Linux default path on Windows where volume is missing and path starts with /etc/v2raya
	if filepath.VolumeName(cleaned) == "" && strings.HasPrefix(normalized, "/etc/v2raya") {
		fallback := defaultConfigDir(isLite)
		log.Warn("Detected Linux-style config path on Windows (%s); falling back to %s", config, fallback)
		return fallback
	}

	return config
}
