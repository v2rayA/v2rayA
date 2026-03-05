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
		return new(Tuic), nil
	})
}

type Tuic struct {
	Name              string `json:"name"`
	Server            string `json:"server"`
	Port              int    `json:"port"`
	UUID              string `json:"uuid"`
	Password          string `json:"password"`
	Sni               string `json:"sni"`
	AllowInsecure     bool   `json:"allowInsecure"`
	DisableSni        bool   `json:"disableSni"`
	Alpn              string `json:"alpn"`
	CongestionControl string `json:"congestionControl"`
	UdpRelayMode      string `json:"udpRelayMode"`
	Protocol          string `json:"protocol"`
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
		Name:              u.Fragment,
		Server:            u.Hostname(),
		Port:              port,
		UUID:              u.User.Username(),
		Password:          u.User.String(),
		Sni:               u.Query().Get("sni"),
		AllowInsecure:     u.Query().Get("allow_insecure") == "true" || u.Query().Get("allow_insecure") == "1",
		DisableSni:        u.Query().Get("disable_sni") == "true" || u.Query().Get("disable_sni") == "1",
		Alpn:              alpn,
		CongestionControl: u.Query().Get("congestion_control"),
		UdpRelayMode:      u.Query().Get("udp_relay_mode"),
		Protocol:          "tuic",
	}
	if data.Password != "" && strings.Contains(data.Password, ":") {
		data.Password = strings.SplitN(data.Password, ":", 2)[1]
	} else if data.Password != "" {
		// handle case where password might be just the password part if user:pass was not used
		// but typically u.User.String() is user:pass or user
	}
	return data, nil
}

func (s *Tuic) Configuration(info PriorInfo) (c Configuration, err error) {
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
	u := &url.URL{
		Scheme:   "tuic",
		User:     url.UserPassword(s.UUID, s.Password),
		Host:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		Fragment: s.Name,
	}
	query := u.Query()
	if s.AllowInsecure {
		query.Set("allow_insecure", "true")
	}
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
	return true
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
