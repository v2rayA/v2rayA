//go:build !with_gvisor

package tun

import (
	tun "github.com/sagernet/sing-tun"
)

type gvisorCloser struct {
	stack tun.Stack
}

func (gc gvisorCloser) Close() error {
	return gc.stack.Close()
}
