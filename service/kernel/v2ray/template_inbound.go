package v2ray

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"sort"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func FilterIPs(ips []string) []string {
	var ret []string
	for _, ip := range ips {
		if net.ParseIP(ip).To4() != nil {
			ret = append(ret, ip)
		}
	}
	if !iptables.IsIPv6Supported() {
		return ret
	}
	for _, ip := range ips {
		if net.ParseIP(ip).To4() == nil {
			ret = append(ret, ip)
		}
	}
	return ret
}
func GetBestLocalIP(preferIPv6 bool) (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var bestIP net.IP

	for _, iface := range interfaces {
		// Skip interfaces that are down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Skip TUN interfaces (typically named tun0, utun0, etc.)
		ifName := strings.ToLower(iface.Name)
		if strings.Contains(ifName, "tun") || strings.Contains(ifName, "tap") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipnet.IP

			// Skip loopback
			if ip.IsLoopback() {
				continue
			}

			// Skip APIPA (169.254.x.x)
			if ip.To4() != nil && ip.To4()[0] == 169 && ip.To4()[1] == 254 {
				continue
			}

			if preferIPv6 {
				if ip.To4() == nil && ip.To16() != nil {
					// IPv6 address
					// Skip link-local for now (fe80::), prefer global
					if !ip.IsLinkLocalUnicast() {
						return ip, nil
					}
					if bestIP == nil {
						bestIP = ip
					}
				}
			} else {
				if ip.To4() != nil {
					// IPv4 address
					return ip, nil
				}
			}
		}
	}

	if bestIP != nil {
		return bestIP, nil
	}

	if preferIPv6 {
		return nil, errors.New("no suitable IPv6 address found")
	}
	return nil, errors.New("no suitable IPv4 address found")
}

// GetLinkLocalIPv6 returns a link-local IPv6 address as fallback
func GetLinkLocalIPv6() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Skip TUN interfaces
		ifName := strings.ToLower(iface.Name)
		if strings.Contains(ifName, "tun") || strings.Contains(ifName, "tap") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipnet.IP
			if ip.To4() == nil && ip.To16() != nil && ip.IsLinkLocalUnicast() {
				return ip, nil
			}
		}
	}

	return nil, errors.New("no link-local IPv6 address found")
}

func GenerateIdFromAccounts() (id string, err error) {
	accounts, err := configure.GetAccounts()
	if err != nil {
		return "", err
	}
	sort.Slice(accounts, func(i, j int) bool {
		if accounts[i][0] == accounts[j][0] {
			return accounts[i][1] < accounts[j][1]
		}
		return accounts[i][0] < accounts[j][0]
	})
	h := sha256.New()
	for _, account := range accounts {
		h.Write([]byte(account[0]))
		h.Write([]byte(account[1]))
	}
	id = common.StringToUUID5(hex.EncodeToString(h.Sum(nil)))
	return id, nil
}

func SetVmessInbound(vmess *coreObj.Inbound) (err error) {
	id, err := GenerateIdFromAccounts()
	if err != nil {
		return err
	}
	vmess.Settings.Clients = []coreObj.VlessClient{{Id: id}}
	return nil
}

