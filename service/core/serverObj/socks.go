package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("socks5", NewSOCKS)
	EmptyRegister("socks5", func() (ServerObj, error) {
		return new(SOCKS), nil
	})
}

type SOCKS struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
}

func NewSOCKS(link string) (ServerObj, error) {
	return ParseSocksURL(link)
}

func ParseSocksURL(u string) (data *SOCKS, err error) {
	t, err := url.Parse(u)
	if err != nil {
		return nil, ErrInvalidParameter
	}
	port, err := strconv.Atoi(t.Port())
	if err != nil {
		return nil, ErrInvalidParameter
	}
	data = &SOCKS{
		Name:   t.Fragment,
		Server: t.Hostname(),
		Port:   port,
	}
	if t.User != nil && len(t.User.String()) > 0 {
		data.Username = t.User.Username()
		data.Password, _ = t.User.Password()
	}
	switch t.Scheme {
	case "socks5":
		data.Protocol = "socks5"
		if data.Port == 0 {
			data.Port = 1080
		}
	default:
		data.Protocol = t.Scheme
	}
	return data, nil
}

func (h *SOCKS) Configuration(info PriorInfo) (c Configuration, err error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	chain := []string{socks5.String(), h.ExportToURL()}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   true,
	}, nil
}

func (h *SOCKS) ExportToURL() string {
	var user *url.Userinfo
	if h.Username != "" && h.Password != "" {
		user = url.UserPassword(h.Username, h.Password)
	}
	u := &url.URL{
		Scheme:   h.Protocol,
		User:     user,
		Host:     net.JoinHostPort(h.Server, strconv.Itoa(h.Port)),
		Fragment: h.Name,
	}
	return u.String()
}

func (h *SOCKS) NeedPluginPort() bool {
	return true
}

func (h *SOCKS) ProtoToShow() string {
	return h.Protocol
}

func (h *SOCKS) GetProtocol() string {
	return h.Protocol
}

func (h *SOCKS) GetHostname() string {
	return h.Server
}

func (h *SOCKS) GetPort() int {
	return h.Port
}

func (h *SOCKS) GetName() string {
	return h.Name
}

func (h *SOCKS) SetName(name string) {
	h.Name = name
}
