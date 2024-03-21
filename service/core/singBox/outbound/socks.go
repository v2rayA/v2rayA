package outbound

import (
	"github.com/v2rayA/v2rayA/core/singBox/net"
)

type SOCKS struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Version    string `json:"version,omitempty"`
	Username   int    `json:"username,omitempty"`
	Password   int    `json:"password,omitempty"`
	Network    int    `json:"network,omitempty"`
	UdpOverTcp any    `json:"udp_over_tcp,omitempty"`
	net.Dial
}

func (o SOCKS) outbound() {}
