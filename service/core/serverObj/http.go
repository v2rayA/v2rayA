package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("http", NewHTTP)
	FromLinkRegister("https", NewHTTP)
	FromLinkRegister("http-proxy", NewHTTP)
	FromLinkRegister("https-proxy", NewHTTP)
	EmptyRegister("http", func() (ServerObj, error) {
		return &HTTP{Protocol: "http"}, nil
	})
	EmptyRegister("https", func() (ServerObj, error) {
		return &HTTP{Protocol: "https"}, nil
	})
}

type HTTP struct {
	Address  string `json:"address" server:"server" hostname:"hostname" add:"add" host:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewHTTP(link string) (ServerObj, error) {
	return ParseHttpURL(link)
}

func (h *HTTP) Configuration(info PriorInfo) (Configuration, error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	chain := []string{socks5.String(), h.ExportToURL()}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   false,
	}, nil
}

func (h *HTTP) ExportToURL() string {
	return h.Link
}

func (h *HTTP) NeedPluginPort() bool {
	return true
}

func (h *HTTP) ProtoToShow() string {
	return strings.ToUpper(h.Protocol)
}

func (h *HTTP) GetProtocol() string {
	return h.Protocol
}

func (h *HTTP) GetHostname() string {
	return h.Address
}

func (h *HTTP) GetPort() int {
	return h.Port
}

func (h *HTTP) GetName() string {
	return h.Name
}

func (h *HTTP) SetName(name string) {
	h.Name = name
}

func ParseHttpURL(link string) (*HTTP, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	host, portStr, _ := net.SplitHostPort(u.Host)
	if portStr == "" {
		if u.Scheme == "https" || u.Scheme == "https-proxy" {
			portStr = "443"
		} else {
			portStr = "80"
		}
	}
	port, _ := strconv.Atoi(portStr)
	h := &HTTP{
		Address:  host,
		Port:     port,
		Name:     u.Fragment,
		Link:     link,
		Protocol: "http",
	}
	if u.Scheme == "https" || u.Scheme == "https-proxy" {
		h.Protocol = "https"
	}
	if u.User != nil {
		h.Username = u.User.Username()
		h.Password, _ = u.User.Password()
	}
	return h, nil
}
