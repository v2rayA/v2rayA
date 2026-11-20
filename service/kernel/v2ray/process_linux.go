package v2ray

import (
	"os/exec"
	"syscall"
)

func setCoreParentDeathSignal(cmd *exec.Cmd) {
	// SIGKILL cannot run v2rayA's cleanup, so let Linux terminate its core.
	// shortcut: exec with set-ID or file capabilities clears this; revisit if such cores are supported.
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
}
