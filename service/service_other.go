//go:build !windows
// +build !windows

package main

import (
	"runtime"
)

// tryRunAsService always returns false on non-Windows platforms
func tryRunAsService() bool {
	return false
}

// checkPlatformSpecific performs platform-specific checks
func checkPlatformSpecific() error {
	if runtime.GOOS == "linux" {
		checkTProxySupportability()
	}
	return nil
}

// runAsService is not needed on non-Windows platforms, but kept for compilation consistency
func runAsService(isDebug bool) error {
	// This function should not be called on non-Windows platforms
	return nil
}
