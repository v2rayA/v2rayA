package serverObj

import (
	"fmt"
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
		return nil, fmt.Errorf("%w: http proxy link is not a valid URL; expected http-proxy://[user:pass@]host:port", ErrInvalidParameter)
	}
	// A proxy link is scheme://[user:pass@]host[:port][#name]. Anything with
	// a path or query is almost always a subscription URL pasted into the
	// wrong box, so name that instead of returning a bare parameter error.
	if (t.Path != "" && t.Path != "/") || t.RawQuery != "" {
		return nil, fmt.Errorf("%w: http proxy link for %q has a path or query; expected http-proxy://[user:pass@]host:port", ErrInvalidParameter, t.Hostname())
	}
	// A default port only makes sense with a host to connect to; an empty
	// hostname used to be rejected by the port conversion.
	if t.Hostname() == "" {
		return nil, fmt.Errorf("%w: http proxy link has no server address; expected http-proxy://[user:pass@]host:port", ErrInvalidParameter)
	}
	// An absent port falls through to the per-scheme default below.
	port := 0
	if p := t.Port(); p != "" {
		if port, err = strconv.Atoi(p); err != nil {
			return nil, fmt.Errorf("%w: http proxy link for %q has an invalid port; expected http-proxy://[user:pass@]host:port", ErrInvalidParameter, t.Hostname())
		}
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
	return true
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
