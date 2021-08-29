package infra

import (
	"github.com/v2rayA/v2rayA/pkg/plugin/infra"
	"net"
	"syscall"
	"time"
)

func newDialer(laddr string, lport uint32, timeout time.Duration) (dialer *net.Dialer) {
	return &net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, c syscall.RawConn) error {
			return infra.BindControl(c, laddr, lport)
		},
	}
}
