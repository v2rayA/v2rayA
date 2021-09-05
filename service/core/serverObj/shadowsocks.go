package serverObj

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
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
		return new(Shadowsocks), nil
	})
	EmptyRegister("ss", func() (ServerObj, error) {
		return new(Shadowsocks), nil
	})
}

type Shadowsocks struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Cipher   string `json:"cipher"`
	Plugin   Sip003 `json:"plugin"`
	Protocol string `json:"protocol"`
}

func NewShadowsocks(link string) (ServerObj, error) {
	return ParseSSURL(link)
}

func ParseSSURL(u string) (data *Shadowsocks, err error) {
	// parse attempts to parse ss:// links
	parse := func(content string) (v *Shadowsocks, ok bool) {
		// try to parse in the format of ss://BASE64(method:password)@server:port/?plugin=xxxx#name
		u, err := url.Parse(content)
		if err != nil {
			return nil, false
		}
		username := u.User.String()
		username, _ = common.Base64URLDecode(username)
		arr := strings.Split(username, ":")
		if len(arr) != 2 {
			return nil, false
		}
		cipher := arr[0]
		password := arr[1]
		var sip003 Sip003
		plugin := u.Query().Get("plugin")
		if len(plugin) > 0 {
			arr = strings.Split(plugin, ";")
			sip003.Name = arr[0]
			switch sip003.Name {
			case "obfs-local", "simpleobfs":
				sip003.Name = "simple-obfs"
			}
			for i := 1; i < len(arr); i++ {
				a := strings.Split(arr[i], "=")
				switch a[0] {
				case "obfs":
					sip003.Obfs = a[1]
				case "obfs-path", "obfs-uri":
					if !strings.HasPrefix(a[1], "/") {
						a[1] += "/"
					}
					sip003.Uri = a[1]
				case "obfs-host":
					sip003.Host = a[1]
				}
			}
		}
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, false
		}
		return &Shadowsocks{
			Cipher:   strings.ToLower(cipher),
			Password: password,
			Server:   u.Hostname(),
			Port:     port,
			Name:     u.Fragment,
			Plugin:   sip003,
			Protocol: "shadowsocks",
		}, true
	}
	var (
		v  *Shadowsocks
		ok bool
	)
	content := u
	// try to parse the ss:// link, if it fails, base64 decode first
	if v, ok = parse(content); !ok {
		// 进行base64解码，并unmarshal到VmessInfo上
		t := content[5:]
		var l, r string
		if ind := strings.Index(t, "#"); ind > -1 {
			l = t[:ind]
			r = t[ind+1:]
		} else {
			l = t
		}
		l, err = common.Base64StdDecode(l)
		if err != nil {
			l, err = common.Base64URLDecode(l)
			if err != nil {
				return
			}
		}
		t = "ss://" + l
		if len(r) > 0 {
			t += "#" + r
		}
		v, ok = parse(t)
	}
	if !ok {
		return nil, fmt.Errorf("%w: unrecognized ss address", InvalidParameterErr)
	}
	return v, nil
}

func (s *Shadowsocks) Configuration(info PriorInfo) (c Configuration, err error) {
	switch s.Cipher {
	case "aes-256-gcm", "aes-128-gcm", "chacha20-poly1305", "chacha20-ietf-poly1305", "plain", "none":
	default:
		return c, fmt.Errorf("unsupported shadowsocks encryption method: %v", s.Cipher)
	}
	v2rayServer := coreObj.Server{
		Address:  s.Server,
		Port:     s.Port,
		Method:   s.Cipher,
		Password: s.Password,
	}
	var chain []string
	switch s.Plugin.Name {
	case "simple-obfs":
		v2rayServer.Address = "127.0.0.1"
		v2rayServer.Port = info.PluginPort

		tcp := url.URL{
			Scheme: "tcp",
			Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
		}
		simpleObfs := url.URL{
			Scheme: "simple-obfs",
			Host:   net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
			RawQuery: url.Values{
				"obfs": []string{s.Plugin.Obfs},
				"host": []string{s.Plugin.Host},
				"uri":  []string{s.Plugin.Uri},
			}.Encode(),
		}
		chain = append(chain, tcp.String(), simpleObfs.String())
		switch s.Plugin.Obfs {
		case "http", "tls":
		default:
			return c, fmt.Errorf("unsupported obfs %v of plugin %v", s.Plugin.Obfs, s.Plugin.Name)
		}
	// TODO: v2ray-plugin
	//case "v2ray-plugin":
	case "":
		// no plugin
	default:
		return c, fmt.Errorf("unsupported plugin %v", s.Plugin.Name)
	}
	return Configuration{
		CoreOutbound: coreObj.OutboundObject{
			Tag:      info.Tag,
			Protocol: "shadowsocks",
			Settings: coreObj.Settings{
				Servers: []coreObj.Server{v2rayServer},
			},
		},
		PluginChain: strings.Join(chain, ","),
		UDPSupport:  true,
	}, nil
}

func (s *Shadowsocks) ExportToURL() string {
	// sip002
	u := &url.URL{
		Scheme:   "ss",
		User:     url.UserPassword(s.Cipher, s.Password),
		Host:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		Fragment: s.Name,
	}
	if s.Plugin.Name != "" {
		q := u.Query()
		q.Set("plugin", s.Plugin.String())
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func (s *Shadowsocks) NeedPlugin() bool {
	return len(s.Plugin.Name) > 0
}

func (s *Shadowsocks) GetProtocol() string {
	return s.Protocol
}

func (s *Shadowsocks) ProtoToShow() string {
	if s.Plugin.Name != "" {
		return fmt.Sprintf("%v(%v+%v)", s.Protocol, s.Cipher, s.Plugin.Obfs)
	}
	return fmt.Sprintf("%v(%v)", s.Protocol, s.Cipher)
}

func (s *Shadowsocks) GetHostname() string {
	return s.Server
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

type Sip003 struct {
	Name string `json:"name"`
	Obfs string `json:"obfs"`
	Host string `json:"host"`
	Uri  string `json:"uri"`
}

func (s *Sip003) String() string {
	list := []string{s.Name}
	if s.Obfs != "" {
		list = append(list, "obfs="+s.Obfs)
	}
	if s.Host != "" {
		list = append(list, "obfs-host="+s.Host)
	}
	if s.Uri != "" {
		list = append(list, "obfs-uri="+s.Uri)
	}
	return strings.Join(list, ";")
}
