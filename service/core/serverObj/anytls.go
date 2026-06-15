package serverObj

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/v2rayA/v2rayA/core/coreObj"
)

func init() {
	FromLinkRegister("anytls", NewAnyTLS)
	EmptyRegister("anytls", func() (ServerObj, error) {
		return &AnyTLS{Protocol: "anytls"}, nil
	})
}

type AnyTLS struct {
	Address  string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewAnyTLS(link string) (ServerObj, error) {
	return ParseAnyTLSURL(link)
}

func ParseAnyTLSURL(link string) (*AnyTLS, error) {
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
	return &AnyTLS{
		Address:  host,
		Port:     port,
		Name:     u.Fragment,
		Link:     link,
		Protocol: "anytls",
	}, nil
}

// anytlsSettings holds the settings serialized into the hybrid-core xray config.
type anytlsSettings struct {
	Address         string `json:"address"`
	Port            int    `json:"port"`
	Password        string `json:"password"`
	SNI             string `json:"sni,omitempty"`
	MinIdleSessions int    `json:"min_idle_sessions,omitempty"`
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
		Address:         s.Address,
		Port:            s.Port,
		Password:        password,
		SNI:             sni,
		MinIdleSessions: minIdle,
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
	return "anytls"
}

func (s *AnyTLS) GetProtocol() string {
	return "anytls"
}

func (s *AnyTLS) GetHostname() string {
	return s.Address
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
