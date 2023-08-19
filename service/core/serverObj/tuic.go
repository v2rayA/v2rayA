package serverObj

import (
	"fmt"
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
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
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
	return &Tuic{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "tuic",
		Link:     link,
	}, nil
}

func (s *Tuic) Configuration(info PriorInfo) (c Configuration, err error) {
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

func (s *Tuic) ExportToURL() string {
	return s.Link
}

func (s *Tuic) NeedPluginPort() bool {
	return true
}

func (s *Tuic) ProtoToShow() string {
	return fmt.Sprintf("tuic")
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
