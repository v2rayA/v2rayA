package v2ray

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"v2rayA/common"
	"v2rayA/core/dnsPoison/entity"
	"v2rayA/core/routingA"
	"v2rayA/core/v2ray/asset"
	"v2rayA/core/vmessInfo"
	"v2rayA/global"
	"v2rayA/persistence/configure"
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
	Log       Log        `json:"log"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   struct {
		DomainStrategy string        `json:"domainStrategy"`
		Rules          []RoutingRule `json:"rules"`
	} `json:"routing"`
	DNS *DNS `json:"dns,omitempty"`
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
	Clients        interface{} `json:"clients,omitempty"`
	Network        string      `json:"network,omitempty"`
	FollowRedirect bool        `json:"followRedirect,omitempty"`
}
type User struct {
	ID       string `json:"id"`
	AlterID  int    `json:"alterId"`
	Security string `json:"security"`
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
	ServerName           interface{} `json:"serverName"`
	AllowInsecureCiphers bool        `json:"allowInsecureCiphers"`
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
}
type HttpSettings struct {
	Path string   `json:"path"`
	Host []string `json:"host"`
}
type Hosts map[string]string

type DNS struct {
	Hosts   Hosts         `json:"hosts,omitempty"`
	Servers []interface{} `json:"servers"`
}
type DnsServer struct {
	Address string   `json:"address"`
	Port    int      `json:"port"`
	Domains []string `json:"domains,omitempty"`
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
		o.StreamSettings = &tmplJson.StreamSettings
		o.StreamSettings.Network = v.Net
		// 根据传输协议(network)修改streamSettings
		switch strings.ToLower(v.Net) {
		case "ws":
			tmplJson.WsSettings.Headers.Host = v.Host
			tmplJson.WsSettings.Path = v.Path
			o.StreamSettings.WsSettings = &tmplJson.WsSettings
		case "mkcp", "kcp":
			tmplJson.KcpSettings.Header.Type = v.Type
			o.StreamSettings.KcpSettings = &tmplJson.KcpSettings
		case "tcp":
			if strings.ToLower(v.Type) != "none" { //那就是http无疑了
				tmplJson.TCPSettings.Header.Request.Headers.Host = strings.Split(v.Host, ",")
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
			ver, e := GetV2rayServiceVersion()
			if e != nil {
				log.Println(newError("cannot get the version of v2ray-core").Base(e))
			} else if !common.VersionMustGreaterEqual(ver, "4.23.2") {
				o.StreamSettings.TLSSettings.AllowInsecureCiphers = true
			}
			// always set SNI
			if v.Host != "" {
				o.StreamSettings.TLSSettings.ServerName = v.Host
			}
		}
	case "shadowsocks", "shadowsocksr":
		v.Net = strings.ToLower(v.Net)
		switch v.Net {
		case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb":
		default:
			return o, newError("unsupported shadowsocks encryption method: " + v.Net)
		}
		if len(strings.TrimSpace(v.Type)) <= 0 {
			v.Type = "origin"
		}
		switch v.Type {
		case "origin", "verify_sha1", "auth_sha1_v4", "auth_aes128_md5", "auth_aes128_sha1":
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
	case "pingtunnel", "trojan":
	default:
		return o, newError("unsupported protocol: " + v.Protocol)
	}
	if v.Protocol != "vmess" && pluginPort != nil {
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

func (t *Template) SetDNS(v vmessInfo.VmessInfo, supportUDP bool, setting *configure.Setting) (dohIPs, dohDomains []string) {
	//先修改DNS设置
	t.DNS = new(DNS)
	switch setting.AntiPollution {
	case configure.DoH:
		//添加DoH服务器
		s := *configure.GetDohListNotNil()
		dohs := strings.Split(strings.TrimSpace(s), "\n")
		for _, doh := range dohs {
			//使用DOHL模式，默认绕过routing和outbound以加快DOH速度
			doh = strings.Replace(doh, "https://", "https+local://", 1)
			t.DNS.Servers = append(t.DNS.Servers, doh)
		}
	case configure.DnsForward:
		if supportUDP {
			t.DNS.Servers = []interface{}{
				DnsServer{
					Address: "8.8.8.8",
					Port:    53,
				},
			}
		} else {
			//由于plugin不支持udp
			//先优先请求DoH（tcp）
			if err := CheckDohSupported(); err == nil {
				//DNS转发，所以使用全球友好的DNS服务器
				t.DNS.Servers = []interface{}{
					"https://1.0.0.1/dns-query",
					"https://dns.google/dns-query",
				}
			}
			if len(t.DNS.Servers) <= 0 {
				//否则使用openDNS非标准端口直连5353
				t.DNS.Servers = []interface{}{
					DnsServer{
						Address: "208.67.220.220",
						Port:    5353,
					},
				}
			}
		}
	case configure.AntipollutionNone:
		t.DNS.Servers = []interface{}{"119.29.29.29", "114.114.114.114"} //防止DNS劫持，使用DNSPod作为主DNS
	}
	if setting.AntiPollution != configure.AntipollutionNone {
		//统计DoH服务器信息
		dohIPs = make([]string, 0)
		dohDomains = make([]string, 0)
		for _, u := range t.DNS.Servers {
			switch u := u.(type) {
			case string:
				if !strings.HasPrefix(strings.ToLower(u), "https://") &&
					!strings.HasPrefix(strings.ToLower(u), "https+local://") {
					break
				}
				uu, e := url.Parse(u)
				if e != nil {
					continue
				}
				//如果是非IP则解析为IP
				if net.ParseIP(uu.Hostname()) == nil {
					dohDomains = append(dohDomains, uu.Hostname())
					addrs, e := net.LookupHost(uu.Hostname())
					if e != nil {
						log.Println("net.LookupHost:", e)
						continue
					}
					dohIPs = append(dohIPs, addrs...)
				} else {
					dohIPs = append(dohIPs, uu.Hostname())
				}
			}
		}

		ds := DnsServer{
			Address: "119.29.29.29",
			Port:    53,
			Domains: []string{
				"geosite:cn",          // 国内白名单走DNSPod
				"domain:ntp.org",      // NTP 服务器
				"domain:dogedoge.com", // mzz2017爱用的多吉
				"full:v2raya.mzz.pub", // v2rayA demo
				"full:v.mzz.pub",      // v2rayA demo
			},
		}
		if len(dohDomains) > 0 {
			ds.Domains = append(ds.Domains, dohDomains...)
		}
		if net.ParseIP(v.Add) == nil {
			//如果节点地址不是IP而是域名，将其二级域名加入白名单
			group := strings.Split(v.Add, ".")
			if len(group) >= 2 {
				domain := strings.Join(group[len(group)-2:], ".")
				ds.Domains = append(ds.Domains, "domain:"+domain)
			}
		}
		t.DNS.Servers = append(t.DNS.Servers,
			ds,
		)
	}
	if t.DNS != nil {
		//修改hosts
		t.DNS.Hosts = getHosts()
	}
	return
}

func (t *Template) SetDNSRouting(v vmessInfo.VmessInfo, dohIPs, dohHosts []string, setting *configure.Setting, supportUDP bool) (serverIPs []string, serverDomain string) {
	dohRouting := make([]RoutingRule, 0)
	if len(dohIPs) > 0 {
		hosts := make([]string, len(dohHosts))
		for i := range dohHosts {
			hosts[i] = "full:" + dohHosts[i]
		}
		dohRouting = append(dohRouting, RoutingRule{
			Type:        "field",
			OutboundTag: "direct", //如果配置了dns转发，此处将被改成proxy
			IP:          dohIPs,
			Port:        "443",
		}, RoutingRule{
			Type:        "field",
			OutboundTag: "direct", //如果配置了dns转发，此处将被改成proxy
			Domain:      hosts,
			Port:        "443",
		})
	}
	if setting.AntiPollution == configure.DnsForward {
		for i := range dohRouting {
			dohRouting[i].OutboundTag = "proxy"
		}
	}
	if supportUDP {
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{ // 国外DNS服务器地址走代理，以防污染
				Type:        "field",
				OutboundTag: "proxy",
				IP:          []string{"8.8.8.8", "1.1.1.1"},
				Port:        "53",
			},
		)
	}
	t.Routing.Rules = append(t.Routing.Rules,
		RoutingRule{ // 国内DNS服务器直连，以分流
			Type:        "field",
			OutboundTag: "direct",
			IP:          []string{"119.29.29.29", "114.114.114.114"},
			Port:        "53",
		},
		RoutingRule{ // 劫持 53 端口流量，使用 V2Ray 的 DNS
			Type:        "field",
			Port:        "53",
			OutboundTag: "dns-out",
		},
		RoutingRule{ // DNSPoison
			Type:        "field",
			IP:          []string{"240.0.0.0/4"},
			OutboundTag: "proxy",
		},
	)
	if setting.AntiPollution != configure.AntipollutionNone {
		t.Routing.Rules = append(t.Routing.Rules, RoutingRule{ // 非标准端口暂时安全，直连
			Type:        "field",
			OutboundTag: "direct",
			IP:          []string{"208.67.222.222", "208.67.220.220"},
			Port:        "5353",
		})
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
	t.Routing.Rules = append(t.Routing.Rules, dohRouting...)
	t.Routing.Rules = append(t.Routing.Rules,
		RoutingRule{ // 直连 123 端口 UDP 流量（NTP 协议）
			Type:        "field",
			OutboundTag: "direct",
			Network:     "udp",
			Port:        "123",
		},
		RoutingRule{ // BT流量直连
			Type:        "field",
			OutboundTag: "direct",
			Protocol:    []string{"bittorrent"},
		},
	)
	serverIPs = []string{v.Add}
	if net.ParseIP(v.Add) == nil {
		//如果不是IP，而是域名，将其加入白名单
		t.Routing.Rules = append([]RoutingRule{{
			Type:        "field",
			OutboundTag: "direct",
			Domain:      []string{"full:" + v.Add},
		}}, t.Routing.Rules...
		)
		serverDomain = v.Add
		//解析IP
		ips, e := net.LookupHost(v.Add)
		if e != nil {
			log.Println("net.LookupHost:", e)
		}
		serverIPs = ips
	}
	//将节点IP加入白名单
	if len(serverIPs) > 0 {
		t.Routing.Rules = append([]RoutingRule{{
			Type:        "field",
			OutboundTag: "direct",
			IP:          serverIPs,
		}}, t.Routing.Rules...
		)
	}
	return
}

func (t *Template) SetPacRouting(setting *configure.Setting) {
	switch setting.PacMode {
	case configure.WhitelistMode:
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{ // 直连中国大陆主流网站域名
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"pac"},
				Domain:      []string{"geosite:cn"},
			},
			RoutingRule{ // 直连中国大陆主流网站 ip 和 私有 ip
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"pac"},
				IP:          []string{"geoip:private", "geoip:cn"},
			},
		)
	case configure.GfwlistMode:
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  []string{"pac"},
				Domain:      []string{"ext:LoyalsoldierSite.dat:geolocation-!cn"},
			},
			RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"pac"},
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
					InboundTag:  []string{"pac"},
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
				InboundTag:  []string{"pac"},
			})
		case configure.DefaultBlockMode:
			//如果默认拦截，则需要加上下述规则
			t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
				Type:        "field",
				OutboundTag: "block",
				InboundTag:  []string{"pac"},
			})
		}
	case configure.RoutingAMode:
		parseRoutingA(t, []string{"pac"})
	}
}
func parseRoutingA(t *Template, inboundTags []string) {
	ra := configure.GetRoutingA()
	rules, err := routingA.Parse(ra)
	if err != nil {
		log.Println(err)
		return
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
			case "outbound":
				switch o := rule.Value.(type) {
				case routingA.Outbound:
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
									u.Level, _ = strconv.Atoi(levels[i])
								}
								server.Users = append(server.Users, u)
							}
						}
						t.Outbounds = append(t.Outbounds, Outbound{
							Tag:      o.Name,
							Protocol: o.Value.Name,
							Settings: &Settings{
								Servers: []Server{
									server,
								},
							},
						})
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
				InboundTag:  inboundTags,
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
					//rr.Domain = append(rr.Domain, f.Params...)
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
				}
			}
			t.Routing.Rules = append(t.Routing.Rules, rr)
		}
	}
	t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
		Type:        "field",
		OutboundTag: defaultOutbound,
		InboundTag:  []string{"pac"},
	})
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
				if in == "pac" {
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
		t.Outbounds[i].StreamSettings.Sockopt.Mark = &mark
		//t.Outbounds[i].StreamSettings.Sockopt.Tos = &tos // Experimental in the future
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
		if global.SupportTproxy && !setting.EnhancedMode {
			tproxy = "tproxy"
		} else {
			tproxy = "redirect"
		}
		t.AppendDokodemo(&tproxy, 32345, "transparent")
	}
}

func NewTemplateFromVmessInfo(v vmessInfo.VmessInfo) (t Template, info *entity.ExtraInfo, err error) {
	setting := configure.GetSettingNotNil()
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return t, nil, newError("error occurs while reading template json, please check whether templateJson variable is correct json format")
	}
	// 其中Template是基础配置，替换掉t即可
	t = tmplJson.Template
	// 调试模式
	if global.IsDebug() {
		t.Log.Loglevel = "debug"
	}
	// 解析Outbound
	o, err := ResolveOutbound(&v, "proxy", &global.GetEnvironmentConfig().PluginListenPort)
	if err != nil {
		return t, nil, err
	}
	t.Outbounds[0] = o
	var supportUDP = true
	switch o.Protocol {
	case "vmess":
		//是否在设置了里开启了mux
		t.Outbounds[0].Mux = &Mux{
			Enabled:     setting.MuxOn == configure.Yes,
			Concurrency: setting.Mux,
		}
	default:
		supportUDP = false
	}
	//根据配置修改端口
	t.SetInbound(setting)
	//设置DNS
	dohIPs, dohHosts := t.SetDNS(v, supportUDP, setting)
	//再修改outbounds
	t.AppendDNSOutbound()
	//最后是routing
	serverIPs, serverDomain := t.SetDNSRouting(v, dohIPs, dohHosts, setting, supportUDP)
	//添加hosts
	if len(serverDomain) > 0 && len(serverIPs) > 0 {
		t.DNS.Hosts[serverDomain] = serverIPs[0]
	}
	//PAC端口规则
	t.SetPacRouting(setting)
	//根据是否使用全局代理修改路由
	if setting.Transparent != configure.TransparentClose {
		t.SetTransparentRouting(setting)
	}
	//置outboundSockopt
	t.SetOutboundSockopt(supportUDP, setting)

	return t, &entity.ExtraInfo{
		DohIps:       dohIPs,
		DohDomains:   dohHosts,
		ServerIps:    serverIPs,
		ServerDomain: serverDomain,
	}, nil
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}

func WriteV2rayConfig(content []byte) (err error) {
	err = ioutil.WriteFile(asset.GetConfigPath(), content, os.FileMode(0600))
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
func (t *Template) AddMappingOutbound(v vmessInfo.VmessInfo, inboundPort string, udpSupport bool, pluginPort int, protocol string) (err error) {
	o, err := ResolveOutbound(&v, "outbound"+inboundPort, &pluginPort)
	if err != nil {
		return
	}
	var mark = 0xff
	//var tos = 184
	if o.StreamSettings == nil {
		o.StreamSettings = new(StreamSettings)
	}
	if o.StreamSettings.Sockopt == nil {
		o.StreamSettings.Sockopt = new(Sockopt)
	}
	o.StreamSettings.Sockopt.Mark = &mark
	//o.StreamSettings.Sockopt.Tos = &tos
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
	b, err := ioutil.ReadFile("/etc/hosts")
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
