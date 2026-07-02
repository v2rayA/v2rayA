package serverObj

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/v2rayA/v2rayA/kernel/coreObj"
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
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
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
	return &Hysteria2{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "hysteria2",
		Link:     link,
	}, nil
}

// hysteria2Settings holds the settings serialized into the v2raya-core xray config.
type hysteria2Settings struct {
	Address                string `json:"address"` // "host:port" format — maps to ClientConfig.address proto field
	Password               string `json:"password"`
	SNI                    string `json:"sni,omitempty"`
	AllowInsecure          bool   `json:"allow_insecure,omitempty"`
	PinnedCertchainSha256  string `json:"pinned_certchain_sha256,omitempty"`
	Obfs                   string `json:"obfs,omitempty"`
	ObfsPassword           string `json:"obfs_password,omitempty"`
	UpMbps                 int    `json:"up_mbps,omitempty"`
	DownMbps               int    `json:"down_mbps,omitempty"`
}

func (s *Hysteria2) Configuration(info PriorInfo) (c Configuration, err error) {
	u, err := url.Parse(s.Link)
	if err != nil {
		return c, fmt.Errorf("hysteria2: parse link: %w", err)
	}

	password := ""
	if u.User != nil {
		password = u.User.Username()
	}
	q := u.Query()
	sni := q.Get("sni")
	obfs := q.Get("obfs")
	if obfs == "" {
		obfs = "none"
	}
	obfsPassword := q.Get("obfs-password")

	allowInsecure := false
	if q.Get("allowInsecure") == "1" || q.Get("insecure") == "1" {
		allowInsecure = true
	}
	upMbps, _ := strconv.Atoi(q.Get("upmbps"))
	if upMbps < 0 {
		upMbps = 0
	}
	downMbps, _ := strconv.Atoi(q.Get("downmbps"))
	if downMbps < 0 {
		downMbps = 0
	}

	settingsJSON, err := json.Marshal(hysteria2Settings{
		Address:                net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		Password:               password,
		SNI:                    sni,
		AllowInsecure:          allowInsecure,
		PinnedCertchainSha256:  q.Get("pinnedPeerCertSha256"),
		Obfs:                   obfs,
		ObfsPassword:           obfsPassword,
		UpMbps:                 upMbps,
		DownMbps:               downMbps,
	})
	if err != nil {
		return c, fmt.Errorf("hysteria2: marshal settings: %w", err)
	}

	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:      info.Tag,
			Protocol: "hysteria2",
			Settings: coreObj.Settings{Inlined: settingsJSON},
		},
		UDPSupport: true,
	}, nil
}

func (s *Hysteria2) ExportToURL() string {
	return s.Link
}

func (s *Hysteria2) NeedPluginPort() bool {
	return false
}

func (s *Hysteria2) ProtoToShow() string {
	return "hysteria2"
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
