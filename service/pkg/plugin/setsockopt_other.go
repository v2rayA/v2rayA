//go:build !linux && !openbsd && !freebsd
// +build !linux,!openbsd,!freebsd

package plugin

import "syscall"

func SoMarkControl(c syscall.RawConn) error {
	return nil
}
func BindControl(c syscall.RawConn, laddr string, lport uint32) error {
	return nil
}