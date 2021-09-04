package v2ray

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/routingA"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	dnsParser2 "github.com/v2rayA/v2rayA/infra/dnsParser"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*对应template.json*/
type TmplJson struct {
	Template           Template       `json:"template"`
	TmplTCPSettings    TCPSettings    `json:"tcpSettings"`
	TmplWsSettings     WsSettings     `json:"wsSettings"`
	TmplTLSSettings    TLSSettings    `json:"tlsSettings"`
	TmplKcpSettings    KcpSettings    `json:"kcpSettings"`
	TmplHttpSettings   HttpSettings   `json:"httpSettings"`
	TmplStreamSettings StreamSettings `json:"streamSettings"`
}
type Template struct {
	Log       *Log             `json:"log,omitempty"`
	Inbounds  []Inbound        `json:"inbounds"`
	Outbounds []OutboundObject `json:"outbounds"`
	Routing   struct {
		DomainStrategy string        `json:"domainStrategy"`
		DomainMatcher  string        `json:"domainMatcher,omitempty"`
		Rules          []RoutingRule `json:"rules"`
		Balancers      []Balancer    `json:"balancers,omitempty"`
	} `json:"routing"`
	DNS         *DNS         `json:"dns,omitempty"`
	FakeDns     *FakeDns     `json:"fakedns,omitempty"`
	Observatory *Observatory `json:"observatory,omitempty"`
	API         *APIObject   `json:"api,omitempty"`
}
type APIObject struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}
type Observatory struct {
	SubjectSelector []string `json:"subjectSelector"`
	ProbeURL        string   `json:"probeURL,omitempty"`
	ProbeInterval   string   `json:"ProbeInterval,omitempty"`
}
type Balancer struct {
	Tag      string           `json:"tag"`
	Selector []string         `json:"selector"`
	Strategy BalancerStrategy `json:"strategy"`
}
type BalancerStrategy struct {
	Type string `json:"type"`
}
type FakeDns struct {
	IpPool   string `json:"ipPool"`
	PoolSize int    `json:"poolSize"`
}
type RoutingRule struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Network     string   `json:"network,omitempty"`
	Port        string   `json:"port,omitempty"`
	SourcePort  string   `json:"sourcePort,omitempty"`
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
	DestOverride []string `json:"destOverride,omitempty"`
	MetadataOnly bool     `json:"metadataOnly"`
}
type Inbound struct {
	Port           int              `json:"port"`
	Protocol       string           `json:"protocol"`
	Listen         string           `json:"listen,omitempty"`
	Sniffing       Sniffing         `json:"sniffing,omitempty"`
	Settings       *InboundSettings `json:"settings,omitempty"`
	StreamSettings *StreamSettings  `json:"streamSettings"`
	Tag            string           `json:"tag,omitempty"`
}
type InboundSettings struct {
	Auth           string      `json:"auth,omitempty"`
	UDP            bool        `json:"udp,omitempty"`
	IP             interface{} `json:"ip,omitempty"`
	Accounts       []Account   `json:"accounts,omitempty"`
	Clients        interface{} `json:"clients,omitempty"`
	Decryption     string      `json:"decryption,omitempty"`
	Network        string      `json:"network,omitempty"`
	UserLevel      int         `json:"userLevel,omitempty"`
	Address        string      `json:"address,omitempty"`
	Port           int         `json:"port,omitempty"`
	FollowRedirect bool        `json:"followRedirect,omitempty"`
}
type VlessClient struct {
	Id    string `json:"id"`
	Level int    `json:"level,omitempty"`
	Email string `json:"email,omitempty"`
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
	AllowInsecure                    bool          `json:"allowInsecure"`
	ServerName                       interface{}   `json:"serverName,omitempty"`
	AllowInsecureCiphers             bool          `json:"allowInsecureCiphers"`
	Alpn                             []string      `json:"alpn,omitempty"`
	PinnedPeerCertificateChainSha256 string        `json:"pinnedPeerCertificateChainSha256,omitempty"`
	Certificates                     []Certificate `json:"certificates,omitempty"`
}
type Certificate struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
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
	XTLSSettings *TLSSettings  `json:"xtlsSettings,omitempty"`
	TCPSettings  *TCPSettings  `json:"tcpSettings,omitempty"`
	KcpSettings  *KcpSettings  `json:"kcpSettings,omitempty"`
	WsSettings   *WsSettings   `json:"wsSettings,omitempty"`
	HTTPSettings *HttpSettings `json:"httpSettings,omitempty"`
	GrpcSettings *GrpcSettings `json:"grpcSettings,omitempty"`
	Sockopt      *Sockopt      `json:"sockopt,omitempty"`
}
type GrpcSettings struct {
	ServiceName string `json:"serviceName"`
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
type OutboundObject struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       Settings        `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	ProxySettings  *ProxySettings  `json:"proxySettings,omitempty"`
	Mux            *Mux            `json:"mux,omitempty"`
	balancers      []string
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
	Path   string   `json:"path"`
	Host   []string `json:"host"`
	Method string   `json:"method,omitempty"`
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

/*
根据传入的 VmessInfo 填充模板
*/
func ResolveOutbound(v *vmessInfo.VmessInfo, tag string, pluginPort *int) (o OutboundObject, err error) {
	setting := configure.GetSettingNotNil()
	socksPlugin := false
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return o, fmt.Errorf("error occurs while reading template json, please check whether templateJson variable is correct json format")
	}
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
	o = OutboundObject{
		Tag:      tag,
		Protocol: v.Protocol,
	}
	port, _ := strconv.Atoi(v.Port)
	aid, _ := strconv.Atoi(v.Aid)
	switch strings.ToLower(v.Protocol) {
	case "vmess", "vless":
		id := v.ID
		if l := len([]byte(id)); l < 32 || l > 36 {
			id = common.StringToUUID5(id)
		}
		switch strings.ToLower(v.Protocol) {
		case "vmess":
			o.Settings.Vnext = []Vnext{
				{
					Address: v.Add,
					Port:    port,
					Users: []User{
						{
							ID:       id,
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
							ID: id,
							//AlterID:    0, // keep AEAD on
							Encryption: "none",
						},
					},
				},
			}
		}
		streamSetting := tmplJson.TmplStreamSettings
		o.StreamSettings = &streamSetting
		o.StreamSettings.Network = v.Net
		// 根据传输协议(network)修改streamSettings
		//TODO: QUIC
		switch strings.ToLower(v.Net) {
		case "grpc":
			o.StreamSettings.GrpcSettings = &GrpcSettings{ServiceName: v.Path}
		case "ws":
			wsSetting := tmplJson.TmplWsSettings
			wsSetting.Headers.Host = v.Host
			wsSetting.Path = v.Path
			o.StreamSettings.WsSettings = &wsSetting
		case "mkcp", "kcp":
			kcpSetting := tmplJson.TmplKcpSettings
			kcpSetting.Header.Type = v.Type
			o.StreamSettings.KcpSettings = &kcpSetting
			o.StreamSettings.KcpSettings.Seed = v.Path
		case "tcp":
			if strings.ToLower(v.Type) == "http" {
				tcpSetting := tmplJson.TmplTCPSettings
				tcpSetting.Header.Request.Headers.Host = strings.Split(v.Host, ",")
				if v.Path != "" {
					tcpSetting.Header.Request.Path = strings.Split(v.Path, ",")
					for i := range tcpSetting.Header.Request.Path {
						if !strings.HasPrefix(tcpSetting.Header.Request.Path[i], "/") {
							tcpSetting.Header.Request.Path[i] = "/" + tcpSetting.Header.Request.Path[i]
						}
					}
				}
				o.StreamSettings.TCPSettings = &tcpSetting
			}
		case "h2", "http":
			httpSetting := tmplJson.TmplHttpSettings
			httpSetting.Host = strings.Split(v.Host, ",")
			httpSetting.Path = v.Path
			o.StreamSettings.HTTPSettings = &httpSetting
		}
		muxOn := setting.MuxOn == configure.Yes
		if strings.ToLower(v.TLS) == "tls" {
			o.StreamSettings.Security = "tls"
			tlsSetting := tmplJson.TmplTLSSettings
			o.StreamSettings.TLSSettings = &tlsSetting
			if v.AllowInsecure {
				o.StreamSettings.TLSSettings.AllowInsecure = true
			}
			ver, e := where.GetV2rayServiceVersion()
			if e != nil {
				log.Warn("cannot get the version of v2ray-core: %v", e)
			} else if !common.VersionMustGreaterEqual(ver, "4.23.2") {
				o.StreamSettings.TLSSettings.AllowInsecureCiphers = true
			}
			// SNI
			if v.Host != "" {
				o.StreamSettings.TLSSettings.ServerName = v.Host
			}
			// Alpn
			if v.Alpn != "" {
				alpn := strings.Split(v.Alpn, ",")
				for i := range alpn {
					alpn[i] = strings.TrimSpace(alpn[i])
				}
				o.StreamSettings.TLSSettings.Alpn = alpn
			}
		} else if strings.ToLower(v.TLS) == "xtls" {
			o.StreamSettings.Security = "xtls"
			o.StreamSettings.XTLSSettings = new(TLSSettings)
			// always set SNI
			if v.Host != "" {
				o.StreamSettings.XTLSSettings.ServerName = v.Host
			}
			if v.Flow == "" {
				v.Flow = "xtls-rprx-origin"
			}
			if v.Alpn != "" {
				alpn := strings.Split(v.Alpn, ",")
				for i := range alpn {
					alpn[i] = strings.TrimSpace(alpn[i])
				}
				o.StreamSettings.TLSSettings.Alpn = alpn
			}
			vnext := o.Settings.Vnext.([]Vnext)
			vnext[0].Users[0].Flow = v.Flow
			o.Settings.Vnext = vnext
			//xtls does not support mux
			muxOn = false
		}
		o.Mux = &Mux{
			Enabled:     muxOn,
			Concurrency: setting.Mux,
		}
	case "shadowsocks":
		v.Net = strings.ToLower(v.Net)
		switch v.Net {
		case "aes-256-gcm", "aes-128-gcm", "chacha20-poly1305", "chacha20-ietf-poly1305", "plain", "none":
		default:
			return o, fmt.Errorf("unsupported shadowsocks encryption method: %v", v.Net)
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
			return o, fmt.Errorf("unsupported shadowsocks obfuscation method: %v", v.TLS)
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
			return o, fmt.Errorf("ResolveOutbound: %v", err)
		}
		if ok, err := common.VersionGreaterEqual(version, "4.31.0"); err != nil || !ok {
			return o, fmt.Errorf("unsupported shadowsocks obfuscation method: %v", v.TLS)
		}
		o.Settings.Servers = []Server{{
			Address:  v.Add,
			Port:     port,
			Password: v.ID,
		}}

		//tls
		streamSetting := tmplJson.TmplStreamSettings
		o.StreamSettings = &streamSetting
		o.StreamSettings.Network = "tcp"
		o.StreamSettings.Security = "tls"
		tlsSetting := tmplJson.TmplTLSSettings
		o.StreamSettings.TLSSettings = &tlsSetting
		if v.AllowInsecure {
			o.StreamSettings.TLSSettings.AllowInsecure = true
		}
		if v.Host != "" {
			o.StreamSettings.TLSSettings.ServerName = v.Host
		} else {
			o.StreamSettings.TLSSettings.ServerName = v.Add
		}
	case "http2":
		//TODO:
		return OutboundObject{}, fmt.Errorf("TODO")
	case "http", "https":
		o.Protocol = "http"
		var users []OutboundUser
		if v.ID != "" && v.Aid != "" {
			users = []OutboundUser{
				{
					User: v.ID,
					Pass: v.Aid,
				},
			}
		}
		o.Settings.Servers = []Server{{
			Address: v.Add,
			Port:    port,
			Users:   users,
		}}
		if v.Protocol == "https" {
			//tls
			streamSetting := tmplJson.TmplStreamSettings
			o.StreamSettings = &streamSetting
			o.StreamSettings.Network = "tcp"
			o.StreamSettings.Security = "tls"
			tlsSetting := tmplJson.TmplTLSSettings
			o.StreamSettings.TLSSettings = &tlsSetting
			o.StreamSettings.TLSSettings.ServerName = v.Add
		}
	case "shadowsocksr":
		v.Net = strings.ToLower(v.Net)
		switch v.Net {
		case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb", "none":
		default:
			return o, fmt.Errorf("unsupported shadowsocks encryption method: %v", v.Net)
		}
		if len(strings.TrimSpace(v.Type)) <= 0 {
			v.Type = "origin"
		}
		switch v.Type {
		case "origin", "verify_sha1", "auth_sha1_v4", "auth_aes128_md5", "auth_aes128_sha1", "auth_chain_a", "auth_chain_b":
		default:
			return o, fmt.Errorf("unsupported shadowsocksR protocol: %v", v.Type)
		}
		if len(strings.TrimSpace(v.TLS)) <= 0 {
			v.TLS = "plain"
		}
		switch v.TLS {
		case "plain", "http_simple", "http_post", "random_head", "tls1.2_ticket_auth":
		default:
			return o, fmt.Errorf("unsupported shadowsocksr obfuscation method: %v", v.TLS)
		}
		socksPlugin = true
	case "pingtunnel":
		socksPlugin = true
	case "trojan-go":
		socksPlugin = true
	default:
		return o, fmt.Errorf("unsupported protocol: %v", v.Protocol)
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
	return
}

type Addr struct {
	host string
	port string
	udp  bool
}

func parseDnsAddr(addr string) Addr {
	// 223.5.5.5
	if net.ParseIP(addr) != nil {
		return Addr{
			host: addr,
			port: "53",
			udp:  true,
		}
	}
	// dns.google:53
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if _, err = strconv.Atoi(port); err == nil {
			return Addr{
				host: host,
				port: port,
				udp:  true,
			}
		}
	}
	// tcp://8.8.8.8:53, https://dns.google/dns-query
	// TODO: quic:// uses UDP
	if u, err := url.Parse(addr); err == nil {
		return Addr{
			host: u.Hostname(),
			port: u.Port(),
			udp:  false,
		}
	}
	// dns.google, dns.pub, etc.
	return Addr{
		host: addr,
		port: "53",
		udp:  true,
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

// outName -> isGroup
func (t *Template) outNames() map[string]bool {
	tags := make(map[string]bool)
	for _, o := range t.Outbounds {
		if len(o.balancers) > 0 {
			for _, groupName := range o.balancers {
				tags[groupName] = true
			}
		} else {
			tags[o.Tag] = false
		}
	}
	return tags
}

func (t *Template) FirstProxyOutboundName(filter func(outboundName string, isGroup bool) bool) (outboundName string, isGroup bool) {
	if filter == nil {
		filter = func(outboundName string, isGroup bool) bool {
			return true
		}
	}
	// deduplicate
	m := make(map[string]struct{})

	for _, o := range t.Outbounds {
		switch o.Tag {
		case "direct", "block", "dns-out":
			continue
		}
		if len(o.balancers) > 0 {
			for _, v := range o.balancers {
				if _, ok := m[v]; !ok {
					if filter(v, true) {
						return v, true
					}
					m[v] = struct{}{}
				}
			}
		} else {
			if filter(o.Tag, false) {
				return o.Tag, false
			}
		}
	}
	return
}

func (t *Template) SetDNS(outbounds []serverInfo, setting *configure.Setting, supportUDP map[string]bool) (routing []RoutingRule, err error) {
	firstOutboundTag, _ := t.FirstProxyOutboundName(nil)
	firstUDPSupportedOutboundTag, _ := t.FirstProxyOutboundName(func(outboundName string, isGroup bool) bool {
		return supportUDP[outboundName]
	})
	outboundTags := t.outNames()
	var internal, external, all []string
	var allThroughProxy = false
	if setting.AntiPollution == configure.AntipollutionAdvanced {
		// advanced
		internal = configure.GetInternalDnsListNotNil()
		external = configure.GetExternalDnsListNotNil()
		all = append(all, internal...)
		all = append(all, external...)
		if len(external) == 0 {
			allThroughProxy = true
			for _, line := range internal {
				dns := dnsParser2.Parse(line)
				if dns.Out == "direct" {
					allThroughProxy = false
					break
				}
			}
		}
		// check if outbounds exist
		for _, line := range all {
			dns := dnsParser2.Parse(line)
			if _, ok := outboundTags[dns.Out]; !ok {
				return nil, fmt.Errorf(`your DNS rule "%v" depends on the outbound "%v", thus it should connect to a server`, line, dns.Out)
			}
		}
		// check UDP support
		for _, line := range all {
			dns := dnsParser2.Parse(line)
			if dns.Out == "direct" || dns.Out == "block" {
				continue
			}
			if parseDnsAddr(dns.Val).udp && !supportUDP[dns.Out] {
				return nil, fmt.Errorf(`due to the protocol of outbound "%v" with no UDP supported, please use tcp:// and doh:// DNS rule instead, or change the connected server`, dns.Out)
			}
		}
	} else if setting.AntiPollution != configure.AntipollutionClosed {
		// preset
		internal = []string{"223.6.6.6 -> direct", "119.29.29.29 -> direct"}
		switch setting.AntiPollution {
		case configure.AntipollutionAntiHijack:
			break
		case configure.AntipollutionDnsForward:
			if firstUDPSupportedOutboundTag != "" {
				external = []string{"8.8.8.8 -> " + firstUDPSupportedOutboundTag, "1.1.1.1 -> " + firstUDPSupportedOutboundTag}
			} else {
				if err := CheckTcpDnsSupported(); err == nil {
					external = []string{"tcp://dns.opendns.com:5353 -> " + firstOutboundTag, "tcp://dns.google -> " + firstOutboundTag}
				} else if err = CheckDohSupported(); err == nil {
					external = []string{"https://1.1.1.1/dns-query -> " + firstOutboundTag, "https://dns.google/dns-query -> " + firstOutboundTag}
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

		if net.ParseIP(addr.host) != nil {
			routing = append(routing, RoutingRule{
				Type: "field", InboundTag: []string{"dns"}, OutboundTag: dns.Out, IP: []string{addr.host}, Port: addr.port,
			})
		} else {
			routing = append(routing, RoutingRule{
				Type: "field", InboundTag: []string{"dns"}, OutboundTag: dns.Out, Domain: []string{addr.host}, Port: addr.port,
			})
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
			log.Error("[Fakedns]: NOT REASONABLE. Please report your config.")
			t.DNS.Servers = append(t.DNS.Servers, "localhost")
		}
		t.DNS.Servers = append(t.DNS.Servers, ds)
	}

	if t.DNS.Servers == nil {
		t.DNS.Servers = []interface{}{"localhost"}
	}
	var domainsToLookup []string
	for _, v := range outbounds {
		if net.ParseIP(v.Info.Add) == nil {
			domainsToLookup = append(domainsToLookup, v.Info.Add)
		}
	}
	for _, r := range routing {
		if len(r.Domain) > 0 {
			domainsToLookup = append(domainsToLookup, r.Domain...)
		}
	}
	domainsToLookup = common.Deduplicate(domainsToLookup)
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
		ips, err := resolv.LookupHost(domain)
		if err != nil {
			return routing, fmt.Errorf("[Error] %w: please make sure you're connected to the Internet", err)
		}
		if t.DNS.Hosts == nil {
			t.DNS.Hosts = make(Hosts)
		}
		ips = FilterIPs(ips)
		if CheckHostsListSupported() == nil {
			t.DNS.Hosts[domain] = ips
		} else {
			t.DNS.Hosts[domain] = ips[0]
		}
	}
	return
}

// The order are from v4 IPs to v6 IPs. If the system does not support IPv6, v6 IPs will not be returned.
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
func (t *Template) SetDNSRouting(routing []RoutingRule, supportUDP map[string]bool) {
	firstOutboundTag, _ := t.FirstProxyOutboundName(nil)
	t.Routing.Rules = append(t.Routing.Rules, routing...)
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
		if specialMode.ShouldLocalDnsListen() {
			if couldListenLocalhost, _ := specialMode.CouldLocalDnsListen(); couldListenLocalhost {
				dnsOut.InboundTag = []string{"dns-in"}
			}
		}
		t.Routing.Rules = append(t.Routing.Rules, dnsOut)
	}
	if specialMode.ShouldUseSupervisor() || specialMode.ShouldUseFakeDns() {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				IP:          []string{"198.18.0.0/15"},
				OutboundTag: firstOutboundTag,
			},
		)
	}
	if !supportUDP[firstOutboundTag] {
		// find a outbound that supports UDP and redirect all leaky UDP traffic to it
		var found bool
		for outbound, support := range supportUDP {
			if support {
				t.Routing.Rules = append(t.Routing.Rules,
					RoutingRule{
						Type:        "field",
						OutboundTag: outbound,
						Network:     "udp",
					},
				)
				found = true
				break
			}
		}
		if !found {
			// no outbound with UDP supported, so redirect all leaky UDP traffic to outbound direct
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: "direct",
					Network:     "udp",
				},
			)
		}
	} else {
		if setting.Transparent != configure.TransparentClose {
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: "direct",
					InboundTag: []string{
						"transparent",
					},
					Port: "53",
					IP:   []string{"geoip:private"},
				})
		}
	}
	return
}

