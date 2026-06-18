package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"

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
	Address              string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port                 int    `json:"port"`
	Auth                 string `json:"auth"`
	Obfs                 string `json:"obfs"`
	ObfsPassword         string `json:"obfsPassword"`
	Sni                  string `json:"sni"`
	Up                   string `json:"up"`
	Down                 string `json:"down"`
	Congestion           string `json:"congestion"`
	FinalMask            bool   `json:"finalMask"`
	PinnedPeerCertSha256 string `json:"pinnedPeerCertSha256"`
	Name                 string `json:"name"`
	Protocol             string `json:"protocol"`
	Link                 string `json:"link"`
}

func NewHysteria2(link string) (ServerObj, error) {
	return ParseHysteria2URL(link)
}

func ParseHysteria2URL(link string) (data *Hysteria2, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		portStr = "443"
	}
	port, _ := strconv.Atoi(portStr)
	auth := u.User.Username()
	if p, ok := u.User.Password(); ok {
		auth += ":" + p
	}
	q := u.Query()
	pinSHA256 := q.Get("pinSHA256")
	if pinSHA256 == "" {
		pinSHA256 = q.Get("pinnedPeerCertSha256")
	}
	return &Hysteria2{
		Name:                 u.Fragment,
		Address:              host,
		Port:                 port,
		Auth:                 auth,
		Obfs:                 q.Get("obfs"),
		ObfsPassword:         q.Get("obfs-password"),
		Sni:                  q.Get("sni"),
		Up:                   q.Get("upmbps"),
		Down:                 q.Get("downmbps"),
		Congestion:           q.Get("congestion"),
		FinalMask:            q.Get("finalmask") == "1",
		PinnedPeerCertSha256: pinSHA256,
		Protocol:             "hysteria2",
		Link:                 link,
	}, nil
}

func (s *Hysteria2) Configuration(info PriorInfo) (c Configuration, err error) {
	if !s.FinalMask {
		// Bridge mode for backward compatibility
		socks5 := url.URL{
			Scheme: "socks5",
			Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
		}
		chain := []string{socks5.String(), s.Link}
		return Configuration{
			CoreOutbound: info.PluginObj(),
			PluginChain:  strings.Join(chain, ","),
			UDPSupport:   true,
		}, nil
	}

	// Native Hysteria2 mode for latest Xray
	core := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "hysteria",
	}
	core.Settings.Version = 2
	core.Settings.Address = s.Address
	core.Settings.Port = s.Port

	core.StreamSettings = &coreObj.StreamSettings{
		Network:  "hysteria",
		Security: "tls",
		TLSSettings: &coreObj.TLSSettings{
			ServerName:           s.Sni,
			Alpn:                 []string{"h3"},
			PinnedPeerCertSha256: s.PinnedPeerCertSha256,
		},
		HysteriaSettings: &coreObj.HysteriaSettings{
			Version: 2,
			Auth:    s.Auth,
		},
	}
	if s.Obfs == "salamander" || s.Up != "" || s.Down != "" || s.Congestion != "" {
		core.StreamSettings.FinalMask = &coreObj.FinalMask{
			QuicParams: &coreObj.QuicParams{
				Up:         s.Up,
				Down:       s.Down,
				Congestion: s.Congestion,
			},
		}
		if s.Obfs == "salamander" {
			core.StreamSettings.FinalMask.UDP = []coreObj.MaskItem{
				{
					Type: "salamander",
					Settings: map[string]string{
						"password": s.ObfsPassword,
					},
				},
			}
		}
	}
	return Configuration{
		CoreOutbound: core,
		PluginChain:  "",
		UDPSupport:   true,
	}, nil
}

func (s *Hysteria2) ExportToURL() string {
	var u *url.URL
	if s.Link != "" {
		var err error
		u, err = url.Parse(s.Link)
		if err != nil {
			u = nil
		}
	}
	if u == nil {
		u = &url.URL{
			Scheme: "hysteria2",
		}
	}
	u.Host = net.JoinHostPort(s.Address, strconv.Itoa(s.Port))
	u.User = url.User(s.Auth)
	u.Fragment = s.Name
	q := u.Query()
	if s.Sni != "" {
		q.Set("sni", s.Sni)
	} else {
		q.Del("sni")
	}
	if s.Obfs != "none" && s.Obfs != "" {
		q.Set("obfs", s.Obfs)
		q.Set("obfs-password", s.ObfsPassword)
	} else {
		q.Del("obfs")
		q.Del("obfs-password")
	}
	if s.Up != "" {
		q.Set("upmbps", s.Up)
	} else {
		q.Del("upmbps")
	}
	if s.Down != "" {
		q.Set("downmbps", s.Down)
	} else {
		q.Del("downmbps")
	}
	if s.Congestion != "" {
		q.Set("congestion", s.Congestion)
	} else {
		q.Del("congestion")
	}
	if s.FinalMask {
		q.Set("finalmask", "1")
	} else {
		q.Del("finalmask")
	}
	if s.PinnedPeerCertSha256 != "" {
		q.Set("pinSHA256", s.PinnedPeerCertSha256)
	} else {
		q.Del("pinSHA256")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Hysteria2) NeedPluginPort() bool {
	return false
}

func (s *Hysteria2) ProtoToShow() string {
	return "Hysteria2"
}

func (s *Hysteria2) GetProtocol() string {
	return s.Protocol
}

func (s *Hysteria2) GetHostname() string {
	return s.Address
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
