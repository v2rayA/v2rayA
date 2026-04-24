package serverObj

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("anytls", NewAnyTLS)
	EmptyRegister("anytls", func() (ServerObj, error) {
		return &AnyTLS{Protocol: "anytls"}, nil
	})
}

type AnyTLS struct {
	Address  string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewAnyTLS(link string) (ServerObj, error) {
	return ParseAnyTLSURL(link)
}

func (s *AnyTLS) Configuration(info PriorInfo) (Configuration, error) {
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
	return "anytls"
}

func (s *AnyTLS) GetProtocol() string {
	return "anytls"
}

func (s *AnyTLS) GetHostname() string {
	return s.Address
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

func ParseAnyTLSURL(link string) (*AnyTLS, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	host, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)
	return &AnyTLS{
		Address:  host,
		Port:     port,
		Name:     u.Fragment,
		Link:     link,
		Protocol: "anytls",
	}, nil
}
