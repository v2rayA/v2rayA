//go:build !windows

package v2ray

import (
	"os/exec"
	"syscall"
)

// setHookProcessGroup puts the hook in its own process group so a deadline
// kills the children it spawned, not only the shell.
func setHookProcessGroup(cmd *exec.Cmd) bool {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return true
}
