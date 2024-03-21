package outbound

import (
	"github.com/v2rayA/v2rayA/core/singBox/net"
)

type Direct struct {
	OverrideAddress string `json:"override_address,omitempty"`
	OverridePort    int    `json:"override_port,omitempty"`
	ProxyProtocol   int    `json:"proxy_protocol,omitempty"`
	net.Dial
}

func (o Direct) outbound() {}
