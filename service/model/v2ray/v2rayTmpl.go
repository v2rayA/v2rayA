package v2ray

import (
	"V2RayA/global"
	"V2RayA/model/iptables"
	"V2RayA/model/vmessInfo"
	"V2RayA/persistence/configure"
	"errors"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net"
	"os"
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
	Address  string `json:"address"`
	Method   string `json:"method,omitempty"`
	Ota      bool   `json:"ota,omitempty"`
	Password string `json:"password,omitempty"`
	Port     int    `json:"port"`
}
type Settings struct {
	Vnext          interface{} `json:"vnext,omitempty"`
	Servers        interface{} `json:"servers,omitempty"`
	DomainStrategy string      `json:"domainStrategy,omitempty"`
}
type TLSSettings struct {
	AllowInsecure bool        `json:"allowInsecure"`
	ServerName    interface{} `json:"serverName"`
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
	Mux            *Mux            `json:"mux,omitempty"`
	Network        string          `json:"network,omitempty"`
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
type DNS struct {
	Hosts   map[string]string `json:"hosts"`
	Servers []interface{}     `json:"servers"`
}
type DnsServer struct {
	Address string   `json:"address"`
	Port    int      `json:"port"`
	Domains []string `json:"domains"`
}

func NewTemplate() (tmpl Template) {
	return
}

/*
根据传入的 VmessInfo 填充模板
当协议是ss时，v.Net对应Method，v.ID对应Password
函数会规格化传入的v
*/

func ResolveOutbound(v *vmessInfo.VmessInfo, tag string, ssrLocalPortIfNeed int) (o Outbound, err error) {
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return o, errors.New("读入模板json出错，请检查templateJson变量是否是正确的json格式")
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
		}
	case "shadowsocks", "shadowsocksr":
		v.Net = strings.ToLower(v.Net)
		switch v.Net {
		case "aes-128-cfb", "aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "aes-128-ofb", "aes-192-ofb", "aes-256-ofb", "des-cfb", "bf-cfb", "cast5-cfb", "rc4-md5", "chacha20", "chacha20-ietf", "salsa20", "camellia-128-cfb", "camellia-192-cfb", "camellia-256-cfb", "idea-cfb", "rc2-cfb", "seed-cfb":
		default:
			return o, errors.New("不支持的shadowsocks加密方法: " + v.Net)
		}
		if len(strings.TrimSpace(v.Type)) <= 0 {
			v.Type = "origin"
		}
		switch v.Type {
		case "origin", "verify_sha1", "auth_sha1_v4", "auth_aes128_md5", "auth_aes128_sha1":
		default:
			return o, errors.New("不支持的shadowsocksr协议: " + v.Type)
		}
		if len(strings.TrimSpace(v.TLS)) <= 0 {
			v.TLS = "plain"
		}
		switch v.TLS {
		case "plain", "http_simple", "http_post", "random_head", "tls1.2_ticket_auth":
		default:
			return o, errors.New("不支持的shadowsocksr混淆方法: " + v.TLS)
		}
		o.Protocol = "socks"
		o.Settings.Servers = []Server{
			{
				Address: "127.0.0.1",
				Port:    ssrLocalPortIfNeed,
			},
		}
	default:
		return o, errors.New("不支持的协议: " + v.Protocol)
	}
	o.Tag = tag
	return
}

