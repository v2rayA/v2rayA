package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"
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

func (s *Juicity) Configuration(info PriorInfo) (Configuration, error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	chain := []string{socks5.String(), s.ExportToURL()}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   true,
	}, nil
}

func (s *Juicity) ExportToURL() string {
	return s.Link
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

func (s *Juicity) NeedPluginPort() bool {
	return true
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
	return &Juicity{
		Address:               host,
		Port:                  port,
		UUID:                  uuid,
		Password:              password,
		Sni:                   q.Get("sni"),
		AllowInsecure:         q.Get("allow_insecure") == "1",
		CC:                    q.Get("congestion_control"),
		PinnedCertchainSha256: q.Get("pinned_certchain_sha256"),
		Name:                  u.Fragment,
		Link:                  link,
		Protocol:              "juicity",
	}, nil
}
