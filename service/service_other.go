//go:build !windows
// +build !windows

package main

import (
	"runtime"
)

// tryRunAsService 非 Windows 平台始终返回 false
func tryRunAsService() bool {
	return false
}

// checkPlatformSpecific 平台特定检查
func checkPlatformSpecific() error {
	if runtime.GOOS == "linux" {
		checkTProxySupportability()
	}
	return nil
}

// runAsService 非 Windows 平台不需要此函数，但为了编译一致性保留
func runAsService(isDebug bool) error {
	// 在非 Windows 平台此函数不应被调用
	return nil
}
