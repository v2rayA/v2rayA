package serverObj

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func init() {
	FromLinkRegister("anytls", NewAnyTLS)
	EmptyRegister("anytls", func() (ServerObj, error) {
		return new(AnyTLS), nil
	})
}

type AnyTLS struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewAnyTLS(link string) (ServerObj, error) {
	return ParseAnyTLSURL(link)
}

func ParseAnyTLSURL(link string) (data *AnyTLS, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	return &AnyTLS{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "anytls",
		Link:     link,
	}, nil
}

// anytlsSettings holds the settings serialized into the hybrid-core xray config.
type anytlsSettings struct {
	Address              string `json:"address"`
	Port                 int    `json:"port"`
	Password             string `json:"password"`
	SNI                  string `json:"sni,omitempty"`
	MinIdleSessions      int    `json:"min_idle_sessions,omitempty"`
	PinnedPeerCertSha256 string `json:"pinnedPeerCertSha256,omitempty"`
	VerifyPeerCertByName string `json:"verifyPeerCertByName,omitempty"`
}

func (s *AnyTLS) Configuration(info PriorInfo) (c Configuration, err error) {
	u, err := url.Parse(s.Link)
	if err != nil {
		return c, fmt.Errorf("anytls: parse link: %w", err)
	}

	password := ""
	if u.User != nil {
		password = u.User.Username()
	}
	q := u.Query()
	sni := q.Get("sni")
	minIdle, _ := strconv.Atoi(q.Get("minIdleSession"))

	settingsJSON, err := json.Marshal(anytlsSettings{
		Address:              s.Server,
		Port:                 s.Port,
		Password:             password,
		SNI:                  sni,
		MinIdleSessions:      minIdle,
		PinnedPeerCertSha256: q.Get("pinnedPeerCertSha256"),
		VerifyPeerCertByName: q.Get("verifyPeerCertByName"),
	})
	if err != nil {
		return c, fmt.Errorf("anytls: marshal settings: %w", err)
	}

	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:      info.Tag,
			Protocol: "anytls",
			Settings: coreObj.Settings{Inlined: settingsJSON},
		},
		UDPSupport: true,
	}, nil
}

func (s *AnyTLS) ExportToURL() string {
	return s.Link
}

func (s *AnyTLS) NeedPluginPort() bool {
	return false
}

func (s *AnyTLS) ProtoToShow() string {
	return fmt.Sprintf("anytls")
}

func (s *AnyTLS) GetProtocol() string {
	return s.Protocol
}

func (s *AnyTLS) GetHostname() string {
	return s.Server
}

func (s *AnyTLS) GetPort() int {
	return s.Port
}

func (s *AnyTLS) GetName() string {
	return s.Name
}

func (s *AnyTLS) SetName(name string) {
	s.Name = name
}