func (t *Template) AppendRoutingRuleByMode(mode configure.RulePortMode, inbounds []string) (err error) {
	firstOutboundTag, _ := t.FirstProxyOutboundName(nil)
	switch mode {
	case configure.WhitelistMode:
		// foreign domains with intranet IP should be proxied first rather than directly connected
		if asset.LoyalsoldierSiteDatExists() {
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  inbounds,
					Domain:      []string{"ext:LoyalsoldierSite.dat:geolocation-!cn"},
				})
		} else {
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  inbounds,
					Domain:      []string{"geosite:geolocation-!cn"},
				})
		}
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  inbounds,
				Domain:      []string{"geosite:cn"},
			},
			RoutingRule{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  inbounds,
				IP:          []string{"geoip:hk", "geoip:mo"},
			},
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  inbounds,
				IP:          []string{"geoip:private", "geoip:cn"},
			},
		)
	case configure.GfwlistMode:
		if asset.LoyalsoldierSiteDatExists() {
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  inbounds,
					Domain:      []string{"ext:LoyalsoldierSite.dat:gfw"},
				},
				RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  inbounds,
					Domain:      []string{"ext:LoyalsoldierSite.dat:greatfire"},
				})
		} else {
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  inbounds,
					Domain:      []string{"geosite:geolocation-!cn"},
				})
		}
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"rule"},
			},
		)
	case configure.RoutingAMode:
		if err := parseRoutingA(t, []string{"rule"}); err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) SetRulePortRouting(setting *configure.Setting) error {
	return t.AppendRoutingRuleByMode(setting.RulePortMode, []string{"rule"})
}
func parseRoutingA(t *Template, routingInboundTags []string) error {
	ra := configure.GetRoutingA()
	rules, err := routingA.Parse(ra)
	if err != nil {
		log.Warn("parseRoutingA: %v", err)
		return err
	}
	defaultOutbound, _ := t.FirstProxyOutboundName(nil)
	for _, rule := range rules {
		switch rule := rule.(type) {
		case routingA.Define:
			switch rule.Name {
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
							t.Outbounds = append(t.Outbounds, OutboundObject{
								Tag:      o.Name,
								Protocol: o.Value.Name,
								Settings: Settings{
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
						settings := Settings{}
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
						t.Outbounds = append(t.Outbounds, OutboundObject{
							Tag:      o.Name,
							Protocol: o.Value.Name,
							Settings: settings,
						})
					}
				}
			}
		}
	}
	outboundTags := t.outNames()
	for _, rule := range rules {
		switch rule := rule.(type) {
		case routingA.Define:
			switch rule.Name {
			case "default":
				switch v := rule.Value.(type) {
				case string:
					defaultOutbound = v
					if _, ok := outboundTags[v]; !ok {
						return fmt.Errorf(`your RoutingA rules depend on the outbound "%v", thus it should connect to a server`, v)
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
				case "domain", "domains":
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
				case "sourcePort":
					rr.SourcePort = strings.Join(f.Params, ",")
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
			if rr.OutboundTag != "" {
				if _, ok := outboundTags[rr.OutboundTag]; !ok {
					return fmt.Errorf(`your RoutingA rules depend on the outbound "%v", thus it should connect to a server`, rr.OutboundTag)
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
func (t *Template) SetTransparentRouting(setting *configure.Setting) (err error) {
	switch setting.Transparent {
	case configure.TransparentProxy:
	case configure.TransparentWhitelist:
		return t.AppendRoutingRuleByMode(configure.WhitelistMode, []string{"transparent"})
	case configure.TransparentGfwlist:
		return t.AppendRoutingRuleByMode(configure.GfwlistMode, []string{"transparent"})
	case configure.TransparentFollowRule:
		// transparent mode is the same as rule
		for i := range t.Routing.Rules {
			ok := false
			for _, in := range t.Routing.Rules[i].InboundTag {
				if in == "rule" {
					ok = true
					break
				}
			}
			if ok {
				t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, "transparent")
			}
		}
	}
	return nil
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
		dokodemo.StreamSettings = &StreamSettings{Sockopt: &Sockopt{Tproxy: tproxy}}
		dokodemo.Settings.FollowRedirect = true

	}
	t.Inbounds = append(t.Inbounds, dokodemo)
}

func (t *Template) SetOutboundSockopt(setting *configure.Setting) {
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
func (t *Template) SetDualStack(setting *configure.Setting) {
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
			if t.Inbounds[i].Tag == "dns-in" {
				t.Inbounds[i].Listen = "127.2.0.17"
				inbounds6[i].Tag = "THIS_IS_A_DROPPED_TAG"
				continue
			} else {
				t.Inbounds[i].Listen = "127.0.0.1"
			}
			inbounds6[i].Listen = "::1"
			if t.Inbounds[i].Tag != "" {
				tagMap[t.Inbounds[i].Tag] = struct{}{}
				t.Inbounds[i].Tag += tag4Suffix
				inbounds6[i].Tag += tag6Suffix
			}
		}
		for i := len(inbounds6) - 1; i >= 0; i-- {
			if inbounds6[i].Tag == "THIS_IS_A_DROPPED_TAG" {
				inbounds6 = append(inbounds6[:i], inbounds6[i+1:]...)
			}
		}

		if iptables.IsIPv6Supported() {
			t.Inbounds = append(t.Inbounds, inbounds6...)
		}

		// set routing
		for i := range t.Routing.Rules {
			tag6 := make([]string, 0)
			for j, tag := range t.Routing.Rules[i].InboundTag {
				if _, ok := tagMap[tag]; ok {
					t.Routing.Rules[i].InboundTag[j] += tag4Suffix
					tag6 = append(tag6, tag+tag6Suffix)
				}
			}
			if len(tag6) > 0 && iptables.IsIPv6Supported() {
				t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, tag6...)
			}
		}
	} else {
		// specially listen 127.2.0.17
		hasDnsIn := false
		for i := range t.Inbounds {
			if t.Inbounds[i].Tag == "dns-in" {
				if couldListenLocalhost, e := specialMode.CouldLocalDnsListen(); couldListenLocalhost && e != nil {
					// listen only 127.2.0.17
					t.Inbounds[i].Listen = "127.2.0.17"
				} else {
					// listen both 0.0.0.0 and 127.2.0.17
					localDnsInbound := t.Inbounds[i]
					localDnsInbound.Listen = "127.2.0.17"
					localDnsInbound.Tag = "dns-in-local"
					t.Inbounds = append(t.Inbounds, localDnsInbound)
					hasDnsIn = true
				}
				break
			}
		}
		if hasDnsIn {
			// set routing
			for i := range t.Routing.Rules {
				for _, tag := range t.Routing.Rules[i].InboundTag {
					if tag == "dns-in" {
						t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, "dns-in-local")
					}
				}
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
	}
}

func (t *Template) AppendDNSOutbound() {
	t.Outbounds = append(t.Outbounds, OutboundObject{
		Tag:      "dns-out",
		Protocol: "dns",
	})
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

func SetVlessGrpcInbound(vlessGrpc *Inbound) (err error) {
	config := conf.GetEnvironmentConfig()
	if len(config.VlessGrpcInboundCertKey) < 2 {
		return fmt.Errorf("VLESS-GPRC inbound depends on TLS cert, close the inbound or add cert by comand line argument --vless-grpc-inbound-cert-key or environment variable V2RAYA_VLESS_GRPC_INBOUND_CERT_KEY")
	}
	cert, key := config.VlessGrpcInboundCertKey[0], config.VlessGrpcInboundCertKey[1]
	vlessGrpc.StreamSettings.TLSSettings.Certificates[0].CertificateFile = cert
	vlessGrpc.StreamSettings.TLSSettings.Certificates[0].KeyFile = key
	id, err := GenerateIdFromAccounts()
	if err != nil {
		return err
	}
	vlessGrpc.Settings.Clients = []VlessClient{{Id: id}}
	return nil
}

func (t *Template) SetInbound(setting *configure.Setting) error {
	p := configure.GetPortsNotNil()
	if p != nil {
		t.Inbounds[0].Port = p.Socks5
		t.Inbounds[1].Port = p.Http
		t.Inbounds[2].Port = p.HttpWithPac
		vlessGrpc := &t.Inbounds[3]
		vlessGrpc.Port = p.VlessGrpc
		if p.VlessGrpc > 0 {
			if err := SetVlessGrpcInbound(vlessGrpc); err != nil {
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
	if setting.Transparent != configure.TransparentClose {
		var tproxy string
		switch setting.TransparentType {
		case configure.TransparentTproxy, configure.TransparentRedirect:
			tproxy = string(setting.TransparentType)
		}
		t.AppendDokodemo(&tproxy, 32345, "transparent")
	}
	if specialMode.ShouldLocalDnsListen() {
		if couldListenLocalhost, _ := specialMode.CouldLocalDnsListen(); couldListenLocalhost {
			// FIXME: xray cannot use fakedns+others (2021-07-17)
			// set up a solo dokodemo-door for dns
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
	return nil
}

type serverInfo struct {
	Info         vmessInfo.VmessInfo
	OutboundName string
	PluginPort   int
}

func Ps2OutboundTag(ps string) string {
	return fmt.Sprintf("『%v』", ps)
}

func (t *Template) UpdatePrivateRouting() {
	privateAddrs, _ := iptables.GetLocalCIDR()
	if len(privateAddrs) == 0 {
		return
	}
	for i := range t.Routing.Rules {
		for j := range t.Routing.Rules[i].IP {
			if t.Routing.Rules[i].IP[j] == "geoip:private" {
				t.Routing.Rules[i].IP = append(t.Routing.Rules[i].IP, privateAddrs...)
				break
			}
		}
	}
}

func (t *Template) OptimizeGeoipMemoryOccupation() {
	if asset.IsGeoipOnlyCnPrivateExists() {
		for i := range t.Routing.Rules {
			for j := range t.Routing.Rules[i].IP {
				switch t.Routing.Rules[i].IP[j] {
				case "geoip:private", "geoip:cn":
					t.Routing.Rules[i].IP[j] = "ext:geoip-only-cn-private.dat:" + strings.TrimPrefix(t.Routing.Rules[i].IP[j], "geoip:")
				}
			}
		}
	}
}

func (t *Template) SetWhitelistRouting(whitelist []Addr) {
	var rules []RoutingRule
	for _, addr := range whitelist {
		if net.ParseIP(addr.host) != nil {
			rules = append(rules, RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				IP:          []string{addr.host},
				Port:        addr.port,
			})
		} else {
			rules = append(rules, RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				Domain:      []string{addr.host},
				Port:        addr.port,
			})
		}
	}
	if len(rules) > 0 {
		t.Routing.Rules = append(rules, t.Routing.Rules...)
	}
}

func (t *Template) SetGroupRouting(outboundName2VmessInfos map[string][]vmessInfo.VmessInfo) (err error) {
	outbounds := t.outNames()
	mSubjectSelector := make(map[string]struct{})
	for outbound, isGroup := range outbounds {
		if !isGroup {
			continue
		}

		strategy := "leastPing"
		interval := 10 * time.Second
		var selector []string

		for _, vi := range outboundName2VmessInfos[outbound] {
			selector = append(selector, Ps2OutboundTag(vi.Ps))
		}

		t.Routing.Balancers = append(t.Routing.Balancers, Balancer{
			Tag:      outbound,
			Selector: selector,
			Strategy: BalancerStrategy{
				//TODO: configure.GetOutboundSetting
				Type: strategy,
			},
		})

		if strategy == "leastPing" {
			if err = CheckObservatorySupported(); err != nil {
				return fmt.Errorf("not support observatory based load balance: %w", err)
			}
			if t.Observatory == nil {
				t.Observatory = &Observatory{
					ProbeURL:      "http://www.msftconnecttest.com/connecttest.txt",
					ProbeInterval: interval.String(),
				}
			}
			for _, s := range selector {
				mSubjectSelector[s] = struct{}{}
			}
		}
	}
	if t.Observatory != nil {
		var subjectSelector []string
		for s := range mSubjectSelector {
			subjectSelector = append(subjectSelector, s)
		}
		t.Observatory.SubjectSelector = subjectSelector
	}
	for i := range t.Routing.Rules {
		if t.Routing.Rules[i].OutboundTag != "" &&
			outbounds[t.Routing.Rules[i].OutboundTag] == true {
			t.Routing.Rules[i].BalancerTag, t.Routing.Rules[i].OutboundTag = t.Routing.Rules[i].OutboundTag, ""
		}
	}
	return nil
}

func RefineOutboundInfos(serverInfos []serverInfo) (
	vmessInfo2ServerInfos map[vmessInfo.VmessInfo][]*serverInfo,
	outboundName2VmessInfos map[string][]vmessInfo.VmessInfo,
) {
	// guarantee that an v2ray outbound is reusable for balancers
	vmessInfo2ServerInfos = make(map[vmessInfo.VmessInfo][]*serverInfo)
	for i, info := range serverInfos {
		vmessInfo2ServerInfos[info.Info] = append(vmessInfo2ServerInfos[info.Info], &serverInfos[i])
	}
	// make ps unique
	vmessInfo2OutboundInfoAfter := make(map[vmessInfo.VmessInfo][]*serverInfo)
	mPsRenaming := make(map[string]struct{})
	for vi, ois := range vmessInfo2ServerInfos {
		ps := vi.Ps
		cnt := 2
		for {
			if _, ok := mPsRenaming[ps]; !ok {
				mPsRenaming[ps] = struct{}{}
				vi.Ps = ps
				vmessInfo2OutboundInfoAfter[vi] = ois
				break
			}
			ps = fmt.Sprintf("%v(%v)", vi.Ps, strconv.Itoa(cnt))
			cnt++
		}
	}
	outboundName2VmessInfos = make(map[string][]vmessInfo.VmessInfo)
	for vi, ois := range vmessInfo2OutboundInfoAfter {
		for _, oi := range ois {
			outboundName2VmessInfos[oi.OutboundName] = append(outboundName2VmessInfos[oi.OutboundName], vi)
		}
	}
	return vmessInfo2OutboundInfoAfter, outboundName2VmessInfos
}

func SupportUDP(v vmessInfo.VmessInfo) bool {
	if plugin.HasProperPlugin(v) {
		return false
	}
	switch v.Protocol {
	case "http", "https", "socks":
		return false
	}
	return true
}

func (t *Template) ResolveOutbounds(
	serverInfos []serverInfo,
	vmessInfo2ServerInfos map[vmessInfo.VmessInfo][]*serverInfo,
	outboundName2VmessInfos map[string][]vmessInfo.VmessInfo) (supportUDP map[string]bool, outboundTags []string, err error) {

	supportUDP = make(map[string]bool)
	type _outbound struct {
		weight   int
		outbound OutboundObject
		balancer bool
	}
	serverInfo2Index := make(map[*serverInfo]int)
	for i := range serverInfos {
		serverInfo2Index[&serverInfos[i]] = i
	}
	// keep order with serverInfos
	outboundTags = make([]string, len(serverInfos))
	var outbounds []_outbound
	for vmessinfo, sInfos := range vmessInfo2ServerInfos {
		vi := vmessinfo
		var (
			usedByBalancer     bool
			balancerPluginPort int
		)
		// a vmessInfo(server template) may is used by multiple serverInfos(a connected server)

		// outbound name is not just v2ray outbound tag, it may is a balancer
		type balancer struct {
			name       string
			serverInfo *serverInfo
		}
		var balancers []balancer
		for _, sInfo := range sInfos {
			if len(outboundName2VmessInfos[sInfo.OutboundName]) > 1 {
				// balancer
				if err = CheckBalancerSupported(); err != nil {
					return nil, nil, err
				}
				if !usedByBalancer {
					usedByBalancer = true
					balancerPluginPort = sInfo.PluginPort
				}
				balancers = append(balancers, balancer{
					name:       sInfo.OutboundName,
					serverInfo: sInfo,
				})
			} else {
				// pure outbound
				outboundTag := sInfo.OutboundName
				o, err := ResolveOutbound(&vi, outboundTag, &sInfo.PluginPort)
				if err != nil {
					return nil, nil, err
				}
				outbounds = append(outbounds, _outbound{
					weight:   serverInfo2Index[sInfo],
					outbound: o,
					balancer: false,
				})
				outboundTags[serverInfo2Index[sInfo]] = outboundTag
				supportUDP[sInfo.OutboundName] = SupportUDP(sInfo.Info)
			}
		}
		if usedByBalancer {
			// the v2ray outbound is shared by balancers
			outboundTag := Ps2OutboundTag(vi.Ps)
			o, err := ResolveOutbound(&vi, outboundTag, &balancerPluginPort)
			if err != nil {
				return nil, nil, err
			}
			for _, v := range balancers {
				o.balancers = append(o.balancers, v.name)
			}

			// we use the lowest serverInfo index as the order weight of the balancer outbound
			weight := -1
			for _, v := range balancers {
				index := serverInfo2Index[v.serverInfo]
				if weight == -1 || weight > index {
					weight = index
				}
				// tag
				outboundTags[index] = outboundTag
			}
			outbounds = append(outbounds, _outbound{
				weight:   weight,
				outbound: o,
				balancer: true,
			})

			// if any node does not support UDP, the outbound should be tagged as UDP unsupported
			for _, outboundName := range o.balancers {
				_supportUDP := SupportUDP(vi)
				if _, ok := supportUDP[outboundName]; !ok {
					supportUDP[outboundName] = _supportUDP
				}
				if supportUDP[outboundName] && !_supportUDP {
					supportUDP[outboundName] = false
				}
			}
		}
	}
	sort.Slice(outbounds, func(i, j int) bool {
		return outbounds[i].weight < outbounds[j].weight
	})
	for _, v := range outbounds {
		t.Outbounds = append(t.Outbounds, v.outbound)
	}
	t.Outbounds = append(t.Outbounds, OutboundObject{
		Tag:      "direct",
		Protocol: "freedom",
	}, OutboundObject{
		Tag:      "block",
		Protocol: "blackhole",
	})
	return supportUDP, outboundTags, nil
}

func (t *Template) SetAPI() (port int) {
	// TODO: refine code
	for _, closeFunc := range apiCloseFuncs {
		closeFunc()
	}
	apiCloseFuncs = []func(){}
	services := []string{
		"LoggerService",
	}
	if t.Observatory != nil {
		services = append(services, "ObservatoryService")
		apiCloseFuncs = append(apiCloseFuncs, ObservatoryProducer())
	}
	t.API = &APIObject{
		Tag:      "api-out",
		Services: services,
	}
	// find a valid port
	for {
		if l, err := net.Listen("tcp4", "127.0.0.1:0"); err == nil {
			port = l.Addr().(*net.TCPAddr).Port
			_ = l.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Inbounds = append(t.Inbounds, Inbound{
		Port:     port,
		Protocol: "dokodemo-door",
		Listen:   "127.0.0.1",
		Settings: &InboundSettings{
			Address: "127.0.0.1",
		},
		Tag: "api-in",
	})
	t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
		Type:        "field",
		InboundTag:  []string{"api-in"},
		OutboundTag: "api-out",
	})
	apiPort = port
	return port
}

func (t *Template) SetVlessGrpcRouting() {
	if configure.GetPortsNotNil().VlessGrpc <= 0 {
		return
	}
	for i := range t.Routing.Rules {
		var bHasRule bool
		for _, tag := range t.Routing.Rules[i].InboundTag {
			if tag == "rule" {
				bHasRule = true
			}
		}
		if bHasRule {
			t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, "vlessGrpc")
		}
	}
}

func NewTemplate(serverInfos []serverInfo) (t Template, outboundTags []string, err error) {
	vmessInfo2ServerInfos, outboundName2VmessInfos := RefineOutboundInfos(serverInfos)
	ps2OutboundNames := make(map[string][]string)
	for outboundName, vis := range outboundName2VmessInfos {
		for _, vi := range vis {
			ps2OutboundNames[vi.Ps] = append(ps2OutboundNames[vi.Ps], outboundName)
		}
	}
	setting := configure.GetSettingNotNil()
	var tmplJson TmplJson
	// read template json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return Template{}, nil, fmt.Errorf("error occurs while reading template json, please check whether templateJson variable is correct json format")
	}
	// tmplJson.Template is the basic configuration
	t = tmplJson.Template
	// log
	t.Log = new(Log)
	if logLevel := log.ParseLevel(conf.GetEnvironmentConfig().LogLevel); logLevel >= log.ParseLevel("debug") {
		t.Log.Loglevel = "info"
		t.Log.Access = ""
		t.Log.Error = ""
	} else if logLevel >= log.ParseLevel("info") && CheckLogNoneSupported() == nil {
		t.Log.Loglevel = "info"
		t.Log.Access = ""
		t.Log.Error = "none"
	} else {
		t.Log = nil
	}
	// fakedns
	if specialMode.ShouldUseFakeDns() && CheckFakednsAutoConfigureSupported() != nil {
		t.FakeDns = &FakeDns{
			IpPool:   "198.18.0.0/15",
			PoolSize: 65535,
		}
	}
	// resolve Outbounds
	supportUDP, outboundTags, err := t.ResolveOutbounds(serverInfos, vmessInfo2ServerInfos, outboundName2VmessInfos)
	if err != nil {
		return Template{}, nil, err
	}

	//set inbound ports according to the setting
	if err = t.SetInbound(setting); err != nil {
		return Template{}, nil, err
	}
	//set DNS
	dnsRouting, err := t.SetDNS(serverInfos, setting, supportUDP)
	if err != nil {
		return Template{}, nil, err
	}
	//append a DNS outbound
	t.AppendDNSOutbound()
	//DNS routing
	t.Routing.DomainMatcher = "mph"
	t.SetDNSRouting(dnsRouting, supportUDP)
	//rule port routing
	if err = t.SetRulePortRouting(setting); err != nil {
		return Template{}, nil, err
	}
	//transparent routing
	if setting.Transparent != configure.TransparentClose {
		if err = t.SetTransparentRouting(setting); err != nil {
			return Template{}, nil, err
		}
	}
	//set group routing
	if err = t.SetGroupRouting(outboundName2VmessInfos); err != nil {
		return Template{}, nil, err
	}
	//set vlessGrpc routing
	t.SetVlessGrpcRouting()
	// set api
	t.SetAPI()

	// set routing whitelist
	var whitelist []Addr
	for _, info := range serverInfos {
		whitelist = append(whitelist, Addr{
			host: info.Info.Add,
			port: info.Info.Port,
		})
	}
	t.SetWhitelistRouting(whitelist)

	t.UpdatePrivateRouting()

	t.OptimizeGeoipMemoryOccupation()

	//set outboundSockopt
	t.SetOutboundSockopt(setting)

	//set fakedns destOverride
	t.SetInboundFakeDnsDestOverride()

	//set inbound listening address and routing
	t.SetDualStack(setting)

	//check if there are any duplicated tags
	if err = t.CheckDuplicatedTags(); err != nil {
		return Template{}, nil, err
	}
	//check if there are any duplicated inbound ports
	if err = t.CheckDuplicatedInboundSockets(); err != nil {
		return Template{}, nil, err
	}

	return t, outboundTags, nil
}

func (t *Template) CheckDuplicatedTags() error {
	inboundTagsSet := make(map[string]interface{})
	for _, in := range t.Inbounds {
		tag := in.Tag
		if _, exists := inboundTagsSet[tag]; exists {
			return fmt.Errorf("duplicated inbound tag: %v", tag)
		} else {
			inboundTagsSet[tag] = nil
		}
	}
	outboundTagsSet := make(map[string]interface{})
	for _, out := range t.Outbounds {
		tag := out.Tag
		if _, exists := outboundTagsSet[tag]; exists {
			return fmt.Errorf("duplicated outbound tag: %v", tag)
		} else {
			outboundTagsSet[tag] = nil
		}
	}
	return nil
}

func (t *Template) CheckDuplicatedInboundSockets() error {
	inboundSocketSet := make(map[string]interface{})
	for _, in := range t.Inbounds {
		if in.Listen == "" {
			// https://www.v2fly.org/config/inbounds.html#inboundobject
			in.Listen = "0.0.0.0"
		}
		socket := net.JoinHostPort(in.Listen, strconv.Itoa(in.Port))
		if _, exists := inboundSocketSet[socket]; exists {
			return fmt.Errorf("duplicated inbound listening address: %v", socket)
		} else {
			inboundSocketSet[socket] = nil
		}
	}
	return nil
}

var OccupiedErr = fmt.Errorf("port is occupied")

func PortOccupied(syntax []string) (err error) {
	occupied, sockets, err := ports.IsPortOccupied(syntax)
	if err != nil {
		if errors.Is(err, netstat.ErrorNotSupportOSErr) {
			log.Trace("PortOccupied: %v", err)
			return nil
		}
		return
	}
	if occupied {
		if err = netstat.FillProcesses(sockets); err != nil {
			if errors.Is(err, netstat.ErrorNotSupportOSErr) {
				log.Warn("cannot judge port occupation: %v", err)
				return nil
			}
			return fmt.Errorf("failed to check if port is occupied: %w", err)
		}
		for _, s := range sockets {
			p := s.Proc
			if p == nil {
				continue
			}
			if ownPID := strconv.Itoa(os.Getpid());
				p.PPID == ownPID ||
					p.PID == ownPID {
				continue
			}
			occupiedErr := fmt.Errorf("%w by %v(%v): %v", OccupiedErr, p.Name, p.PID, s.LocalAddress.Port)
			if configure.GetSettingNotNil().IntranetSharing {
				// want to listen 0.0.0.0, which conflicts with all IPs
				return occupiedErr
			}
			if s.LocalAddress.IP.IsUnspecified() {
				return occupiedErr
			}
			if s.LocalAddress.IP.IsLoopback() {
				return occupiedErr
			}
		}
	}
	return nil
}

func (t *Template) CheckInboundPortsOccupied() (err error) {
	var st []string
	for _, in := range t.Inbounds {
		switch strings.ToLower(in.Protocol) {
		case "http", "vmess", "vless", "trojan":
			st = append(st, strconv.Itoa(in.Port)+":tcp")
		case "dokodemo-door":
			if strings.HasPrefix(in.Tag, "dns-in") {
				// checked before
				continue
			} else if in.Settings != nil && in.Settings.Network != "" {
				st = append(st, strconv.Itoa(in.Port)+":"+in.Settings.Network)
			} else {
				st = append(st, strconv.Itoa(in.Port)+":tcp,udp")
			}
		default:
			st = append(st, strconv.Itoa(in.Port)+":tcp,udp")
		}
	}
	return PortOccupied(st)
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}

func WriteV2rayConfig(content []byte) (err error) {
	err = os.WriteFile(asset.GetV2rayConfigPath(), content, os.FileMode(0600))
	if err != nil {
		return fmt.Errorf("WriteV2rayConfig: %w", err)
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

func checkAndSetMark(o *OutboundObject, mark int) {
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
		return fmt.Errorf("port of inbound must be a positive number with string type")
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
