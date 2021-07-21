package v2ray

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/routingA"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	dnsParser2 "github.com/v2rayA/v2rayA/infra/dnsParser"
	"log"
	"net"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

/*对应template.json*/
type TmplJson struct {
	Template       Template       `json:"template"`
	TCPSettings    TCPSettings    `json:"tcpSettings"`
	WsSettings     WsSettings     `json:"wsSettings"`
	TLSSettings    TLSSettings    `json:"tlsSettings"`
	KcpSettings    KcpSettings    `json:"kcpSettings"`
	HttpSettings   HttpSettings   `json:"httpSettings"`
	StreamSettings StreamSettings `json:"streamSettings"`
	Whitelist      []RoutingRule  `json:"whitelist"`
	Gfwlist        []RoutingRule  `json:"gfwlist"`
	Mux            Mux            `json:"mux"`
}
type Template struct {
	Log       *Log       `json:"log,omitempty"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   struct {
		DomainStrategy string        `json:"domainStrategy"`
		DomainMatcher  string        `json:"domainMatcher"`
		Rules          []RoutingRule `json:"rules"`
	} `json:"routing"`
	DNS     *DNS     `json:"dns,omitempty"`
	FakeDns *FakeDns `json:"fakedns,omitempty"`
}
type FakeDns struct {
	IpPool   string `json:"ipPool"`
	PoolSize int    `json:"poolSize"`
}
type RoutingRule struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Network     string   `json:"network,omitempty"`
	Port        string   `json:"port,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	Source      []string `json:"source,omitempty"`
	User        []string `json:"user,omitempty"`
}
type Log struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}
type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
	MetadataOnly bool     `json:"metadataOnly,omitempty"`
}
type Inbound struct {
	Port           int              `json:"port"`
	Protocol       string           `json:"protocol"`
	Listen         string           `json:"listen,omitempty"`
	Sniffing       Sniffing         `json:"sniffing,omitempty"`
	Settings       *InboundSettings `json:"settings,omitempty"`
	StreamSettings interface{}      `json:"streamSettings"`
	Tag            string           `json:"tag,omitempty"`
}
type InboundSettings struct {
	Auth           string      `json:"auth,omitempty"`
	UDP            bool        `json:"udp,omitempty"`
	IP             interface{} `json:"ip,omitempty"`
	Accounts       []Account   `json:"accounts,omitempty"`
	Clients        interface{} `json:"clients,omitempty"`
	Network        string      `json:"network,omitempty"`
	UserLevel      int         `json:"userLevel,omitempty"`
	Address        string      `json:"address,omitempty"`
	Port           int         `json:"port,omitempty"`
	FollowRedirect bool        `json:"followRedirect,omitempty"`
}
type Account struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}
type User struct {
	ID         string `json:"id"`
	AlterID    int    `json:"alterId,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Flow       string `json:"flow,omitempty"`
	Security   string `json:"security,omitempty"`
}
type Vnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}
type Server struct {
	Network  string         `json:"network,omitempty"`
	Address  string         `json:"address,omitempty"`
	Method   string         `json:"method,omitempty"`
	Ota      bool           `json:"ota,omitempty"`
	Password string         `json:"password,omitempty"`
	Port     int            `json:"port,omitempty"`
	Users    []OutboundUser `json:"users,omitempty"`
}
type Settings struct {
	Vnext          interface{} `json:"vnext,omitempty"`
	Servers        interface{} `json:"servers,omitempty"`
	DomainStrategy string      `json:"domainStrategy,omitempty"`
	Port           int         `json:"port,omitempty"`
	Address        string      `json:"address,omitempty"`
	Network        string      `json:"network,omitempty"`
	Redirect       string      `json:"redirect,omitempty"`
	UserLevel      *int        `json:"userLevel,omitempty"`
}
type TLSSettings struct {
	AllowInsecure        bool        `json:"allowInsecure"`
	ServerName           interface{} `json:"serverName,omitempty"`
	AllowInsecureCiphers bool        `json:"allowInsecureCiphers"`
}
type XTLSSettings struct {
	ServerName interface{} `json:"serverName,omitempty"`
}
type Headers struct {
	Host string `json:"Host"`
}
type WsSettings struct {
	ConnectionReuse bool    `json:"connectionReuse"`
	Path            string  `json:"path"`
	Headers         Headers `json:"headers"`
}
type StreamSettings struct {
	Network      string        `json:"network,omitempty"`
	Security     string        `json:"security,omitempty"`
	TLSSettings  *TLSSettings  `json:"tlsSettings,omitempty"`
	XTLSSettings *XTLSSettings `json:"xtlsSettings,omitempty"`
	TCPSettings  *TCPSettings  `json:"tcpSettings,omitempty"`
	KcpSettings  *KcpSettings  `json:"kcpSettings,omitempty"`
	WsSettings   *WsSettings   `json:"wsSettings,omitempty"`
	HTTPSettings *HttpSettings `json:"httpSettings,omitempty"`
	Sockopt      *Sockopt      `json:"sockopt,omitempty"`
}
type Sockopt struct {
	Mark        *int    `json:"mark,omitempty"`
	Tos         *int    `json:"tos,omitempty"`
	TCPFastOpen *bool   `json:"tcpFastOpen,omitempty"`
	Tproxy      *string `json:"tproxy,omitempty"`
}
type Mux struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency"`
}
type Outbound struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       *Settings       `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	ProxySettings  *ProxySettings  `json:"proxySettings,omitempty"`
	Mux            *Mux            `json:"mux,omitempty"`
}
type OutboundUser struct {
	User  string `json:"user"`
	Pass  string `json:"pass"`
	Level int    `json:"level,omitempty"`
}
type ProxySettings struct {
	Tag string `json:"tag,omitempty"`
}
type TCPSettings struct {
	ConnectionReuse bool `json:"connectionReuse"`
	Header          struct {
		Type    string `json:"type"`
		Request struct {
			Version string   `json:"version"`
			Method  string   `json:"method"`
			Path    []string `json:"path"`
			Headers struct {
				Host           []string `json:"Host"`
				UserAgent      []string `json:"User-Agent"`
				AcceptEncoding []string `json:"Accept-Encoding"`
				Connection     []string `json:"Connection"`
				Pragma         string   `json:"Pragma"`
			} `json:"headers"`
		} `json:"request"`
		Response struct {
			Version string `json:"version"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Headers struct {
				ContentType      []string `json:"Content-Type"`
				TransferEncoding []string `json:"Transfer-Encoding"`
				Connection       []string `json:"Connection"`
				Pragma           string   `json:"Pragma"`
			} `json:"headers"`
		} `json:"response"`
	} `json:"header"`
}
type KcpSettings struct {
	Mtu              int  `json:"mtu"`
	Tti              int  `json:"tti"`
	UplinkCapacity   int  `json:"uplinkCapacity"`
	DownlinkCapacity int  `json:"downlinkCapacity"`
	Congestion       bool `json:"congestion"`
	ReadBufferSize   int  `json:"readBufferSize"`
	WriteBufferSize  int  `json:"writeBufferSize"`
	Header           struct {
		Type     string      `json:"type"`
		Request  interface{} `json:"request"`
		Response interface{} `json:"response"`
	} `json:"header"`
	Seed string `json:"seed"`
}
type HttpSettings struct {
	Path string   `json:"path"`
	Host []string `json:"host"`
}
type Hosts map[string]interface{}

