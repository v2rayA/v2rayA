// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"errors"
	"net/netip"
)

// ConfigJSON is the "settings" object of a tun-mips inbound as the service
// writes it. Required fields are checked here so a missing one fails to
// load instead of silently becoming a zero value; the optional ones default
// to "no IPv6", "no exclusions" and "direct".
type ConfigJSON struct {
	Name             string   `json:"name"`
	MTU              uint32   `json:"mtu"`
	Address4         string   `json:"address4"`
	Address6         string   `json:"address6"`
	DnsTarget        string   `json:"dnsTarget"`
	ExcludeProcesses []string `json:"excludeProcesses"`
	SelfPids         []uint32 `json:"selfPids"`
	DirectTag        string   `json:"directTag"`
	UserLevel        uint32   `json:"userLevel"`
}

// Build validates and converts to the proto Config.
func (c *ConfigJSON) Build() (*Config, error) {
	if c.Name == "" {
		return nil, errors.New("tun-mips: name is required")
	}
	if c.MTU == 0 {
		return nil, errors.New("tun-mips: mtu is required")
	}
	if p, err := netip.ParsePrefix(c.Address4); err != nil || !p.Addr().Is4() {
		return nil, errors.New("tun-mips: address4 must be an IPv4 prefix such as 10.0.85.2/30")
	}
	if c.Address6 != "" {
		if p, err := netip.ParsePrefix(c.Address6); err != nil || !p.Addr().Is6() {
			return nil, errors.New("tun-mips: address6 must be an IPv6 prefix or empty")
		}
	}
	if _, err := netip.ParseAddrPort(c.DnsTarget); err != nil {
		return nil, errors.New("tun-mips: dnsTarget must be host:port")
	}
	return &Config{
		Name:             c.Name,
		Mtu:              c.MTU,
		Address4:         c.Address4,
		Address6:         c.Address6,
		DnsTarget:        c.DnsTarget,
		ExcludeProcesses: c.ExcludeProcesses,
		SelfPids:         c.SelfPids,
		DirectTag:        c.DirectTag,
		UserLevel:        c.UserLevel,
	}, nil
}
