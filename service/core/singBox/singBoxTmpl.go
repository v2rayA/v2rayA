package singBox

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/core/singBox/inbound"
	"github.com/v2rayA/v2rayA/core/singBox/net"
	"github.com/v2rayA/v2rayA/core/singBox/outbound"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
)

type Template struct {
	Log       *Log       `json:"log,omitempty"`
	DNS       *DNS       `json:"dns,omitempty"`
	NTP       *NTP       `json:"ntp,omitempty"`
	Route     *Route     `json:"route"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
}

func (t *Template) appendDNS() {
	t.DNS.Servers = append(t.DNS.Servers, DNSServer{
		Tag:     "out_dns",
		Address: "1.1.1.1",
	}, DNSServer{
		Tag:     "local",
		Address: "119.29.29.29",
		Detour:  "direct",
	}, DNSServer{
		Tag:     "block",
		Address: "rcode://success",
	})
	t.DNS.Rules = append(t.DNS.Rules, DNSRule{
		GeoSite:      []string{"cn"},
		Server:       "local",
		DisableCache: true,
	}, DNSRule{
		GeoSite:      []string{"category-ads-all"},
		Server:       "block",
		DisableCache: true,
	})
	t.DNS.Strategy = "ipv4_only"
}

func (t *Template) appendRoute() {
	processNames := where.ServiceNameList
	if runtime.GOOS == "windows" {
		for i, name := range processNames {
			if !strings.HasSuffix(strings.ToLower(name), ".exe") {
				processNames[i] = name + ".exe"
			}
		}
	}
	t.Route.AutoDetectInterface = true
	t.Route.Rules = append(t.Route.Rules, RouteRule{
		Inbound:  []string{"dns_in"},
		Outbound: "dns_out",
	}, RouteRule{
		Protocol: []string{"dns"},
		Outbound: "dns_out",
	}, RouteRule{
		Network:  []string{"udp"},
		Port:     []int{135, 137, 138, 139, 5353},
		Outbound: "block",
	}, RouteRule{
		IPCidr:   []string{"224.0.0.0/3", "ff00::/8"},
		Outbound: "block",
	}, RouteRule{
		SourceIPCidr: []string{"224.0.0.0/3", "ff00::/8"},
		Outbound:     "block",
	}, RouteRule{
		Port:        []int{53},
		ProcessName: processNames,
		Outbound:    "dns_out",
	}, RouteRule{
		ProcessName: processNames,
		Outbound:    "direct",
	})
}

func (t *Template) appendOutbound() {
	t.Outbounds = append(t.Outbounds, Outbound{
		Type: "direct",
		Tag:  "direct",
	}, Outbound{
		Type: "block",
		Tag:  "block",
	}, Outbound{
		Type: "dns",
		Tag:  "dns_out",
	})
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}

func WriteSingBoxConfig(content []byte) (err error) {
	err = os.WriteFile(GetSingBoxConfigPath(), content, os.FileMode(0600))
	if err != nil {
		return fmt.Errorf("WriteSingBoxConfig: %w", err)
	}
	return
}

func NewTunTemplate(setting *configure.Setting) (t *Template) {
	t = new(Template)
	t.DNS = new(DNS)
	t.Route = new(Route)
	var stack string
	switch setting.TransparentType {
	case configure.TransparentGvisorTun:
		stack = "gvisor"
	case configure.TransparentSystemTun:
		stack = "system"
	}
	t.Inbounds = append(t.Inbounds, Inbound{
		Type: "tun",
		Tag:  "tun-in",
		Format: &inbound.Tun{
			Inet4Address: "172.19.0.1/30",
			// Inet6Address: "fdfe:dcba:9876::1/126",
			Mtu:         9000,
			AutoRoute:   true,
			StrictRoute: false,
			Stack:       stack,
			Listen: net.Listen{
				Sniff: true,
			},
		},
	})
	t.Outbounds = append(t.Outbounds, Outbound{
		Type: "socks",
		Tag:  "proxy",
		Format: &outbound.SOCKS{
			Server:     "127.0.0.1",
			ServerPort: 52345,
			Dial: net.Dial{
				UdpFragment: true,
			},
		},
	})
	t.appendDNS()
	t.appendRoute()
	t.appendOutbound()
	return t
}
