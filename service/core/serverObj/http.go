package serverObj

import (
	"net"
	"net/url"
	"strconv"

	"github.com/v2rayA/v2rayA/core/coreObj"
)

func init() {
	FromLinkRegister("http", NewHTTP)
	FromLinkRegister("https", NewHTTP)
	FromLinkRegister("http-proxy", NewHTTP)
	FromLinkRegister("https-proxy", NewHTTP)
	EmptyRegister("http", func() (ServerObj, error) {
		return new(HTTP), nil
	})
	EmptyRegister("https", func() (ServerObj, error) {
		return new(HTTP), nil
	})
	EmptyRegister("http-proxy", func() (ServerObj, error) {
		return new(HTTP), nil
	})
	EmptyRegister("https-proxy", func() (ServerObj, error) {
		return new(HTTP), nil
	})
}

type HTTP struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
}

func NewHTTP(link string) (ServerObj, error) {
	return ParseHttpURL(link)
}

func ParseHttpURL(u string) (data *HTTP, err error) {
	t, err := url.Parse(u)
	if err != nil {
		return nil, ErrInvalidParameter
	}
	port, err := strconv.Atoi(t.Port())
	if err != nil {
		return nil, ErrInvalidParameter
	}
	data = &HTTP{
		Name:   t.Fragment,
		Server: t.Hostname(),
		Port:   port,
	}
	if t.User != nil && len(t.User.String()) > 0 {
		data.Username = t.User.Username()
		data.Password, _ = t.User.Password()
	}
	switch t.Scheme {
	case "https-proxy", "https":
		data.Protocol = "https"
		if data.Port == 0 {
			data.Port = 443
		}
	case "http-proxy", "http":
		data.Protocol = "http"
		if data.Port == 0 {
			data.Port = 80
		}
	default:
		data.Protocol = t.Scheme
	}
	return data, nil
}

func (h *HTTP) Configuration(info PriorInfo) (c Configuration, err error) {
	var users []coreObj.OutboundUser
	if h.Username != "" && h.Password != "" {
		users = []coreObj.OutboundUser{
			{
				User: h.Username,
				Pass: h.Password,
			},
		}
	}
	o := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "http",
		Settings: coreObj.Settings{
			Servers: []coreObj.Server{{
				Address: h.Server,
				Port:    h.Port,
				Users:   users,
			}},
		},
	}
	if h.Protocol == "https" {
		//tls
		o.StreamSettings = &coreObj.StreamSettings{
			Network:  "tcp",
			Security: "tls",
			TLSSettings: &coreObj.TLSSettings{
				ServerName: h.Server,
			},
		}
	}
	return Configuration{
		CoreOutbound: o,
		PluginChain:  "",
		UDPSupport:   false,
	}, nil
}

func (h *HTTP) ExportToURL() string {
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

func (h *HTTP) NeedPluginPort() bool {
	return false
}

func (h *HTTP) ProtoToShow() string {
	return h.Protocol
}

func (h *HTTP) GetProtocol() string {
	return h.Protocol
}

func (h *HTTP) GetHostname() string {
	return h.Server
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
