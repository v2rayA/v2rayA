package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"github.com/v2rayA/v2rayA/core/coreObj"
)

func init() {
	FromLinkRegister("socks5", NewSOCKS)
	FromLinkRegister("socks", NewSOCKS)
	EmptyRegister("socks5", func() (ServerObj, error) {
		return &SOCKS{Protocol: "socks"}, nil
	})
	EmptyRegister("socks", func() (ServerObj, error) {
		return &SOCKS{Protocol: "socks"}, nil
	})
}

type SOCKS struct {
	Address  string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
	Link     string `json:"link"`
}

func NewSOCKS(link string) (ServerObj, error) {
	return ParseSocksURL(link)
}

func (h *SOCKS) Configuration(info PriorInfo) (Configuration, error) {
	servers := []coreObj.Server{
		{
			Address: h.Address,
			Port:    h.Port,
		},
	}
	if h.Username != "" {
		servers[0].Users = []coreObj.OutboundUser{
			{
				User: h.Username,
				Pass: h.Password,
			},
		}
	}
	coreOutbound := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "socks",
		Settings: coreObj.Settings{
			Servers: servers,
		},
	}
	return Configuration{
		CoreOutbound: coreOutbound,
		UDPSupport:   true,
	}, nil
}

func (h *SOCKS) ExportToURL() string {
	return h.Link
}

func (h *SOCKS) NeedPluginPort() bool {
	return false
}

func (h *SOCKS) ProtoToShow() string {
	return "SOCKS5"
}

func (h *SOCKS) GetProtocol() string {
	return "socks"
}

func (h *SOCKS) GetHostname() string {
	return h.Address
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

func ParseSocksURL(link string) (*SOCKS, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	host, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)
	s := &SOCKS{
		Address:  host,
		Port:     port,
		Name:     u.Fragment,
		Link:     link,
		Protocol: "socks",
	}
	if u.User != nil {
		s.Username = u.User.Username()
		s.Password, _ = u.User.Password()
	}
	return s, nil
}
