package serverObj

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/core/coreObj"
)

func init() {
	FromLinkRegister("tuic", NewTuic)
	EmptyRegister("tuic", func() (ServerObj, error) {
		return new(Tuic), nil
	})
}

type Tuic struct {
	Name                 string `json:"name"`
	Server               string `json:"server"`
	Port                 int    `json:"port"`
	UUID                 string `json:"uuid"`
	Password             string `json:"password"`
	Sni                  string `json:"sni"`
	DisableSni           bool   `json:"disableSni"`
	Alpn                 string `json:"alpn"`
	CongestionControl    string `json:"congestionControl"`
	UdpRelayMode         string `json:"udpRelayMode"`
	PinnedPeerCertSha256 string `json:"pinnedPeerCertSha256,omitempty"`
	VerifyPeerCertByName string `json:"verifyPeerCertByName,omitempty"`
	Protocol             string `json:"protocol"`
}

func NewTuic(link string) (ServerObj, error) {
	return ParseTuicURL(link)
}

func ParseTuicURL(link string) (data *Tuic, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	// u.Query().Get("alpn") only returns the first value if it's encoded in a way that url.Values thinks it's multiple
	// But usually it's alpn=h3,h2. However, some clients might use alpn=h3&alpn=h2
	alpn := strings.Join(u.Query()["alpn"], ",")
	if alpn == "" {
		alpn = u.Query().Get("alpn")
	}
	alpn = strings.ReplaceAll(alpn, " ", "")
	if alpn == "" {
		alpn = "h3"
	}

	data = &Tuic{
		Name:                 u.Fragment,
		Server:               u.Hostname(),
		Port:                 port,
		UUID:                 u.User.Username(),
		Password:             u.User.String(),
		Sni:                  u.Query().Get("sni"),
		DisableSni:           u.Query().Get("disable_sni") == "true" || u.Query().Get("disable_sni") == "1",
		Alpn:                 alpn,
		CongestionControl:    u.Query().Get("congestion_control"),
		UdpRelayMode:         u.Query().Get("udp_relay_mode"),
		PinnedPeerCertSha256: u.Query().Get("pinnedPeerCertSha256"),
		VerifyPeerCertByName: u.Query().Get("verifyPeerCertByName"),
		Protocol:             "tuic",
	}
	if data.Password != "" && strings.Contains(data.Password, ":") {
		data.Password = strings.SplitN(data.Password, ":", 2)[1]
	} else if data.Password != "" {
		// handle case where password might be just the password part if user:pass was not used
		// but typically u.User.String() is user:pass or user
	}
	return data, nil
}

// tuicSettings mirrors hint/proxy/tuic.ClientConfig JSON tags for serialization.
type tuicSettings struct {
	Address              string   `json:"address"`
	UUID                 string   `json:"uuid"`
	Password             string   `json:"password"`
	Sni                  string   `json:"sni,omitempty"`
	CongestionControl    string   `json:"congestion_control,omitempty"`
	UdpRelayMode         string   `json:"udp_relay_mode,omitempty"`
	Alpn                 []string `json:"alpn,omitempty"`
	DisableSni           bool     `json:"disable_sni,omitempty"`
	PinnedPeerCertSha256 string   `json:"pinnedPeerCertSha256,omitempty"`
	VerifyPeerCertByName string   `json:"verifyPeerCertByName,omitempty"`
}

func (s *Tuic) Configuration(info PriorInfo) (c Configuration, err error) {
	alpn := strings.Split(strings.ReplaceAll(s.Alpn, " ", ""), ",")
	if len(alpn) == 0 || (len(alpn) == 1 && alpn[0] == "") {
		alpn = []string{"h3"}
	}

	settingsJSON, err := json.Marshal(tuicSettings{
		Address:              net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		UUID:                 s.UUID,
		Password:             s.Password,
		Sni:                  s.Sni,
		CongestionControl:    s.CongestionControl,
		UdpRelayMode:         s.UdpRelayMode,
		Alpn:                 alpn,
		DisableSni:           s.DisableSni,
		PinnedPeerCertSha256: s.PinnedPeerCertSha256,
		VerifyPeerCertByName: s.VerifyPeerCertByName,
	})
	if err != nil {
		return c, fmt.Errorf("tuic: marshal settings: %w", err)
	}

	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:      info.Tag,
			Protocol: "tuic",
			Settings: coreObj.Settings{Inlined: settingsJSON},
		},
		UDPSupport: true,
	}, nil
}

func (s *Tuic) ExportToURL() string {
	u := &url.URL{
		Scheme:   "tuic",
		User:     url.UserPassword(s.UUID, s.Password),
		Host:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		Fragment: s.Name,
	}
	query := u.Query()
	setValue(&query, "pinnedPeerCertSha256", s.PinnedPeerCertSha256)
	setValue(&query, "verifyPeerCertByName", s.VerifyPeerCertByName)
	if s.DisableSni {
		query.Set("disable_sni", "true")
	}
	setValue(&query, "sni", s.Sni)
	alpn := strings.ReplaceAll(s.Alpn, " ", "")
	if alpn == "" {
		alpn = "h3"
	}
	setValue(&query, "alpn", alpn)
	setValue(&query, "congestion_control", s.CongestionControl)
	setValue(&query, "udp_relay_mode", s.UdpRelayMode)
	u.RawQuery = query.Encode()
	return u.String()
}

func (s *Tuic) NeedPluginPort() bool {
	return false
}

func (s *Tuic) ProtoToShow() string {
	return "tuic"
}

func (s *Tuic) GetProtocol() string {
	return s.Protocol
}

func (s *Tuic) GetHostname() string {
	return s.Server
}

func (s *Tuic) GetPort() int {
	return s.Port
}

func (s *Tuic) GetName() string {
	return s.Name
}

func (s *Tuic) SetName(name string) {
	s.Name = name
}
