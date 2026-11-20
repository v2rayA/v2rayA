//go:build !linux

package v2ray

import "os/exec"

func setCoreParentDeathSignal(cmd *exec.Cmd) {}
