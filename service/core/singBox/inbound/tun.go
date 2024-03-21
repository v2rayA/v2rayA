package inbound

import "github.com/v2rayA/v2rayA/core/singBox/net"

type Tun struct {
	InterfaceName            string   `json:"interface_name,omitempty"`
	Inet4Address             string   `json:"inet4_address"`
	Inet6Address             string   `json:"inet6_address,omitempty"`
	Mtu                      int      `json:"mtu"`
	AutoRoute                bool     `json:"auto_route"`
	StrictRoute              bool     `json:"strict_route"`
	Inet4RouteAddress        []string `json:"inet4_route_address,omitempty"`
	Inet6RouteAddress        []string `json:"inet6_route_address,omitempty"`
	Inet4RouteExcludeAddress []string `json:"inet4_route_exclude_address,omitempty"`
	Inet6RouteExcludeAddress []string `json:"inet6_route_exclude_address,omitempty"`
	EndpointIndependentNat   bool     `json:"endpoint_independent_nat"`
	Stack                    string   `json:"stack"`
	net.Listen
}

func (i Tun) inbound() {}
