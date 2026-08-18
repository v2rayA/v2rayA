//go:build !tinytun

package v2ray

import "fmt"

// tinytunSocksPort is the SOCKS5 port reserved for TinyTun traffic.
// Kept in the stub so that v2rayTmpl.go can reference it unconditionally.
const tinytunSocksPort = 52345

// IsTinyTunEnabled reports whether TinyTun support was compiled into this binary.
func IsTinyTunEnabled() bool { return false }

// GetTinyTunBinPath is a stub that returns an error when TinyTun is not compiled in.
func GetTinyTunBinPath() (string, error) {
	return "", fmt.Errorf("TinyTun support is not compiled into this binary (build with -tags tinytun)")
}

// startTinyTun is a stub that returns an error when TinyTun is not compiled in.
func startTinyTun(_ *Template) error {
	return fmt.Errorf("TinyTun support is not compiled into this binary (build with -tags tinytun)")
}

// stopTinyTun is a no-op stub when TinyTun is not compiled in.
func stopTinyTun() {}
