package serverObj

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/v2rayA/v2rayA/core/coreObj"
)

func init() {
	FromLinkRegister("hysteria2", NewHysteria2)
	FromLinkRegister("hy2", NewHysteria2)
	EmptyRegister("hysteria2", func() (ServerObj, error) {
		return new(Hysteria2), nil
	})
	EmptyRegister("hy2", func() (ServerObj, error) {
		return new(Hysteria2), nil
	})
}

type Hysteria2 struct {
	Name          string `json:"name"`
	Server        string `json:"server"`
	Port          int    `json:"port"`
	Auth          string `json:"auth"`
	Obfs          string `json:"obfs"`
	ObfsPassword  string `json:"obfsPassword"`
	SNI           string `json:"sni"`
	AllowInsecure bool   `json:"allowInsecure"`
	Protocol      string `json:"protocol"`
	Link          string `json:"link"`
}

func NewHysteria2(link string) (ServerObj, error) {
	return ParseHysteria2URL(link)
}

func ParseHysteria2URL(link string) (data *Hysteria2, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	portStr := u.Port()
	if portStr == "" {
		portStr = "443"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	auth := u.User.Username()
	if p, ok := u.User.Password(); ok {
		auth += ":" + p
	}
	q := u.Query()
	return &Hysteria2{
		Name:          u.Fragment,
		Server:        u.Hostname(),
		Port:          port,
		Auth:          auth,
		Obfs:          q.Get("obfs"),
		ObfsPassword:  q.Get("obfs-password"),
		SNI:           q.Get("sni"),
		AllowInsecure: q.Get("insecure") == "1" || q.Get("allowInsecure") == "true",
		Protocol:      "hysteria2",
		Link:          link,
	}, nil
}

func (s *Hysteria2) Configuration(info PriorInfo) (c Configuration, err error) {
	core := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "hysteria",
	}
	core.Settings.Version = 2
	core.Settings.Address = s.Server
	core.Settings.Port = s.Port

	core.StreamSettings = &coreObj.StreamSettings{
		Network:  "hysteria",
		Security: "tls",
		TLSSettings: &coreObj.TLSSettings{
			ServerName:    s.SNI,
			AllowInsecure: s.AllowInsecure,
			Alpn:          []string{"h3"},
		},
		HysteriaSettings: &coreObj.HysteriaSettings{
			Version: 2,
			Auth:    s.Auth,
		},
	}
	if s.Obfs == "salamander" {
		core.StreamSettings.FinalMask = &coreObj.FinalMask{
			UDP: []coreObj.MaskItem{
				{
					Type: "salamander",
					Settings: map[string]string{
						"password": s.ObfsPassword,
					},
				},
			},
		}
	}
	return Configuration{
		CoreOutbound: core,
		PluginChain:  "",
		UDPSupport:   true,
	}, nil
}

func (s *Hysteria2) ExportToURL() string {
	return s.Link
}

func (s *Hysteria2) NeedPluginPort() bool {
	return false
}

func (s *Hysteria2) ProtoToShow() string {
	return fmt.Sprintf("hysteria2")
}

func (s *Hysteria2) GetProtocol() string {
	return s.Protocol
}

func (s *Hysteria2) GetHostname() string {
	return s.Server
}

func (s *Hysteria2) GetPort() int {
	return s.Port
}

func (s *Hysteria2) GetName() string {
	return s.Name
}

func (s *Hysteria2) SetName(name string) {
	s.Name = name
}
