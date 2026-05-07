package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("tuic", NewTuic)
	EmptyRegister("tuic", func() (ServerObj, error) {
		return &Tuic{Protocol: "tuic"}, nil
	})
}

type Tuic struct {
	Address        string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port           int    `json:"port"`
	UUID           string `json:"uuid"`
	Password       string `json:"password"`
	CC             string `json:"cc"`
	AllowInsecure  bool   `json:"allowInsecure"`
	DisableSni     bool   `json:"disableSni"`
	Sni            string `json:"sni"`
	Alpn           string `json:"alpn"`
	UdpRelayMode   string `json:"udpRelayMode"`
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	Link           string `json:"link"`
}

func NewTuic(link string) (ServerObj, error) {
	return ParseTuicURL(link)
}

func (s *Tuic) Configuration(info PriorInfo) (Configuration, error) {
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

func (s *Tuic) ExportToURL() string {
	return s.Link
}

func (s *Tuic) ProtoToShow() string {
	return "Tuic"
}

func (s *Tuic) GetProtocol() string {
	return "tuic"
}

func (s *Tuic) GetHostname() string {
	return s.Address
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

func (s *Tuic) NeedPluginPort() bool {
	return true
}

func ParseTuicURL(link string) (*Tuic, error) {
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
	return &Tuic{
		Address:        host,
		Port:           port,
		UUID:           uuid,
		Password:       password,
		CC:             q.Get("congestion_control"),
		AllowInsecure:  q.Get("allow_insecure") == "1",
		DisableSni:     q.Get("disable_sni") == "1",
		Sni:            q.Get("sni"),
		Alpn:           q.Get("alpn"),
		UdpRelayMode:   q.Get("udp_relay_mode"),
		Name:           u.Fragment,
		Link:           link,
		Protocol:       "tuic",
	}, nil
}
