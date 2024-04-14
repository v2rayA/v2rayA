//go:build !with_gvisor

package tun

import (
	tun "github.com/sagernet/sing-tun"
)

type gvisorWaiter struct {
	stack tun.Stack
}

func (gc gvisorWaiter) Wait() {
}
