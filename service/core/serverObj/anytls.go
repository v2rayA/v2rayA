package serverObj

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("anytls", NewAnyTLS)
	EmptyRegister("anytls", func() (ServerObj, error) {
		return new(AnyTLS), nil
	})
}

type AnyTLS struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewAnyTLS(link string) (ServerObj, error) {
	return ParseAnyTLSURL(link)
}

func ParseAnyTLSURL(link string) (data *AnyTLS, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	return &AnyTLS{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Protocol: "anytls",
		Link:     link,
	}, nil
}

func (s *AnyTLS) Configuration(info PriorInfo) (c Configuration, err error) {
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

func (s *AnyTLS) ExportToURL() string {
	return s.Link
}

func (s *AnyTLS) NeedPluginPort() bool {
	return true
}

func (s *AnyTLS) ProtoToShow() string {
	return fmt.Sprintf("anytls")
}

func (s *AnyTLS) GetProtocol() string {
	return s.Protocol
}

func (s *AnyTLS) GetHostname() string {
	return s.Server
}

func (s *AnyTLS) GetPort() int {
	return s.Port
}

func (s *AnyTLS) GetName() string {
	return s.Name
}

func (s *AnyTLS) SetName(name string) {
	s.Name = name
}
