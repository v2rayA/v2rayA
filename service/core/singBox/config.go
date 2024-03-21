package singBox

import (
	"github.com/v2rayA/v2rayA/core/singBox/inbound"
	"github.com/v2rayA/v2rayA/core/singBox/net"
	"github.com/v2rayA/v2rayA/core/singBox/outbound"
)

type Log struct {
	Disabled  bool   `json:"disabled"`
	Level     string `json:"level"`
	Output    string `json:"output"`
	Timestamp bool   `json:"timestamp"`
}

type DNSServer struct {
	Tag             string `json:"tag"`
	Address         string `json:"address"`
	AddressResolver string `json:"address_resolver,omitempty"`
	AddressStrategy string `json:"address_strategy,omitempty"`
	Strategy        string `json:"strategy,omitempty"`
	Detour          string `json:"detour,omitempty"`
}

type DNSRule struct {
	Inbound         []string `json:"inbound,omitempty"`
	IPVersion       string   `json:"ip_version,omitempty"`
	QueryType       []any    `json:"query_type,omitempty"`
	Network         string   `json:"network,omitempty"`
	AuthUser        []string `json:"auth_user,omitempty"`
	Protocol        []string `json:"protocol,omitempty"`
	Domain          []string `json:"domain,omitempty"`
	DomainSuffix    []string `json:"domain_suffix,omitempty"`
	DomainKeyword   []string `json:"domain_keyword,omitempty"`
	DomainRegex     []string `json:"domain_regex,omitempty"`
	GeoSite         []string `json:"geosite,omitempty"`
	SourceGeoIP     []string `json:"source_geoip,omitempty"`
	SourceIPCidr    []string `json:"source_ip_cidr,omitempty"`
	SourcePort      []int    `json:"source_port,omitempty"`
	SourcePortRange []string `json:"source_port_range,omitempty"`
	Port            []int    `json:"port,omitempty"`
	PortRange       []string `json:"port_range,omitempty"`
	ProcessName     []string `json:"process_name,omitempty"`
	ProcessPath     []string `json:"process_path,omitempty"`
	PackageName     []string `json:"package_name,omitempty"`
	Outbound        []string `json:"outbound,omitempty"`
	Server          string   `json:"server"`
	DisableCache    bool     `json:"disable_cache,omitempty"`
	RewriteTTL      int      `json:"rewrite_ttl,omitempty"`
}

type FakeIP struct {
	Enabled    bool   `json:"enabled"`
	Inet4Range string `json:"inet4_range"`
	Inet6Range string `json:"inet6_range"`
}

type DNS struct {
	Servers          []DNSServer `json:"servers"`
	Rules            []DNSRule   `json:"rules"`
	Final            string      `json:"final,omitempty"`
	Strategy         string      `json:"strategy"`
	DisableCache     bool        `json:"disable_cache,omitempty"`
	DisableExpire    bool        `json:"disable_expire,omitempty"`
	IndependentCache bool        `json:"independent_cache,omitempty"`
	ReverseMapping   bool        `json:"reverse_mapping,omitempty"`
	Fakeip           *FakeIP     `json:"fakeip,omitempty"`
}

type NTP struct {
	Enabled    bool   `json:"enabled"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port,omitempty"`
	Interval   string `json:"interval,omitempty"`
	net.Dial
}

type GeoIP struct {
	Path           string `json:"path,omitempty"`
	DownloadUrl    string `json:"download_url,omitempty"`
	DownloadDetour string `json:"download_detour,omitempty"`
}

type GeoSite GeoIP

type RouteRule struct {
	Inbound         []string `json:"inbound,omitempty"`
	IPVersion       string   `json:"ip_version,omitempty"`
	AuthUser        []string `json:"auth_user,omitempty"`
	Protocol        []string `json:"protocol,omitempty"`
	Network         []string `json:"network,omitempty"`
	Domain          []string `json:"domain,omitempty"`
	DomainSuffix    []string `json:"domain_suffix,omitempty"`
	DomainKeyword   []string `json:"domain_keyword,omitempty"`
	DomainRegex     []string `json:"domain_regex,omitempty"`
	GeoSite         []string `json:"geosite,omitempty"`
	SourceGeoIP     []string `json:"source_geoip,omitempty"`
	GeoIP           []string `json:"geoip,omitempty"`
	SourceIPCidr    []string `json:"source_ip_cidr,omitempty"`
	IPCidr          []string `json:"ip_cidr,omitempty"`
	SourcePort      []int    `json:"source_port,omitempty"`
	SourcePortRange []string `json:"source_port_range,omitempty"`
	Port            []int    `json:"port,omitempty"`
	PortRange       []string `json:"port_range,omitempty"`
	ProcessName     []string `json:"process_name,omitempty"`
	ProcessPath     []string `json:"process_path,omitempty"`
	PackageName     []string `json:"package_name,omitempty"`
	Outbound        string   `json:"outbound"`
}

type Route struct {
	GeoIP               *GeoIP      `json:"geoip,omitempty"`
	GeoSite             *GeoSite    `json:"geosite,omitempty"`
	Rules               []RouteRule `json:"rules"`
	Final               string      `json:"final,omitempty"`
	AutoDetectInterface bool        `json:"auto_detect_interface,omitempty"`
	OverrideAndroidVpn  bool        `json:"override_android_vpn,omitempty"`
	DefaultInterface    string      `json:"default_interface,omitempty"`
	DefaultMark         int         `json:"default_mark,omitempty"`
}

type Inbound struct {
	Type           string `json:"type"`
	Tag            string `json:"tag"`
	inbound.Format `json:",omitempty"`
}

type Outbound struct {
	Type            string `json:"type"`
	Tag             string `json:"tag"`
	outbound.Format `json:",omitempty"`
}
