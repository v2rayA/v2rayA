package serverObj

import (
	"encoding/base64"
	"fmt"
	"github.com/v2rayA/v2rayA/core/coreObj"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("shadowsocks", NewShadowsocks)
	FromLinkRegister("ss", NewShadowsocks)
	EmptyRegister("shadowsocks", func() (ServerObj, error) {
		return &Shadowsocks{Protocol: "ss"}, nil
	})
	EmptyRegister("ss", func() (ServerObj, error) {
		return &Shadowsocks{Protocol: "ss"}, nil
	})
}

type Shadowsocks struct {
	Address  string `json:"address" server:"server" hostname:"hostname" add:"add"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Cipher   string `json:"cipher" method:"method"`
	Plugin   struct {
		Name string `json:"name"`
		Opts struct {
			Tls  string `json:"tls"`
			Obfs string `json:"obfs"`
			Host string `json:"host"`
			Uri  string `json:"uri"`
			Impl string `json:"impl"`
		} `json:"opts"`
	} `json:"plugin"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

func NewShadowsocks(link string) (ServerObj, error) {
	return ParseSSURL(link)
}

func (s *Shadowsocks) Configuration(info PriorInfo) (Configuration, error) {
	coreOutbound := coreObj.OutboundObject{
		Tag:      info.Tag,
		Protocol: "shadowsocks",
		Settings: coreObj.Settings{
			Servers: []coreObj.Server{
				{
					Address:  s.Address,
					Port:     s.Port,
					Method:   s.Cipher,
					Password: s.Password,
				},
			},
		},
	}
	return Configuration{
		CoreOutbound: coreOutbound,
		UDPSupport:   true,
	}, nil
}

func (s *Shadowsocks) ExportToURL() string {
	return s.Link
}

func (s *Shadowsocks) ProtoToShow() string {
	return fmt.Sprintf("SS(%v)", s.Cipher)
}

func (s *Shadowsocks) GetProtocol() string {
	return "ss"
}

func (s *Shadowsocks) GetHostname() string {
	return s.Address
}

func (s *Shadowsocks) GetPort() int {
	return s.Port
}

func (s *Shadowsocks) GetName() string {
	return s.Name
}

func (s *Shadowsocks) SetName(name string) {
	s.Name = name
}

func (s *Shadowsocks) NeedPluginPort() bool {
	return false
}

func ParseSSURL(link string) (*Shadowsocks, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	var server, password, cipher string
	var port int
	
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
		if password == "" {
			// SIP002: ss://BASE64(cipher:password)@host:port
			data := cipher
			var b []byte
			var err error
			// Try all 4 base64 variants
			if b, err = base64.StdEncoding.DecodeString(data); err != nil {
				if b, err = base64.RawStdEncoding.DecodeString(data); err != nil {
					if b, err = base64.URLEncoding.DecodeString(data); err != nil {
						b, err = base64.RawURLEncoding.DecodeString(data)
					}
				}
			}
			if err == nil {
				pre := strings.SplitN(string(b), ":", 2)
				if len(pre) == 2 {
					cipher = pre[0]
					password = pre[1]
				}
			}
		}
	}
	
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		portStr = "8388"
	}
	port, _ = strconv.Atoi(portStr)
	server = host
	
	q := u.Query()
	ss := &Shadowsocks{
		Address:  server,
		Port:     port,
		Password: password,
		Cipher:   cipher,
		Name:     u.Fragment,
		Link:     link,
		Protocol: "ss",
	}
	if plugin := q.Get("plugin"); plugin != "" {
		parts := strings.SplitN(plugin, ";", 2)
		ss.Plugin.Name = parts[0]
		if len(parts) == 2 {
			opts := parts[1]
			for _, opt := range strings.Split(opts, ";") {
				kv := strings.SplitN(opt, "=", 2)
				if len(kv) == 2 {
					switch kv[0] {
					case "tls":
						ss.Plugin.Opts.Tls = kv[1]
					case "obfs":
						ss.Plugin.Opts.Obfs = kv[1]
					case "host":
						ss.Plugin.Opts.Host = kv[1]
					case "uri":
						ss.Plugin.Opts.Uri = kv[1]
					case "impl":
						ss.Plugin.Opts.Impl = kv[1]
					}
				} else if kv[0] == "tls" {
					ss.Plugin.Opts.Tls = "tls"
				}
			}
		}
	}
	
	return ss, nil
}
