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
	FromLinkRegister("juicity", NewJuicity)
	EmptyRegister("juicity", func() (ServerObj, error) {
		return &Juicity{Protocol: "juicity"}, nil
	})
}

type Juicity struct {
	Address               string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port                  int    `json:"port"`
	UUID                  string `json:"uuid"`
	Password              string `json:"password"`
	Sni                   string `json:"sni"`
	AllowInsecure         bool   `json:"allowInsecure"`
	CC                    string `json:"cc"`
	PinnedCertchainSha256 string `json:"pinnedCertchainSha256"`
	Name                  string `json:"name"`
	Protocol              string `json:"protocol"`
	Link                  string `json:"link"`
}

func NewJuicity(link string) (ServerObj, error) {
	return ParseJuicityURL(link)
}

func ParseJuicityURL(link string) (*Juicity, error) {
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
	var uuid, password string
	if u.User != nil {
		uuid = u.User.Username()
		password, _ = u.User.Password()
	}
	q := u.Query()
	cc := q.Get("congestion_control")
	if cc == "" {
		cc = q.Get("congestionControl")
	}
	return &Juicity{
		Address:               host,
		Port:                  port,
		UUID:                  uuid,
		Password:              password,
		Sni:                   q.Get("sni"),
		AllowInsecure:         q.Get("allow_insecure") == "1" || q.Get("allowInsecure") == "true",
		CC:                    cc,
		PinnedCertchainSha256: q.Get("pinned_certchain_sha256"),
		Name:                  u.Fragment,
		Link:                  link,
		Protocol:              "juicity",
	}, nil
}

// juicitySettings holds the settings serialized into the hybrid-core xray config.
type juicitySettings struct {
	// Address is "host:port".
	Address           string `json:"address"`
	UUID              string `json:"uuid"`
	Password          string `json:"password"`
	SNI               string `json:"sni,omitempty"`
	AllowInsecure     bool   `json:"allow_insecure,omitempty"`
	CongestionControl string `json:"congestion_control,omitempty"`
	PinnedSHA256      string `json:"pinned_certchain_sha256,omitempty"`
}

func (s *Juicity) Configuration(info PriorInfo) (c Configuration, err error) {
	settingsJSON, err := json.Marshal(juicitySettings{
		Address:           net.JoinHostPort(s.Address, strconv.Itoa(s.Port)),
		UUID:              s.UUID,
		Password:          s.Password,
		SNI:               s.Sni,
		AllowInsecure:     s.AllowInsecure,
		CongestionControl: s.CC,
		PinnedSHA256:      s.PinnedCertchainSha256,
	})
	if err != nil {
		return c, fmt.Errorf("juicity: marshal settings: %w", err)
	}

	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:      info.Tag,
			Protocol: "juicity",
			Settings: coreObj.Settings{Inlined: settingsJSON},
		},
		UDPSupport: true,
	}, nil
}

func (s *Juicity) ExportToURL() string {
	return s.Link
}

func (s *Juicity) NeedPluginPort() bool {
	return false
}

func (s *Juicity) ProtoToShow() string {
	return "Juicity"
}

func (s *Juicity) GetProtocol() string {
	return "juicity"
}

func (s *Juicity) GetHostname() string {
	return s.Address
}

func (s *Juicity) GetPort() int {
	return s.Port
}

func (s *Juicity) GetName() string {
	return s.Name
}

func (s *Juicity) SetName(name string) {
	s.Name = name
}
