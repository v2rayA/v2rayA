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
		return new(Juicity), nil
	})
}

type Juicity struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewJuicity(link string) (ServerObj, error) {
	return ParseJuicityURL(link)
}

func ParseJuicityURL(link string) (data *Juicity, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	return &Juicity{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "juicity",
		Link:     link,
	}, nil
}

// juicitySettings holds the settings serialized into the hybrid-core xray config.
type juicitySettings struct {
	// Address is "host:port".
	Address           string `json:"address"`
	UUID              string `json:"uuid"`
	Password          string `json:"password"`
	SNI               string `json:"sni,omitempty"`
	CongestionControl string `json:"congestion_control,omitempty"`
	PinnedSHA256      string `json:"pinned_certchain_sha256,omitempty"`
}

func (s *Juicity) Configuration(info PriorInfo) (c Configuration, err error) {
	u, err := url.Parse(s.Link)
	if err != nil {
		return c, fmt.Errorf("juicity: parse link: %w", err)
	}

	uuid := ""
	password := ""
	if u.User != nil {
		uuid = u.User.Username()
		password, _ = u.User.Password()
	}
	q := u.Query()
	sni := q.Get("sni")
	congestion := q.Get("congestion_control")
	if congestion == "" {
		congestion = q.Get("congestionControl")
	}
	pinnedSHA := q.Get("pinned_certchain_sha256")

	settingsJSON, err := json.Marshal(juicitySettings{
		Address:           net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		UUID:              uuid,
		Password:          password,
		SNI:               sni,
		CongestionControl: congestion,
		PinnedSHA256:      pinnedSHA,
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
	return fmt.Sprintf("Juicity")
}

func (s *Juicity) GetProtocol() string {
	return s.Protocol
}

func (s *Juicity) GetHostname() string {
	return s.Server
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