type DNS struct {
	Hosts           Hosts         `json:"hosts,omitempty"`
	Servers         []interface{} `json:"servers"`
	ClientIp        string        `json:"clientIp,omitempty"`
	Tag             string        `json:"tag,omitempty"`
	DisableFallback *bool         `json:"disableFallback,omitempty"`
	QueryStrategy   string        `json:"queryStrategy,omitempty"`
}
type DnsServer struct {
	Address      string   `json:"address"`
	Port         int      `json:"port,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	SkipFallback bool     `json:"skipFallback,omitempty"`
}
type Policy struct {
	Levels struct {
		Num0 struct {
			Handshake         int  `json:"handshake,omitempty"`
			ConnIdle          int  `json:"connIdle,omitempty"`
			UplinkOnly        int  `json:"uplinkOnly,omitempty"`
			DownlinkOnly      int  `json:"downlinkOnly,omitempty"`
			StatsUserUplink   bool `json:"statsUserUplink,omitempty"`
			StatsUserDownlink bool `json:"statsUserDownlink,omitempty"`
			BufferSize        int  `json:"bufferSize,omitempty"`
		} `json:"0"`
	} `json:"levels"`
	System struct {
		StatsInboundUplink   bool `json:"statsInboundUplink,omitempty"`
		StatsInboundDownlink bool `json:"statsInboundDownlink,omitempty"`
	} `json:"system"`
}

func NewTemplate() (tmpl Template) {
	return
}

/*
根据传入的 VmessInfo 填充模板
当协议是ss时，v.Net对应Method，v.ID对应Password
函数会规格化传入的v
*/

func ResolveOutbound(v *vmessInfo.VmessInfo, tag string, pluginPort *int) (o Outbound, err error) {
	socksPlugin := false
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return o, newError("error occurs while reading template json, please check whether templateJson variable is correct json format")
	}
	// 其中Template是基础配置，替换掉*t即可
	o = tmplJson.Template.Outbounds[0]
	// 默认协议vmess
	switch v.Protocol {
	case "":
		v.Protocol = "vmess"
	case "ss":
		v.Protocol = "shadowsocks"
	case "ssr":
		v.Protocol = "shadowsocksr"
	}
	// 根据vmessInfo修改json配置
	o.Protocol = v.Protocol
	port, _ := strconv.Atoi(v.Port)
	aid, _ := strconv.Atoi(v.Aid)
	switch strings.ToLower(v.Protocol) {
	case "vmess", "vless":
		switch strings.ToLower(v.Protocol) {
		case "vmess":
			o.Settings.Vnext = []Vnext{
				{
					Address: v.Add,
					Port:    port,
					Users: []User{
						{
							ID:       v.ID,
							AlterID:  aid,
							Security: "auto",
						},
					},
				},
			}
		case "vless":
			o.Settings.Vnext = []Vnext{
				{
					Address: v.Add,
					Port:    port,
					Users: []User{
						{
							ID: v.ID,
							//AlterID:    0, // keep AEAD on
							Encryption: "none",
						},
					},
				},
			}
		}
		o.StreamSettings = &tmplJson.StreamSettings
		o.StreamSettings.Network = v.Net
		// 根据传输协议(network)修改streamSettings
		//TODO: QUIC, gRPC
		switch strings.ToLower(v.Net) {
		case "ws":
			tmplJson.WsSettings.Headers.Host = v.Host
			tmplJson.WsSettings.Path = v.Path
			o.StreamSettings.WsSettings = &tmplJson.WsSettings
		case "mkcp", "kcp":
			tmplJson.KcpSettings.Header.Type = v.Type
			o.StreamSettings.KcpSettings = &tmplJson.KcpSettings
			o.StreamSettings.KcpSettings.Seed = v.Path
		case "tcp":
			if strings.ToLower(v.Type) == "http" {
				tmplJson.TCPSettings.Header.Request.Headers.Host = strings.Split(v.Host, ",")
				if v.Path != "" {
					tmplJson.TCPSettings.Header.Request.Path = strings.Split(v.Path, ",")
					for i := range tmplJson.TCPSettings.Header.Request.Path {
						if !strings.HasPrefix(tmplJson.TCPSettings.Header.Request.Path[i], "/") {
							tmplJson.TCPSettings.Header.Request.Path[i] = "/" + tmplJson.TCPSettings.Header.Request.Path[i]
						}
					}
				}
				o.StreamSettings.TCPSettings = &tmplJson.TCPSettings
			}
		case "h2", "http":
			tmplJson.HttpSettings.Host = strings.Split(v.Host, ",")
			tmplJson.HttpSettings.Path = v.Path
			o.StreamSettings.HTTPSettings = &tmplJson.HttpSettings
		}
		if strings.ToLower(v.TLS) == "tls" {
			o.StreamSettings.Security = "tls"
			o.StreamSettings.TLSSettings = &tmplJson.TLSSettings
			if v.AllowInsecure {
				o.StreamSettings.TLSSettings.AllowInsecure = true
			}
			ver, e := where.GetV2rayServiceVersion()
			if e != nil {
				log.Println(newError("cannot get the version of v2ray-core").Base(e))
			} else if !common.VersionMustGreaterEqual(ver, "4.23.2") {
				o.StreamSettings.TLSSettings.AllowInsecureCiphers = true
			}
			// SNI
			if v.Host != "" {
				o.StreamSettings.TLSSettings.ServerName = v.Host
			}
		} else if strings.ToLower(v.TLS) == "xtls" {
			o.StreamSettings.Security = "xtls"
			o.StreamSettings.XTLSSettings = new(XTLSSettings)
			// always set SNI
			if v.Host != "" {
				o.StreamSettings.XTLSSettings.ServerName = v.Host
			}
			if v.Flow == "" {
				v.Flow = "xtls-rprx-origin"
			}
			vnext := o.Settings.Vnext.([]Vnext)
			vnext[0].Users[0].Flow = v.Flow
			o.Settings.Vnext = vnext
		}
	case "shadowsocks":
		v.Net = strings.ToLower(v.Net)
		switch v.Net {
		case "aes-256-gcm", "aes-128-gcm", "chacha20-poly1305", "chacha20-ietf-poly1305", "plain", "none":
		default:
			return o, newError("unsupported shadowsocks encryption method: " + v.Net)
		}
		target := v.Add
		port := 0
		switch v.Type {
		case "http", "tls":
			target = "127.0.0.1"
			port = *pluginPort
		case "":
			port, _ = strconv.Atoi(v.Port)
		default:
			return o, newError("unsupported shadowsocks obfuscation method: " + v.TLS)
		}
		o.Settings.Servers = []Server{{
			Address:  target,
			Port:     port,
			Method:   v.Net,
			Password: v.ID,
		}}
	case "trojan":
		version, err := where.GetV2rayServiceVersion()
		if err != nil {
			return o, newError(err)
		}
		if ok, err := common.VersionGreaterEqual(version, "4.31.0"); err != nil || !ok {
			return o, newError("unsupported shadowsocks obfuscation method: " + v.TLS)
		}
		o.Settings.Servers = []Server{{
			Address:  v.Add,
			Port:     port,
			Password: v.ID,
		}}

		//tls
		o.StreamSettings = &tmplJson.StreamSettings
		o.StreamSettings.Network = "tcp"
		o.StreamSettings.Security = "tls"
		o.StreamSettings.TLSSettings = &tmplJson.TLSSettings
		if v.AllowInsecure {
			o.StreamSettings.TLSSettings.AllowInsecure = true
		}
		// always set SNI
		if v.Host != "" {
			o.StreamSettings.TLSSettings.ServerName = v.Host
		} else {
			o.StreamSettings.TLSSettings.ServerName = v.Add
		}
	case "shadowsocksr":
		v.Net = strings.ToLower(v.Net)
		switch v.Net {
		case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb", "none":
		default:
			return o, newError("unsupported shadowsocks encryption method: " + v.Net)
		}
		if len(strings.TrimSpace(v.Type)) <= 0 {
			v.Type = "origin"
		}
		switch v.Type {
		case "origin", "verify_sha1", "auth_sha1_v4", "auth_aes128_md5", "auth_aes128_sha1", "auth_chain_a", "auth_chain_b":
		default:
			return o, newError("unsupported shadowsocksR protocol: " + v.Type)
		}
		if len(strings.TrimSpace(v.TLS)) <= 0 {
			v.TLS = "plain"
		}
		switch v.TLS {
		case "plain", "http_simple", "http_post", "random_head", "tls1.2_ticket_auth":
		default:
			return o, newError("unsupported shadowsocksr obfuscation method: " + v.TLS)
		}
		socksPlugin = true
	case "pingtunnel":
		socksPlugin = true
	case "trojan-go":
		socksPlugin = true
	default:
		return o, newError("unsupported protocol: " + v.Protocol)
	}
	if socksPlugin && pluginPort != nil {
		o.Protocol = "socks"
		o.Settings.Servers = []Server{
			{
				Address: "127.0.0.1",
				Port:    *pluginPort,
			},
		}
	}
	o.Tag = tag
	return
}

type Addr struct {
	host string
	port string
}

func parseDnsAddr(addr string) Addr {
	// 223.5.5.5
	if net.ParseIP(addr) != nil {
		return Addr{
			host: addr,
			port: "53",
		}
	}
	// dns.google:53
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if _, err = strconv.Atoi(port); err == nil {
			return Addr{
				host: host,
				port: port,
			}
		}
	}
	// tcp://8.8.8.8:53, https://dns.google/dns-query
	if u, err := url.Parse(addr); err == nil {
		return Addr{
			host: u.Hostname(),
			port: u.Port(),
		}
	}
	// dns.google, dns.pub, etc.
	return Addr{
		host: addr,
		port: "53",
	}
}

type DnsRouting struct {
	DirectDomains []Addr
	ProxyDomains  []Addr
	DirectIPs     []Addr
	ProxyIPs      []Addr
}

func appendDnsServers(d *DNS, lines []string, domains []string) {
	for _, line := range lines {
		dns := dnsParser2.Parse(line)
		if u, err := url.Parse(dns.Val); err == nil && strings.Contains(dns.Val, "://") && !strings.Contains(u.Scheme, "://") {
			log.Println(u, u.Scheme, u.Host, u.Path)
			if domains != nil {
				d.Servers = append(d.Servers, DnsServer{
					Address: dns.Val,
					Domains: domains,
				})
			} else {
				d.Servers = append(d.Servers, dns.Val)
			}
		} else {
			addr := parseDnsAddr(dns.Val)
			p, _ := strconv.Atoi(addr.port)
			d.Servers = append(d.Servers, DnsServer{
				Address: addr.host,
				Port:    p,
				Domains: domains,
			})
		}
	}
}

func (t *Template) SetDNS(v vmessInfo.VmessInfo, setting *configure.Setting, supportUDP bool) (routing DnsRouting, err error) {
	// TODO: other countries and regions
	var internal, external []string
	var allThroughProxy = false
	if setting.AntiPollution == configure.AntipollutionAdvanced {
		// advanced
		internal = configure.GetInternalDnsListNotNil()
		external = configure.GetExternalDnsListNotNil()
		if len(external) == 0 {
			allThroughProxy = true
			for _, line := range internal {
				dns := dnsParser2.Parse(line)
				if dns.Out != "proxy" {
					allThroughProxy = false
					break
				}
			}
		}
		// check UDP support
		for _, line := range external {
			dns := dnsParser2.Parse(line)
			if dns.Out == "proxy" {
				if net.ParseIP(dns.Val) != nil {
					return DnsRouting{}, fmt.Errorf("sorry, your DNS setting may be invalid, because UDP is not supported for %v yet by v2rayA. Please use tcp:// or doh:// instead.", v.Protocol)
				}
				if _, port, err := net.SplitHostPort(dns.Val); err == nil {
					if _, err := strconv.Atoi(port); err == nil {
						return DnsRouting{}, fmt.Errorf("sorry, your DNS setting may be invalid, because UDP is not supported for %v yet by v2rayA. Please use tcp:// or doh:// instead.", v.Protocol)
					}
				}
			}
		}
	} else if setting.AntiPollution != configure.AntipollutionClosed {
		// preset
		internal = []string{"223.6.6.6 -> direct", "114.114.114.114 -> direct"}
		switch setting.AntiPollution {
		case configure.AntipollutionAntiHijack:
			break
		case configure.AntipollutionDnsForward:
			if supportUDP {
				external = []string{"8.8.8.8 -> proxy", "1.1.1.1 -> proxy"}
			} else {
				if err := CheckTcpDnsSupported(); err == nil {
					external = []string{"tcp://dns.opendns.com:5353 -> proxy", "tcp://dns.google -> proxy"}
				} else if err = CheckDohSupported(); err == nil {
					external = []string{"https://1.1.1.1/dns-query -> proxy", "https://dns.google/dns-query -> proxy"}
				} else {
					// compromise
					external = []string{"208.67.220.220:5353 -> direct", "208.67.222.222 -> direct"}
				}
			}
		case configure.AntipollutionDoH:
			external = []string{"https://doh.pub/dns-query -> direct", "https://rubyfish.cn/dns-query -> direct"}
		}
	}
	True := true
	t.DNS = &DNS{
		Tag: "dns",
	}
	if allThroughProxy {
		// guess the user want to protect the privacy
		t.DNS.DisableFallback = &True
	}
	if setting.AntiPollution != configure.AntipollutionClosed {
		if len(external) == 0 {
			// not split traffic
			appendDnsServers(t.DNS, internal, nil)
		} else {
			// split traffic
			appendDnsServers(t.DNS, external, nil)
			appendDnsServers(t.DNS, internal, []string{"geosite:cn"})
		}
	}
	// routing
	dnsList := append(append([]string{}, internal...), external...)
	for _, line := range dnsList {
		dns := dnsParser2.Parse(line)
		if dns.Val == "localhost" {
			// no need to routing
			continue
		}
		// we believe all lines are legal
		var addr = parseDnsAddr(dns.Val)

		if dns.Out == "direct" {
			if net.ParseIP(addr.host) == nil {
				routing.DirectDomains = append(routing.DirectDomains, addr)
			} else {
				routing.DirectIPs = append(routing.DirectIPs, addr)
			}
		} else {
			if net.ParseIP(addr.host) == nil {
				routing.ProxyDomains = append(routing.ProxyDomains, addr)
			} else {
				routing.ProxyIPs = append(routing.ProxyIPs, addr)
			}
		}
	}

	// fakedns
	if t.FakeDns != nil {
		ds := DnsServer{
			Address: "fakedns",
			Domains: []string{
				"domain:use-fakedns.com",
			},
		}
		if asset.LoyalsoldierSiteDatExists() {
			// use more accurate list to avoid misadventure
			ds.Domains = append(ds.Domains, "ext:LoyalsoldierSite.dat:gfw")
		} else {
			ds.Domains = append(ds.Domains, "geosite:geolocation-!cn")
		}
		if len(t.DNS.Servers) == 0 {
			log.Println("[Fakedns]: NOT REASONABLE. Please report your config.")
			t.DNS.Servers = append(t.DNS.Servers, "localhost")
		}
		t.DNS.Servers = append(t.DNS.Servers, ds)
	}

	if t.DNS.Servers == nil {
		t.DNS.Servers = []interface{}{"localhost"}
	}
	var domainsToLookup []string
	if net.ParseIP(v.Add) == nil {
		domainsToLookup = append(domainsToLookup, v.Add)
	}
	for _, addr := range routing.ProxyDomains {
		domainsToLookup = append(domainsToLookup, addr.host)
	}
	for _, addr := range routing.DirectDomains {
		domainsToLookup = append(domainsToLookup, addr.host)
	}
	var domainsToHosts []string
	if len(domainsToLookup) > 0 {
		if CheckDohSupported() == nil {
			t.DNS.Servers = append(t.DNS.Servers, DnsServer{
				Address:      "https://doh.pub/dns-query",
				Domains:      domainsToLookup,
				SkipFallback: true,
			})
			t.DNS.Servers = append(t.DNS.Servers, DnsServer{
				Address:      "https://doh.alidns.com/dns-query",
				Domains:      domainsToLookup,
				SkipFallback: true,
			})
			domainsToHosts = append(domainsToHosts, "doh.pub")
			domainsToHosts = append(domainsToHosts, "doh.alidns.com")
		} else {
			t.DNS.Servers = append(t.DNS.Servers, DnsServer{
				Address:      "dns.pub",
				Domains:      domainsToLookup,
				SkipFallback: true,
			})
			t.DNS.Servers = append(t.DNS.Servers, DnsServer{
				Address:      "dns.alidns.com",
				Domains:      domainsToLookup,
				SkipFallback: true,
			})
			domainsToHosts = append(domainsToHosts, "dns.pub")
			domainsToHosts = append(domainsToHosts, "dns.alidns.com")
		}
	}
	// set hosts
	for _, domain := range domainsToHosts {
		ips, err := net.LookupHost(domain)
		if err != nil {
			return routing, fmt.Errorf("[Error] %w: please make sure you're connected to the Internet", err)
		}
		if t.DNS.Hosts == nil {
			t.DNS.Hosts = make(Hosts)
		}
		ips = filterIPs(ips)
		if CheckHostsListSupported() == nil {
			t.DNS.Hosts[domain] = ips
		} else {
			t.DNS.Hosts[domain] = ips[0]
		}
	}
	return
}
func filterIPs(ips []string) []string {
	if iptables.IsIPv6Supported() {
		return ips
	}
	var ret []string
	for _, ip := range ips {
		if net.ParseIP(ip).To4() != nil {
			ret = append(ret, ip)
		}
	}
	return ret
}
func (t *Template) SetDNSRouting(routing DnsRouting, supportUDP bool) {
	for _, r := range routing.DirectIPs {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{Type: "field", InboundTag: []string{"dns"}, OutboundTag: "direct", IP: []string{r.host}, Port: r.port},
		)
	}
	for _, r := range routing.ProxyIPs {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{Type: "field", InboundTag: []string{"dns"}, OutboundTag: "proxy", IP: []string{r.host}, Port: r.port},
		)
	}
	for _, r := range routing.DirectDomains {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{Type: "field", InboundTag: []string{"dns"}, OutboundTag: "direct", Domain: []string{r.host}, Port: r.port},
		)
	}
	for _, r := range routing.ProxyDomains {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{Type: "field", InboundTag: []string{"dns"}, OutboundTag: "proxy", Domain: []string{r.host}, Port: r.port},
		)
	}
	t.Routing.Rules = append(t.Routing.Rules,
		RoutingRule{Type: "field", InboundTag: []string{"dns"}, OutboundTag: "direct"},
	)
	setting := configure.GetSettingNotNil()
	if setting.AntiPollution != configure.AntipollutionClosed {
		dnsOut := RoutingRule{ // hijack traffic to port 53
			Type:        "field",
			Port:        "53",
			OutboundTag: "dns-out",
		}
		if specialMode.ShouldLocalDnsListen() && specialMode.CouldLocalDnsListen() == nil {
			dnsOut.InboundTag = []string{"dns-in"}
		}
		if specialMode.ShouldUseSupervisor() {
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{ // supervisor
					Type:        "field",
					IP:          []string{"240.0.0.0/4"},
					OutboundTag: "proxy",
				},
			)
		}
		t.Routing.Rules = append(t.Routing.Rules, dnsOut)
	}
	if !supportUDP {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				Network:     "udp",
			},
		)
	}
	return
}

func (t *Template) SetRulePortRouting(setting *configure.Setting) error {
	switch setting.RulePortMode {
	case configure.WhitelistMode:
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{ // 直连中国大陆主流网站域名
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"rule"},
				Domain:      []string{"geosite:cn"},
			},
			RoutingRule{ // 直连中国大陆主流网站 ip 和 私有 ip
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"rule"},
				IP:          []string{"geoip:private", "geoip:cn"},
			},
		)
	case configure.GfwlistMode:
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  []string{"rule"},
				Domain:      []string{"ext:LoyalsoldierSite.dat:geolocation-!cn"},
			},
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"rule"},
			},
		)
	case configure.CustomMode:
		customPac := configure.GetCustomPacNotNil()
		var lastOutboundTag configure.PacRuleType
		var lastMatchType configure.PacMatchType
		for _, v := range customPac.RoutingRules {
			reuse := false
			var rule *RoutingRule
			//如果相邻规则的outbound类型以及matchType相同，则合并
			if v.RuleType == lastOutboundTag && v.MatchType == lastMatchType {
				rule = &t.Routing.Rules[len(t.Routing.Rules)-1]
				reuse = true
			} else {
				rule = &RoutingRule{
					Type:        "field",
					OutboundTag: string(v.RuleType),
					InboundTag:  []string{"rule"},
				}
			}
			for i := range v.Tags {
				r := fmt.Sprintf("ext:%v:%v", v.Filename, v.Tags[i])
				switch v.MatchType {
				case configure.DomainMatchRule:
					rule.Domain = append(rule.Domain, r)
				case configure.IpMatchRule:
					rule.IP = append(rule.IP, r)
				}
			}
			if !reuse {
				t.Routing.Rules = append(t.Routing.Rules, *rule)
			}
			lastOutboundTag = v.RuleType
			lastMatchType = v.MatchType
		}
		switch customPac.DefaultProxyMode {
		case configure.DefaultProxyMode:
		case configure.DefaultDirectMode:
			//如果默认直连，则需要加上下述规则
			t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"rule"},
			})
		case configure.DefaultBlockMode:
			//如果默认拦截，则需要加上下述规则
			t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
				Type:        "field",
				OutboundTag: "block",
				InboundTag:  []string{"rule"},
			})
		}
	case configure.RoutingAMode:
		if err := parseRoutingA(t, []string{"rule"}); err != nil {
			return err
		}
	}
	return nil
}
func parseRoutingA(t *Template, routingInboundTags []string) error {
	ra := configure.GetRoutingA()
	rules, err := routingA.Parse(ra)
	if err != nil {
		log.Println(err)
		return err
	}
	defaultOutbound := "proxy"
	for _, rule := range rules {
		switch rule := rule.(type) {
		case routingA.Define:
			switch rule.Name {
			case "default":
				switch v := rule.Value.(type) {
				case string:
					defaultOutbound = v
				}
			case "inbound", "outbound":
				switch o := rule.Value.(type) {
				case routingA.Bound:
					proto := o.Value
					switch proto.Name {
					case "http", "socks":
						if len(proto.NamedParams["address"]) < 1 ||
							len(proto.NamedParams["port"]) < 1 {
							continue
						}
						port, err := strconv.Atoi(proto.NamedParams["port"][0])
						if err != nil {
							continue
						}
						server := Server{
							Port:    port,
							Address: proto.NamedParams["address"][0],
						}
						if unames := proto.NamedParams["user"]; len(unames) > 0 {
							passwords := proto.NamedParams["pass"]
							levels := proto.NamedParams["level"]
							for i, uname := range unames {
								u := OutboundUser{
									User: uname,
								}
								if i < len(passwords) {
									u.Pass = passwords[i]
								}
								if i < len(levels) {
									level, err := strconv.Atoi(levels[i])
									if err == nil {
										u.Level = level
									}
								}
								server.Users = append(server.Users, u)
							}
						}
						switch rule.Name {
						case "outbound":
							t.Outbounds = append(t.Outbounds, Outbound{
								Tag:      o.Name,
								Protocol: o.Value.Name,
								Settings: &Settings{
									Servers: []Server{
										server,
									},
								},
							})
						case "inbound":
							// reform from outbound
							in := Inbound{
								Tag:      o.Name,
								Protocol: o.Value.Name,
								Listen:   server.Address,
								Port:     server.Port,
								Settings: &InboundSettings{
									UDP: false,
								},
								Sniffing: Sniffing{
									Enabled:      true,
									DestOverride: []string{"http", "tls"},
								},
							}
							if proto.Name == "socks" {
								if len(server.Users) > 0 {
									in.Settings.Auth = "password"
								}
								if udp := proto.NamedParams["udp"]; len(udp) > 0 {
									in.Settings.UDP = udp[0] == "true"
								}
								if userLevels := proto.NamedParams["userLevel"]; len(userLevels) > 0 {
									userLevel, err := strconv.Atoi(userLevels[0])
									if err == nil {
										in.Settings.UserLevel = userLevel
									}
								}
							}
							if len(server.Users) > 0 {
								for _, u := range server.Users {
									in.Settings.Accounts = append(in.Settings.Accounts, Account{
										User: u.User,
										Pass: u.Pass,
									})
								}
							}
							t.Inbounds = append(t.Inbounds, in)
							routingInboundTags = append(routingInboundTags, o.Name)
						}
					case "freedom":
						settings := new(Settings)
						if len(proto.NamedParams["domainStrategy"]) > 0 {
							settings.DomainStrategy = proto.NamedParams["domainStrategy"][0]
						}
						if len(proto.NamedParams["redirect"]) > 0 {
							settings.Redirect = proto.NamedParams["redirect"][0]
						}
						if len(proto.NamedParams["userLevel"]) > 0 {
							level, err := strconv.Atoi(proto.NamedParams["userLevel"][0])
							if err == nil {
								settings.UserLevel = &level
							}
						}
						t.Outbounds = append(t.Outbounds, Outbound{
							Tag:      o.Name,
							Protocol: o.Value.Name,
							Settings: settings,
						})
					}
				}
			}
		case routingA.Routing:
			rr := RoutingRule{
				Type:        "field",
				OutboundTag: rule.Out,
				InboundTag:  routingInboundTags,
			}
			for _, f := range rule.And {
				switch f.Name {
				case "domain":
					for k, vv := range f.NamedParams {
						for _, v := range vv {
							if k == "contains" {
								rr.Domain = append(rr.Domain, v)
								continue
							}
							rr.Domain = append(rr.Domain, fmt.Sprintf("%v:%v", k, v))
						}
					}
					//this is not recommended
					rr.Domain = append(rr.Domain, f.Params...)
				case "ip":
					for k, vv := range f.NamedParams {
						for _, v := range vv {
							rr.IP = append(rr.IP, fmt.Sprintf("%v:%v", k, v))
						}
					}
					rr.IP = append(rr.IP, f.Params...)
				case "network":
					rr.Network = strings.Join(f.Params, ",")
				case "port":
					rr.Port = strings.Join(f.Params, ",")
				case "protocol":
					rr.Protocol = f.Params
				case "source":
					rr.Source = f.Params
				case "user":
					rr.User = f.Params
				case "inboundTag":
					rr.InboundTag = f.Params
				}
			}
			t.Routing.Rules = append(t.Routing.Rules, rr)
		}
	}
	t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
		Type:        "field",
		OutboundTag: defaultOutbound,
		InboundTag:  []string{"rule"},
	})
	return nil
}
func (t *Template) SetTransparentRouting(setting *configure.Setting) {
	switch setting.Transparent {
	case configure.TransparentProxy:
	case configure.TransparentWhitelist:
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{ // 直连中国大陆主流网站域名
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"transparent"},
				Domain:      []string{"geosite:cn"},
			},
			RoutingRule{ // 直连中国大陆主流网站 ip 和 私有 ip
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"transparent"},
				IP:          []string{"geoip:private", "geoip:cn"},
			},
		)
	case configure.TransparentGfwlist:
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  []string{"transparent"},
				Domain:      []string{"ext:LoyalsoldierSite.dat:geolocation-!cn"},
			},
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"transparent"},
			},
		)
	case configure.TransparentPac:
		//transparent模式跟随pac
		for i := range t.Routing.Rules {
			bIncludePac := false
			for _, in := range t.Routing.Rules[i].InboundTag {
				if in == "rule" {
					bIncludePac = true
					break
				}
			}
			if bIncludePac {
				t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, "transparent")
			}
		}
	}
}
func (t *Template) AppendDokodemo(tproxy *string, port int, tag string) {
	dokodemo := Inbound{
		Listen:   "0.0.0.0",
		Port:     port,
		Protocol: "dokodemo-door",
		Sniffing: Sniffing{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
		Settings: &InboundSettings{Network: "tcp,udp"},
		Tag:      tag,
	}
	if tproxy != nil {
		dokodemo.StreamSettings = StreamSettings{Sockopt: &Sockopt{Tproxy: tproxy}}
		dokodemo.Settings.FollowRedirect = true

	}
	t.Inbounds = append(t.Inbounds, dokodemo)
}

func (t *Template) SetOutboundSockopt(supportUDP bool, setting *configure.Setting) {
	mark := 0xff
	//tos := 184
	for i := range t.Outbounds {
		if t.Outbounds[i].Protocol == "blackhole" {
			continue
		}
		if t.Outbounds[i].StreamSettings == nil {
			t.Outbounds[i].StreamSettings = new(StreamSettings)
		}
		if t.Outbounds[i].StreamSettings.Sockopt == nil {
			t.Outbounds[i].StreamSettings.Sockopt = new(Sockopt)
		}
		if t.Outbounds[i].Protocol == "freedom" && t.Outbounds[i].Tag == "direct" {
			t.Outbounds[i].Settings.DomainStrategy = "UseIP"
		}
		if setting.TcpFastOpen != configure.Default {
			tmp := setting.TcpFastOpen == configure.Yes
			t.Outbounds[i].StreamSettings.Sockopt.TCPFastOpen = &tmp
		}
		checkAndSetMark(&t.Outbounds[i], mark)
	}
}
func (t *Template) SetInboundListenAddress(setting *configure.Setting) {
	const (
		tag4Suffix = "_ipv4"
		tag6Suffix = "_ipv6"
	)
	tagMap := make(map[string]struct{})
	inbounds6 := make([]Inbound, len(t.Inbounds))
	copy(inbounds6, t.Inbounds)
	if !setting.IntranetSharing {
		// copy a group of ipv6 inbounds and set the tag
		for i := range t.Inbounds {
			t.Inbounds[i].Listen = "127.0.0.1"
			inbounds6[i].Listen = "::1"
			if t.Inbounds[i].Tag != "" {
				tagMap[t.Inbounds[i].Tag] = struct{}{}
				t.Inbounds[i].Tag += tag4Suffix
				inbounds6[i].Tag += tag6Suffix
			}
		}
		t.Inbounds = append(t.Inbounds, inbounds6...)

		// set routing
		for i := range t.Routing.Rules {
			tag6 := make([]string, 0)
			for j, tag := range t.Routing.Rules[i].InboundTag {
				if _, ok := tagMap[tag]; ok {
					t.Routing.Rules[i].InboundTag[j] += tag4Suffix
					tag6 = append(tag6, tag+tag6Suffix)
				}
			}
			if len(tag6) > 0 {
				t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, tag6...)
			}
		}
	}
}
func (t *Template) SetInboundFakeDnsDestOverride() {
	if t.FakeDns == nil {
		return
	}
	for i := range t.Inbounds {
		if t.Inbounds[i].Sniffing.Enabled == false {
			continue
		}
		t.Inbounds[i].Sniffing.DestOverride = append(t.Inbounds[i].Sniffing.DestOverride, "fakedns")
		//t.Inbounds[i].Sniffing.DestOverride = []string{"fakedns"}
	}
}

func (t *Template) AppendDNSOutbound() {
	t.Outbounds = append(t.Outbounds, Outbound{
		Tag:      "dns-out",
		Protocol: "dns",
	})
}

func (t *Template) SetInbound(setting *configure.Setting) {
	ports := configure.GetPorts()
	if ports != nil {
		t.Inbounds[2].Port = ports.HttpWithPac
		t.Inbounds[1].Port = ports.Http
		t.Inbounds[0].Port = ports.Socks5
		//端口为0则删除
		for i := 2; i >= 0; i-- {
			if t.Inbounds[i].Port == 0 {
				t.Inbounds = append(t.Inbounds[:i], t.Inbounds[i+1:]...)
			}
		}
	}
	if setting.Transparent != configure.TransparentClose {
		var tproxy string
		switch setting.TransparentType {
		case configure.TransparentTproxy, configure.TransparentRedirect:
			tproxy = string(setting.TransparentType)
		}
		t.AppendDokodemo(&tproxy, 32345, "transparent")
	}
	if specialMode.ShouldLocalDnsListen() && specialMode.CouldLocalDnsListen() == nil {
		// FIXME: xray cannot use fakedns+others (2021-07-17), set up a solo dokodemo-door for fakedns
		t.Inbounds = append(t.Inbounds, Inbound{
			Port:     53,
			Protocol: "dokodemo-door",
			Listen:   "0.0.0.0",
			Settings: &InboundSettings{
				Network: "udp",
				Address: "2.0.1.7",
				Port:    53,
			},
			Tag: "dns-in",
		})
	}
}

func NewTemplateFromVmessInfo(v vmessInfo.VmessInfo) (t Template, err error) {
	setting := configure.GetSettingNotNil()
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return t, newError("error occurs while reading template json, please check whether templateJson variable is correct json format")
	}
	// 其中Template是基础配置，替换掉t即可
	t = tmplJson.Template
	// 调试模式
	if global.GetEnvironmentConfig().Verbose {
		t.Log.Loglevel = "info"
		t.Log.Access = ""
		t.Log.Error = ""
	} else if global.IsDebug() {
		os.WriteFile(t.Log.Access, nil, 0777)
		os.WriteFile(t.Log.Error, nil, 0777)
		os.Chmod(t.Log.Access, 0777)
		os.Chmod(t.Log.Error, 0777)
		t.Log.Loglevel = "debug"
	} else {
		t.Log = nil
	}
	// fakedns
	if specialMode.ShouldUseFakeDns() {
		t.FakeDns = &FakeDns{
			IpPool:   "198.18.0.0/15",
			PoolSize: 65535,
		}
	}
	// 解析Outbound
	o, err := ResolveOutbound(&v, "proxy", &global.GetEnvironmentConfig().PluginListenPort)
	if err != nil {
		return t, err
	}
	t.Outbounds[0] = o
	var supportUDP = true
	switch o.Protocol {
	case "vmess", "vless":
		//是否在设置了里开启了mux
		muxon := setting.MuxOn == configure.Yes
		if v.TLS == "xtls" {
			//xtls与mux不共存
			muxon = false
		}
		t.Outbounds[0].Mux = &Mux{
			Enabled:     muxon,
			Concurrency: setting.Mux,
		}
	case "ss", "shadowsocks", "trojan":
		break
	default:
		supportUDP = false
	}
	//根据配置修改端口
	t.SetInbound(setting)
	//设置DNS
	dnsRouting, err := t.SetDNS(v, setting, supportUDP)
	if err != nil {
		return
	}
	//再修改outbounds
	t.AppendDNSOutbound()
	//最后是routing
	t.Routing.DomainMatcher = "mph"
	t.SetDNSRouting(dnsRouting, supportUDP)
	//规则端口规则
	if err = t.SetRulePortRouting(setting); err != nil {
		return
	}
	//根据是否使用全局代理修改路由
	if setting.Transparent != configure.TransparentClose {
		t.SetTransparentRouting(setting)
	}
	//置outboundSockopt
	t.SetOutboundSockopt(supportUDP, setting)

	//置fakedns destOverride
	t.SetInboundFakeDnsDestOverride()

	//置inbound listen address

	t.SetInboundListenAddress(setting)

	//check dulplicated tags
	if err = t.CheckDuplicatedTags(); err != nil {
		return
	}

	return t, nil
}

func (t *Template) CheckDuplicatedTags() error {
	inboundTagsSet := make(map[string]interface{})
	for _, in := range t.Inbounds {
		tag := in.Tag
		if _, exists := inboundTagsSet[tag]; exists {
			return newError("duplicated inbound tag: ", tag).AtError()
		} else {
			inboundTagsSet[tag] = nil
		}
	}
	outboundTagsSet := make(map[string]interface{})
	for _, out := range t.Outbounds {
		tag := out.Tag
		if _, exists := outboundTagsSet[tag]; exists {
			return newError("duplicated outbound tag: ", tag)
		} else {
			outboundTagsSet[tag] = nil
		}
	}
	return nil
}

func (t *Template) CheckInboundPortsOccupied() (occupied bool, port string, pname string) {
	var st []string
	for _, in := range t.Inbounds {
		switch strings.ToLower(in.Protocol) {
		case "http", "vmess", "vless", "trojan":
			st = append(st, strconv.Itoa(in.Port)+":tcp")
		case "dokodemo-door":
			if in.Settings != nil && in.Settings.Network != "" {
				st = append(st, strconv.Itoa(in.Port)+":"+in.Settings.Network)
			} else {
				st = append(st, strconv.Itoa(in.Port)+":tcp,udp")
			}
		default:
			st = append(st, strconv.Itoa(in.Port)+":tcp,udp")
		}
	}
	v2rayPath, _ := where.GetV2rayBinPath()
	occupied, socket, err := ports.IsPortOccupiedWithWhitelist(st, map[string]struct{}{path.Base(v2rayPath): {}})
	if err != nil {
		return true, "unknown", err.Error()
	}
	if occupied {
		port = strconv.Itoa(socket.LocalAddress.Port)
		process, _ := socket.Process()
		pname = process.Name
	}
	return
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}

func WriteV2rayConfig(content []byte) (err error) {
	err = os.WriteFile(asset.GetV2rayConfigPath(), content, os.FileMode(0600))
	if err != nil {
		return newError("WriteV2rayConfig").Base(err)
	}
	return
}

func NewTemplateFromConfig() (t Template, err error) {
	b, err := asset.GetConfigBytes()
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(b, &t)
	return
}

func checkAndSetMark(o *Outbound, mark int) {
	if configure.GetSettingNotNil().Transparent == configure.TransparentClose {
		return
	}
	if o.StreamSettings == nil {
		o.StreamSettings = new(StreamSettings)
	}
	if o.StreamSettings.Sockopt == nil {
		o.StreamSettings.Sockopt = new(Sockopt)
	}
	o.StreamSettings.Sockopt.Mark = &mark
}

func (t *Template) AddMappingOutbound(v vmessInfo.VmessInfo, inboundPort string, udpSupport bool, pluginPort int, protocol string) (err error) {
	o, err := ResolveOutbound(&v, "outbound"+inboundPort, &pluginPort)
	if err != nil {
		return
	}
	var mark = 0xff
	checkAndSetMark(&o, mark)
	t.Outbounds = append(t.Outbounds, o)
	iPort, err := strconv.Atoi(inboundPort)
	if err != nil || iPort <= 0 {
		return newError("port of inbound must be a positive number with string type")
	}
	if protocol == "" {
		protocol = "socks"
	}
	t.Inbounds = append(t.Inbounds, Inbound{
		Port:     iPort,
		Protocol: protocol,
		Listen:   "0.0.0.0",
		Sniffing: Sniffing{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
		Settings: &InboundSettings{
			Auth: "noauth",
			UDP:  udpSupport,
		},
		Tag: "inbound" + inboundPort,
	})
	if t.Routing.DomainStrategy == "" {
		t.Routing.DomainStrategy = "IPOnDemand"
	}
	//插入最前
	tmp := make([]RoutingRule, 1, len(t.Routing.Rules)+1)
	tmp[0] = RoutingRule{
		Type:        "field",
		OutboundTag: "outbound" + inboundPort,
		InboundTag:  []string{"inbound" + inboundPort},
	}
	t.Routing.Rules = append(tmp, t.Routing.Rules...)
	return
}

func getHosts() (h Hosts) {
	h = make(Hosts)
	b, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return
	}
	regex := regexp.MustCompile(`\s+`)
	lines := bytes.Split(b, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}
		s := string(regex.ReplaceAll(line, []byte(" ")))
		arr := strings.Split(s, " ")
		lenArr := len(arr)
		if lenArr > 1 {
			for i := 1; i < lenArr; i++ {
				h[arr[i]] = arr[0]
			}
		}
	}
	return
}