func NewTemplateFromVmessInfo(v vmessInfo.VmessInfo) (t Template, err error) {
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err = jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return t, errors.New("读入模板json出错，请检查templateJson变量是否是正确的json格式")
	}
	// 其中Template是基础配置，替换掉*t即可
	t = tmplJson.Template
	o, err := ResolveOutbound(&v, "proxy", global.GetEnvironmentConfig().SSRListenPort)
	if err != nil {
		return t, err
	}
	t.Outbounds[0] = o
	setting := configure.GetSettingNotNil()
	switch o.Protocol {
	case "vmess":
		//是否在设置里开启了TCPFastOpen
		if setting.TcpFastOpen != configure.Default {
			t.Outbounds[0].StreamSettings.Sockopt = new(Sockopt)
			tmp := setting.TcpFastOpen == configure.Yes
			t.Outbounds[0].StreamSettings.Sockopt.TCPFastOpen = &tmp
		}
		//是否在设置了里开启了mux
		t.Outbounds[0].Mux = &Mux{
			Enabled:     setting.MuxOn == configure.Yes,
			Concurrency: setting.Mux,
		}
	}

	//根据配置修改端口
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
	//根据设置修改透明代理
	if setting.Transparent != configure.TransparentClose && CheckTProxySupported() == nil {
		//先修改DNS设置
		t.DNS = new(DNS)
		ds := DnsServer{
			Address: "119.29.29.29",
			Port:    53,
			Domains: []string{
				"geosite:cn",     // 国内白名单走DNSPod
				"ntp.org",        // NTP 服务器
				"v2raya.mzz.pub", // V2RayA demo
			},
		}
		if net.ParseIP(v.Add) == nil {
			//如果不是IP，而是域名，将其二级域名加入白名单
			group := strings.Split(v.Add, ".")
			if len(group) >= 2 {
				domain := strings.Join(group[len(group)-2:], ".")
				ds.Domains = append(ds.Domains, domain)
			}
		}
		t.DNS.Servers = []interface{}{
			ds,
			"8.8.8.8", // 非中中国大陆域名使用 Google 的 DNS
			"1.1.1.1",
			"127.0.0.1",
		}
		//再修改inbounds
		tproxy := "tproxy"
		t.Inbounds = append(t.Inbounds, Inbound{
			Listen:   "0.0.0.0",
			Port:     12345,
			Protocol: "dokodemo-door",
			Sniffing: Sniffing{
				Enabled:      true,
				DestOverride: []string{"http", "tls"},
			},
			Settings:       &InboundSettings{FollowRedirect: true},
			StreamSettings: StreamSettings{Sockopt: &Sockopt{Tproxy: &tproxy}},
			Tag:            "transparent",
		})
		//再修改outbounds
		mark := 0xff
		t.Outbounds = append(t.Outbounds, Outbound{
			Tag:      "dns-out",
			Protocol: "dns",
		})
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
			if t.Outbounds[i].Protocol == "freedom" {
				t.Outbounds[i].Settings.DomainStrategy = "UseIP"
			}
			t.Outbounds[i].StreamSettings.Sockopt.Mark = &mark
		}
		//最后是routing
		t.Routing.Rules = append(t.Routing.Rules,
			RoutingRule{ // 劫持 53 端口 UDP 流量，使用 V2Ray 的 DNS
				Type:        "field",
				InboundTag:  []string{"transparent"},
				Port:        "53",
				OutboundTag: "dns-out",
				Network:     "udp",
			},
			RoutingRule{ // 直连 123 端口 UDP 流量（NTP 协议）
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"transparent"},
				Network:     "udp",
				Port:        "123",
			},
			RoutingRule{ // 设置 DNS 配置中的国内 DNS 服务器地址直连，以达到 DNS 分流目的
				Type:        "field",
				OutboundTag: "direct",
				IP:          []string{"119.29.29.29", "114.114.114.114", "223.5.5.5"},
			},
			RoutingRule{ // 设置 DNS 配置中的国外 DNS 服务器地址走代理，以达到 DNS 分流目的
				Type:        "field",
				OutboundTag: "proxy",
				IP:          []string{"8.8.8.8", "1.1.1.1"},
			},
			RoutingRule{ // BT 流量直连
				Type:        "field",
				OutboundTag: "direct",
				Protocol:    []string{"bittorrent"},
			},
		)
		if net.ParseIP(v.Add) == nil {
			//如果不是IP，而是域名，将其二级域名加入白名单
			group := strings.Split(v.Add, ".")
			if len(group) >= 2 {
				domain := strings.Join(group[len(group)-2:], ".")
				t.Routing.Rules = append(t.Routing.Rules,
					RoutingRule{
						Type:        "field",
						OutboundTag: "direct",
						Domain:      []string{"domain:" + domain},
					})
			}

		}
		switch setting.Transparent {
		case configure.TransparentProxy:
		case configure.TransparentWhitelist:
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{ // 直连中国大陆主流网站 ip 和 私有 ip
					Type:        "field",
					OutboundTag: "direct",
					IP:          []string{"geoip:private", "geoip:cn"},
				},
				RoutingRule{ // 直连中国大陆主流网站域名
					Type:        "field",
					OutboundTag: "direct",
					Domain:      []string{"geosite:cn"},
				},
			)
		case configure.TransparentGfwlist:
			t.Routing.Rules = append(t.Routing.Rules,
				RoutingRule{
					Type:        "field",
					OutboundTag: "proxy",
					Domain:      []string{"ext:h2y.dat:gfw"},
				},
				RoutingRule{
					Type:        "field",
					OutboundTag: "direct",
					Network:     "tcp,udp",
				},
			)
		}
	} else {
		_ = iptables.DeleteRules()
		// 不是全局模式，根据设置修改路由部分的PAC规则
		switch setting.PacMode {
		case configure.WhitelistMode:
			t.Routing.Rules = append(t.Routing.Rules, tmplJson.Whitelist...)
		case configure.GfwlistMode:
			t.Routing.Rules = append(t.Routing.Rules, tmplJson.Gfwlist...)
		case configure.CustomMode:
			for _, v := range setting.CustomPac.RoutingRules {
				rule := RoutingRule{
					Type:        "field",
					OutboundTag: string(v.RuleType),
					InboundTag:  []string{"pac"},
				}
				for i := range v.Tags {
					v.Tags[i] = "ext:custom.dat:" + v.Tags[i]
				}
				switch v.MatchType {
				case configure.DomainMatchRule:
					rule.Domain = v.Tags
				case configure.IpMatchRule:
					rule.IP = v.Tags
				}
				t.Routing.Rules = append(t.Routing.Rules, rule)
			}
			//如果默认直连，规则内的才走代理，则需要加上下述规则
			if setting.CustomPac.DefaultProxyMode == "direct" {
				t.Routing.Rules = append(t.Routing.Rules, RoutingRule{
					Type:        "field",
					OutboundTag: "direct",
					InboundTag:  []string{"pac"},
					Network:     "tcp,udp",
				})
			}
		}
	}
	return t, nil
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}

