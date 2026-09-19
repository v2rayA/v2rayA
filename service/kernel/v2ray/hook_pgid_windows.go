//go:build windows

package v2ray

import "os/exec"

func setHookProcessGroup(cmd *exec.Cmd) bool { return false }
