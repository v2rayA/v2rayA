package v2rayTmpl

import (
	"V2RayA/global"
	"V2RayA/model/transparentProxy"
	"V2RayA/model/vmessInfo"
	"V2RayA/persistence/configure"
	"errors"
	"github.com/json-iterator/go"
	"net"
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
	Port        int      `json:"port,omitempty"`
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
	Method   string `json:"method"`
	Ota      bool   `json:"ota"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}
type Settings struct {
	Vnext          interface{} `json:"vnext"`
	Servers        interface{} `json:"servers"`
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
	Settings       Settings        `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	Mux            *Mux            `json:"mux,omitempty"`
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
	Servers []interface{} `json:"servers"`
}
type DnsServer struct {
	Address string   `json:"address"`
	Port    int      `json:"port"`
	Domains []string `json:"domains"`
}

func NewTemplate() (tmpl *Template) {
	tmpl = new(Template)
	return
}

/*
根据传入的 VmessInfo 填充模板
当协议是ss时，v.Net对应Method，v.ID对应Password
*/
func (t *Template) FillWithVmessInfo(v vmessInfo.VmessInfo) error {
	var tmplJson TmplJson
	// 读入模板json
	raw := []byte(TemplateJson)
	err := jsoniter.Unmarshal(raw, &tmplJson)
	if err != nil {
		return errors.New("读入模板json出错，请检查templateJson变量是否是正确的json格式")
	}
	// 其中Template是基础配置，替换掉*t即可
	*t = tmplJson.Template
	// 默认协议vmess
	if v.Protocol == "" {
		v.Protocol = "vmess"
	}
	// 根据vmessInfo修改json配置
	t.Outbounds[0].Protocol = v.Protocol
	port, _ := strconv.Atoi(v.Port)
	aid, _ := strconv.Atoi(v.Aid)
	setting := configure.GetSetting()
	switch strings.ToLower(v.Protocol) {
	case "vmess":
		t.Outbounds[0].Settings.Vnext = []Vnext{
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
		t.Outbounds[0].StreamSettings = &tmplJson.StreamSettings
		t.Outbounds[0].StreamSettings.Network = v.Net
		// 根据传输协议(network)修改streamSettings
		switch strings.ToLower(v.Net) {
		case "ws":
			tmplJson.WsSettings.Headers.Host = v.Host
			tmplJson.WsSettings.Path = v.Path
			t.Outbounds[0].StreamSettings.WsSettings = &tmplJson.WsSettings
		case "mkcp", "kcp":
			tmplJson.KcpSettings.Header.Type = v.Type
			t.Outbounds[0].StreamSettings.KcpSettings = &tmplJson.KcpSettings
		case "tcp":
			if strings.ToLower(v.Type) != "none" { //那就是http无疑了
				tmplJson.TCPSettings.Header.Request.Headers.Host = strings.Split(v.Host, ",")
				t.Outbounds[0].StreamSettings.TCPSettings = &tmplJson.TCPSettings
			}
		case "h2", "http":
			tmplJson.HttpSettings.Host = strings.Split(v.Host, ",")
			tmplJson.HttpSettings.Path = v.Path
			t.Outbounds[0].StreamSettings.HTTPSettings = &tmplJson.HttpSettings
		}
		if strings.ToLower(v.TLS) == "tls" {
			t.Outbounds[0].StreamSettings.Security = "tls"
			t.Outbounds[0].StreamSettings.TLSSettings = &tmplJson.TLSSettings
		}
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
	case "ss":
		t.Outbounds[0].Settings.Servers = []Server{
			{
				Address:  v.Add,
				Port:     port,
				Method:   v.Net,
				Password: v.ID,
				Ota:      false, //避免chacha20无法工作
			},
		}
	default:
		return errors.New("不支持的协议: " + v.Protocol)
	}

	//根据设置修改透明代理
	if global.ServiceControlMode != global.DockerMode && setting.Transparent != configure.TransparentClose {
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
			//如果不是IP，而是域名，也加入白名单
			ds.Domains = append(ds.Domains, v.Add)
		}
		t.DNS.Servers = []interface{}{
			"8.8.8.8",         // 非中中国大陆域名使用 Google 的 DNS
			"1.1.1.1",         // 非中中国大陆域名使用 Cloudflare 的 DNS(备用)
			"114.114.114.114", // 114 的 DNS (备用)
			ds,
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
			Settings:       &InboundSettings{Network: "tcp,udp", FollowRedirect: true},
			StreamSettings: StreamSettings{Sockopt: &Sockopt{Tproxy: &tproxy}},
			Tag:            "transparent",
		})
		//再修改outbounds
		mark := 0xff
		t.Outbounds = append(t.Outbounds, Outbound{
			Tag:      "dns-out",
			Protocol: "dns",
			StreamSettings: &StreamSettings{
				Sockopt: &Sockopt{Mark: &mark},
			},
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
				Port:        53,
				OutboundTag: "dns-out",
				Network:     "udp",
			},
			RoutingRule{ // 直连 123 端口 UDP 流量（NTP 协议）
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  []string{"transparent"},
				Network:     "udp",
				Port:        123,
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
		_ = transparentProxy.DeleteRules()
		err = transparentProxy.WriteRules()
		if err != nil {
			return err
		}
	} else {
		_ = transparentProxy.DeleteRules()
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
	return nil
}

func (t *Template) ToConfigBytes() []byte {
	b, _ := jsoniter.Marshal(t)
	return b
}
