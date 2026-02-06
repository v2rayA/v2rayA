package serverObj

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("trojan", NewTrojan)
	FromLinkRegister("trojan-go", NewTrojan)
	EmptyRegister("trojan", func() (ServerObj, error) {
		return new(Trojan), nil
	})
	EmptyRegister("trojan-go", func() (ServerObj, error) {
		return new(Trojan), nil
	})
}

type Trojan struct {
	Name          string `json:"name"`
	Server        string `json:"server"`
	Port          int    `json:"port"`
	Password      string `json:"password"`
	Sni           string `json:"sni"`
	Type          string `json:"type"`
	Encryption    string `json:"encryption"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	ServiceName   string `json:"serviceName"`
	AllowInsecure bool   `json:"allowInsecure"`
	Alpn          string `json:"alpn,omitempty"`
	Protocol      string `json:"protocol"`
}

func NewTrojan(link string) (ServerObj, error) {
	return ParseTrojanURL(link)
}

func ParseTrojanURL(u string) (data *Trojan, err error) {
	//trojan://password@server:port#escape(remarks)
	t, err := url.Parse(u)
	if err != nil {
		err = fmt.Errorf("invalid trojan format")
		return
	}
	allowInsecure := t.Query().Get("allowInsecure")
	sni := t.Query().Get("peer")
	if sni == "" {
		sni = t.Query().Get("sni")
	}
	if sni == "" {
		sni = t.Hostname()
	}

	port, err := strconv.Atoi(t.Port())
	if err != nil {
		return nil, ErrInvalidParameter
	}
	data = &Trojan{
		Name:          t.Fragment,
		Server:        t.Hostname(),
		Port:          port,
		Password:      t.User.Username(),
		Sni:           sni,
		Alpn:          t.Query().Get("alpn"),
		Type:          t.Query().Get("type"),
		Path:          t.Query().Get("path"),
		ServiceName:   t.Query().Get("serviceName"),
		AllowInsecure: allowInsecure == "1" || allowInsecure == "true",
		Protocol:      "trojan",
	}
	if t.Scheme == "trojan-go" {
		data.Protocol = "trojan-go"
		data.Encryption = t.Query().Get("encryption")
		data.Host = t.Query().Get("host")
		data.Path = t.Query().Get("path")
		data.ServiceName = t.Query().Get("serviceName")
		data.Type = t.Query().Get("type")
		data.AllowInsecure = false
	}
	return data, nil
}

func (t *Trojan) Configuration(info PriorInfo) (c Configuration, err error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	chain := []string{socks5.String(), t.ExportToURL()}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   true,
	}, nil
}

func (t *Trojan) ExportToURL() string {

	u := &url.URL{
		Scheme:   "trojan",
		User:     url.User(t.Password),
		Host:     net.JoinHostPort(t.Server, strconv.Itoa(t.Port)),
		Fragment: t.Name,
	}

	query := u.Query()
	setValue(&query, "type", t.Type)

	net := strings.ToLower(t.Type)

	switch net {
	case "websocket", "ws", "http", "h2":
		setValue(&query, "path", t.Path)
		setValue(&query, "host", t.Host)
	case "mkcp", "kcp":
		setValue(&query, "headerType", t.Type)
		setValue(&query, "seed", t.Path)
	case "tcp":
		setValue(&query, "headerType", t.Type)
		setValue(&query, "host", t.Host)
		setValue(&query, "path", t.Path)
	case "grpc":
		setValue(&query, "serviceName", t.ServiceName)
	}

	if t.AllowInsecure {
		query.Set("allowInsecure", "1")
	}
	setValue(&query, "sni", t.Sni)

	if t.Protocol == "trojan-go" {
		u.Scheme = "trojan-go"
		setValue(&query, "host", t.Host)
		setValue(&query, "encryption", t.Encryption)
		setValue(&query, "type", t.Type)
		setValue(&query, "path", t.Path)
		setValue(&query, "serviceName", t.ServiceName)
	}
	u.RawQuery = query.Encode()
	return u.String()
}

func (t *Trojan) NeedPluginPort() bool {
	return true
}

func (t *Trojan) ProtoToShow() string {
	if t.Protocol == "trojan" {
		return t.Protocol
	}
	if t.Encryption == "" {
		return fmt.Sprintf("%v(%v)", t.Protocol, t.Type)
	}
	return fmt.Sprintf("%v(%v+%v)", t.Protocol, t.Type, strings.Split(t.Encryption, ";")[0])
}

func (t *Trojan) GetProtocol() string {
	return t.Protocol
}

func (t *Trojan) GetHostname() string {
	return t.Server
}

func (t *Trojan) GetPort() int {
	return t.Port
}

func (t *Trojan) GetName() string {
	return t.Name
}

func (t *Trojan) SetName(name string) {
	t.Name = name
}
