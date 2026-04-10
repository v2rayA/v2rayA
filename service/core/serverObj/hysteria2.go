package serverObj

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("hysteria2", func(link string) (ServerObj, error) {
		return ParseHysteria2URL(link)
	})
	FromLinkRegister("hy2", func(link string) (ServerObj, error) {
		return ParseHysteria2URL(link)
	})
	EmptyRegister("hysteria2", func() (ServerObj, error) {
		return new(Hysteria2), nil
	})
	EmptyRegister("hy2", func() (ServerObj, error) {
		return new(Hysteria2), nil
	})
}

type Hysteria2 struct {
	coreObj.ServerCommon
	Password      string `json:"password"`
	Insecure      bool   `json:"insecure"`
	Sni           string `json:"sni"`
	Obfs          string `json:"obfs"`
	ObfsPassword  string `json:"obfs-password"`
	UpMbps        string `json:"upmbps"`
	DownMbps      string `json:"downmbps"`
	Congestion    string `json:"congestion"`
	Link          string `json:"link"`
}

func (h *Hysteria2) ExportToURL() string {
	return h.Link
}

func (h *Hysteria2) ProtoToShow() string {
	return "Hysteria2"
}

func (h *Hysteria2) GetProtocol() string {
	return "hysteria2"
}

func (h *Hysteria2) GetHostname() string {
	return h.Hostname
}

func (h *Hysteria2) GetPort() int {
	return h.Port
}

func (h *Hysteria2) GetName() string {
	return h.Hostname
}

func (h *Hysteria2) SetName(name string) {
	h.Hostname = name
}

func (h *Hysteria2) Configuration(info PriorInfo) (Configuration, error) {
	// Re-parse the link to get the latest credentials and params
	parsed, err := ParseHysteria2URL(h.Link)
	if err == nil {
		h.Password = parsed.Password
		h.Insecure = parsed.Insecure
		h.Sni = parsed.Sni
		h.Obfs = parsed.Obfs
		h.ObfsPassword = parsed.ObfsPassword
		h.UpMbps = parsed.UpMbps
		h.DownMbps = parsed.DownMbps
		h.Congestion = parsed.Congestion
		h.Address = parsed.Address
		h.Port = parsed.Port
	}
	
	coreOutbound := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "hysteria2",
		Settings: coreObj.Settings{
			Address: h.Address,
			Port:    h.Port,
		},
		StreamSettings: &coreObj.StreamSettings{
			Network: "udp",
			HysteriaSettings: &coreObj.HysteriaSettings{
				Auth: h.Password,
			},
			Security: "tls",
			TLSSettings: &coreObj.TLSSettings{
				ServerName: h.Sni,
				AllowInsecure: h.Insecure,
			},
			Sockopt: &coreObj.Sockopt{
				Mark: func() *int { i := 128; return &i }(),
			},
		},
	}
	
	if h.Obfs != "" && h.Obfs != "none" {
		coreOutbound.StreamSettings.HysteriaSettings.Obfs = &coreObj.MaskItem{
			Type:     h.Obfs,
			Password: h.ObfsPassword,
		}
	}
	
	if h.UpMbps != "" || h.DownMbps != "" {
		up, _ := strconv.Atoi(strings.TrimSuffix(strings.ToLower(h.UpMbps), " mbps"))
		down, _ := strconv.Atoi(strings.TrimSuffix(strings.ToLower(h.DownMbps), " mbps"))
		if coreOutbound.StreamSettings.HysteriaSettings.QuicParams == nil {
			coreOutbound.StreamSettings.HysteriaSettings.QuicParams = &coreObj.QuicParams{}
		}
		coreOutbound.StreamSettings.HysteriaSettings.QuicParams.Up = up
		coreOutbound.StreamSettings.HysteriaSettings.QuicParams.Down = down
	}
	
	if h.Congestion != "" {
		if coreOutbound.StreamSettings.HysteriaSettings.QuicParams == nil {
			coreOutbound.StreamSettings.HysteriaSettings.QuicParams = &coreObj.QuicParams{}
		}
		coreOutbound.StreamSettings.HysteriaSettings.QuicParams.Congestion = h.Congestion
	}
	
	return Configuration{
		CoreOutbound: coreOutbound,
		UDPSupport:   true,
	}, nil
}

func (h *Hysteria2) NeedPluginPort() bool {
	return false
}

func (h *Hysteria2) GetLink() string {
	return h.Link
}

func ParseHysteria2URL(link string) (*Hysteria2, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "hysteria2" && u.Scheme != "hy2" {
		return nil, fmt.Errorf("invalid scheme")
	}
	host, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)
	h := &Hysteria2{
		ServerCommon: coreObj.ServerCommon{
			Address: host,
			Port:    port,
		},
		Password:     u.User.Username(),
		Insecure:     u.Query().Get("insecure") == "1",
		Sni:          u.Query().Get("sni"),
		Obfs:         u.Query().Get("obfs"),
		ObfsPassword: u.Query().Get("obfs-password"),
		UpMbps:       u.Query().Get("upmbps"),
		DownMbps:     u.Query().Get("downmbps"),
		Congestion:   u.Query().Get("congestion"),
		Link:         link,
	}
	if h.Password == "" && u.User != nil {
		if p, ok := u.User.Password(); ok {
			h.Password = p
		}
	}
	h.Hostname = u.Fragment
	return h, nil
}