func WriteV2rayConfig(content []byte) (err error) {
	err = ioutil.WriteFile(GetConfigPath(), content, os.ModeAppend)
	if err != nil {
		return errors.New("WriteV2rayConfig: " + err.Error())
	}
	return
}

func NewTemplateFromConfig() (t Template, err error) {
	b, err := GetConfigBytes()
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(b, &t)
	return
}
func (t *Template) AddMappingOutbound(v vmessInfo.VmessInfo, inboundPort string, udpSupport bool, ssrLocalPortIfNeed int) (err error) {
	o, err := ResolveOutbound(&v, "outbound"+inboundPort, ssrLocalPortIfNeed)
	if err != nil {
		return
	}
	var mark = 0xff
	if o.StreamSettings == nil {
		o.StreamSettings = new(StreamSettings)
	}
	if o.StreamSettings.Sockopt == nil {
		o.StreamSettings.Sockopt = new(Sockopt)
	}
	o.StreamSettings.Sockopt.Mark = &mark
	t.Outbounds = append(t.Outbounds, o)
	iPort, err := strconv.Atoi(inboundPort)
	if err != nil || iPort <= 0 {
		return errors.New("inboundPort必须为string类型的正数")
	}
	t.Inbounds = append(t.Inbounds, Inbound{
		Port:     iPort,
		Protocol: "socks",
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
	//将routingRule插入最前
	t.Routing.Rules = append(make([]RoutingRule, 1), t.Routing.Rules...)
	t.Routing.Rules[0] = RoutingRule{
		Type:        "field",
		OutboundTag: "outbound" + inboundPort,
		InboundTag:  []string{"inbound" + inboundPort},
	}
	return
}
