package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/v2rayA/v2rayA/core/coreObj"
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
	core := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "http",
	}
	var users []coreObj.OutboundUser
	if h.Username != "" {
		users = append(users, coreObj.OutboundUser{
			User: h.Username,
			Pass: h.Password,
		})
	}
	core.Settings.Servers = []coreObj.Server{
		{
			Address: h.Address,
			Port:    h.Port,
			Users:   users,
		},
	}
	if strings.ToLower(h.Protocol) == "https" {
		core.StreamSettings = &coreObj.StreamSettings{
			Security: "tls",
			TLSSettings: &coreObj.TLSSettings{
				ServerName: h.Address,
				Alpn:       []string{"http/1.1"},
			},
		}
	}
	return Configuration{
		CoreOutbound: core,
		PluginChain:  "",
		UDPSupport:   false,
	}, nil
}

func (h *HTTP) ExportToURL() string {
	return h.Link
}

func (h *HTTP) NeedPluginPort() bool {
	return false
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