func (t *Template) setInbound(setting *configure.Setting) error {
	p := configure.GetPortsNotNil()
	if p != nil {
		t.Inbounds[0].Port = p.Socks5
		t.Inbounds[1].Port = p.Http
		t.Inbounds[2].Port = p.Socks5WithPac
		t.Inbounds[3].Port = p.HttpWithPac
		listenAddr := "127.0.0.1"
		if t.Setting.PortSharing {
			listenAddr = "0.0.0.0"
		}
		for i := 0; i < 5 && i < len(t.Inbounds); i++ {
			t.Inbounds[i].Listen = listenAddr
		}
		vmess := &t.Inbounds[4]
		vmess.Port = p.Vmess
		if p.Vmess > 0 {
			if err := SetVmessInbound(vmess); err != nil {
				return err
			}
		}
	}
	// remove those inbounds with zero port number
	for i := len(t.Inbounds) - 1; i >= 0; i-- {
		if t.Inbounds[i].Port == 0 {
			t.Inbounds = append(t.Inbounds[:i], t.Inbounds[i+1:]...)
		}
	}
	// Append user-defined custom inbounds (SOCKS / HTTP only)
	listenAddrForCustom := "127.0.0.1"
	if t.Setting != nil && t.Setting.PortSharing {
		listenAddrForCustom = "0.0.0.0"
	}
	for _, ci := range configure.GetCustomInbounds() {
		if ci.Port <= 0 || (ci.Protocol != "socks" && ci.Protocol != "http") {
			continue
		}
		ib := coreObj.Inbound{
			Port:     ci.Port,
			Protocol: ci.Protocol,
			Listen:   listenAddrForCustom,
			Tag:      ci.Tag,
		}
		if ci.Protocol == "socks" {
			ib.Settings = &coreObj.InboundSettings{UDP: true}
		}
		t.Inbounds = append(t.Inbounds, ib)

		// Generate per-inbound routing rules based on the bound outbound group
		if ci.Outbound != "" && ci.OutboundType != "" {
			switch ci.OutboundType {
			case "direct":
				// Direct single outbound: route all traffic from this inbound to the bound outbound group
				t.Routing.Rules = append(t.Routing.Rules, coreObj.RoutingRule{
					Type:        "field",
					OutboundTag: ci.Outbound,
					InboundTag:  []string{ci.Tag},
				})
			case "routingA":
				// RoutingA rules: parse and apply with InboundTag filter
				if ci.RoutingARules != "" {
					rules, err := parseRoutingARules(ci.RoutingARules, ci.Tag)
					if err != nil {
						log.Warn("setInbound: failed to parse RoutingA rules for inbound %s: %v", ci.Tag, err)
					} else {
						t.Routing.Rules = append(t.Routing.Rules, rules...)
					}
				}
			}
		}
	}
	if IsTransparentOn(t.Setting) {
		switch t.Setting.TransparentType {
		case configure.TransparentTproxy, configure.TransparentRedirect:
			t.AppendDokodemoTProxy(string(t.Setting.TransparentType), 52345, "transparent")
		case configure.TransparentSystemProxy:
			t.Inbounds = append(t.Inbounds, coreObj.Inbound{
				Port:     52345,
				Protocol: "http",
				Listen:   "127.0.0.1",
				Tag:      "transparent",
			}, coreObj.Inbound{
				Port:     52306,
				Protocol: "socks",
				Listen:   "127.0.0.1",
				Settings: &coreObj.InboundSettings{
					UDP: true,
				},
				Tag: "transparent-socks",
			})
		case configure.TransparentTun:
			t.Inbounds = append(t.Inbounds, coreObj.Inbound{
				Port:     tinytunSocksPort,
				Protocol: "socks",
				Listen:   "127.0.0.1",
				Settings: &coreObj.InboundSettings{
					UDP: true,
				},
				Tag: "transparent",
			})
			// TinyTun v0.0.2+ handles DNS routing natively via its own DNS groups.
			// The former dns-in-tun dokodemo-door (127.0.0.1:6053) is no longer needed;
			// v2ray acts as a pure SOCKS5 forwarder for non-DNS traffic.
		}

	}
	if ShouldLocalDnsListen() {
		log.Trace("DNS module handles DNS on port 52353, no dns-in needed")
	}

	// Set up domain sniffing
	if setting.InboundSniffing != configure.InboundSniffingDisable && setting.InboundSniffing != "" {
		enableSniffingRouteOnly := configure.GetSettingNotNil().RouteOnly
		domainsExcluded := splitNonEmptyLines(configure.GetDomainsExcluded())
		for i := len(t.Inbounds) - 1; i >= 0; i-- {
			if setting.InboundSniffing == configure.InboundSniffingHttpTLS {
				t.Inbounds[i].Sniffing.DestOverride = []string{"http", "tls"}
			} else {
				t.Inbounds[i].Sniffing.DestOverride = []string{"http", "tls", "quic"}
			}
			t.Inbounds[i].Sniffing.DomainsExcluded = domainsExcluded
			t.Inbounds[i].Sniffing.Enabled = true
			t.Inbounds[i].Sniffing.RouteOnly = enableSniffingRouteOnly
		}

	}
	return nil
}

func splitNonEmptyLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
