package common

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// SystemInfo contains system information for subscription headers
type SystemInfo struct {
	DeviceOS    string // e.g., "iOS", "Linux", "Windows", "Darwin"
	VersionOS   string // e.g., "18.3", "5.15.0", "10.0.19041"
	DeviceModel string // e.g., "iPhone 14 Pro Max", "Generic PC"
}

// GetSystemInfo returns system information for subscription headers
func GetSystemInfo() SystemInfo {
	info := SystemInfo{
		DeviceOS:    getDeviceOS(),
		VersionOS:   getVersionOS(),
		DeviceModel: getDeviceModel(),
	}
	return info
}

func getDeviceOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	case "freebsd":
		return "FreeBSD"
	case "openbsd":
		return "OpenBSD"
	default:
		// Capitalize first letter
		if len(runtime.GOOS) > 0 {
			return strings.ToUpper(runtime.GOOS[:1]) + runtime.GOOS[1:]
		}
		return runtime.GOOS
	}
}

func getVersionOS() string {
	hostInfo, err := host.Info()
	if err == nil && hostInfo.PlatformVersion != "" {
		// Extract version number (e.g., "5.15.0" from "5.15.0-generic")
		parts := strings.Fields(hostInfo.PlatformVersion)
		if len(parts) > 0 {
			return parts[0]
		}
		return hostInfo.PlatformVersion
	}
	
	// Fallback: try to get kernel version
	if hostInfo != nil && hostInfo.KernelVersion != "" {
		parts := strings.Fields(hostInfo.KernelVersion)
		if len(parts) > 0 {
			return parts[0]
		}
		return hostInfo.KernelVersion
	}
	
	// Final fallback
	return "Unknown"
}

func getDeviceModel() string {
	hostInfo, err := host.Info()
	if err == nil {
		// Try to get platform information
		if hostInfo.Platform != "" {
			if hostInfo.PlatformFamily != "" {
				return fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformFamily)
			}
			return hostInfo.Platform
		}
		if hostInfo.Hostname != "" {
			return hostInfo.Hostname
		}
	}
	
	// Fallback: use OS and architecture
	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

