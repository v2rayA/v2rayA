package serverObj

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("shadowsocksr", NewShadowsocksR)
	FromLinkRegister("ssr", NewShadowsocksR)
	EmptyRegister("shadowsocksr", func() (ServerObj, error) {
		return &ShadowsocksR{Protocol: "shadowsocksr"}, nil
	})
	EmptyRegister("ssr", func() (ServerObj, error) {
		return &ShadowsocksR{Protocol: "shadowsocksr"}, nil
	})
}

type ShadowsocksR struct {
	Address    string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
	Cipher     string `json:"cipher" method:"method"`
	Proto      string `json:"proto"`
	ProtoParam string `json:"protoParam"`
	Obfs       string `json:"obfs"`
	ObfsParam  string `json:"obfsParam"`
	Name       string `json:"name" remarks:"remarks"`
	Protocol   string `json:"protocol"`
	Link       string `json:"link"`
}

func decodeSSR(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if pad := len(s) % 4; pad != 0 {
		s += strings.Repeat("=", 4-pad)
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s, nil
	}
	return string(b), nil
}

func NewShadowsocksR(link string) (ServerObj, error) {
	return ParseSSRURL(link)
}

func ParseSSRURL(link string) (*ShadowsocksR, error) {
	content := link
	if strings.HasPrefix(link, "ssr://") {
		content = link[6:]
	}
	if idx := strings.IndexAny(content, "#?"); idx != -1 {
		content = content[:idx]
	}
	decodedContent, err := decodeSSR(content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SSR content: %v", err)
	}
	arr := strings.Split(decodedContent, "/?")
	pre := strings.Split(arr[0], ":")
	if len(pre) < 6 {
		return nil, fmt.Errorf("invalid ssr format")
	}
	passIdx := len(pre) - 1
	obfsIdx := len(pre) - 2
	methodIdx := len(pre) - 3
	protoIdx := len(pre) - 4
	portIdx := len(pre) - 5
	server := strings.Join(pre[:portIdx], ":")
	port, _ := strconv.Atoi(pre[portIdx])
	password, _ := decodeSSR(pre[passIdx])
	s := &ShadowsocksR{
		Address:  server,
		Port:     port,
		Proto:    pre[protoIdx],
		Cipher:   pre[methodIdx],
		Obfs:     pre[obfsIdx],
		Password: password,
		Protocol: "shadowsocksr",
		Link:     link,
	}
	if len(arr) > 1 {
		q, _ := url.ParseQuery(arr[1])
		s.Name, _ = decodeSSR(q.Get("remarks"))
		s.ProtoParam, _ = decodeSSR(q.Get("protoparam"))
		s.ObfsParam, _ = decodeSSR(q.Get("obfsparam"))
	}
	if s.Name == "" {
		if u, err := url.Parse(link); err == nil {
			s.Name = u.Fragment
		}
	}
	return s, nil
}

func (s *ShadowsocksR) Configuration(info PriorInfo) (Configuration, error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	chain := []string{socks5.String(), s.ExportToURL()}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   true,
	}, nil
}

func (s *ShadowsocksR) ExportToURL() string {
	return s.Link
}

func (s *ShadowsocksR) NeedPluginPort() bool {
	return true
}

func (s *ShadowsocksR) ProtoToShow() string {
	return fmt.Sprintf("SSR(%v+%v)", s.Proto, s.Obfs)
}

func (s *ShadowsocksR) GetProtocol() string {
	return "shadowsocksr"
}

func (s *ShadowsocksR) GetHostname() string {
	return s.Address
}

func (s *ShadowsocksR) GetPort() int {
	return s.Port
}

func (s *ShadowsocksR) GetName() string {
	return s.Name
}

func (s *ShadowsocksR) SetName(name string) {
	s.Name = name
}
