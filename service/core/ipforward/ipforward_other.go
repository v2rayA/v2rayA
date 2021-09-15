//go:build windows
// +build windows

package ipforward

func IsIpForwardOn() bool {
	return true
}

func WriteIpForward(on bool) (err error) {
	return nil
}
