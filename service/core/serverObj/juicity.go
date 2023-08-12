package serverObj

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("juicity", NewJuicity)
	EmptyRegister("juicity", func() (ServerObj, error) {
		return new(Juicity), nil
	})
}

type Juicity struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewJuicity(link string) (ServerObj, error) {
	return ParseJuicityURL(link)
}

func ParseJuicityURL(link string) (data *Juicity, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	return &Juicity{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "juicity",
		Link:     link,
	}, nil
}

func (s *Juicity) Configuration(info PriorInfo) (c Configuration, err error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	chain := []string{socks5.String(), s.Link}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   true,
	}, nil
}

func (s *Juicity) ExportToURL() string {
	return s.Link
}

func (s *Juicity) NeedPluginPort() bool {
	return true
}

func (s *Juicity) ProtoToShow() string {
	return fmt.Sprintf("Juicity")
}

func (s *Juicity) GetProtocol() string {
	return s.Protocol
}

func (s *Juicity) GetHostname() string {
	return s.Server
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
