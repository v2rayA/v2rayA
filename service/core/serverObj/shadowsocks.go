package serverObj

import (
	"encoding/base64"
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
		arr := strings.SplitN(username, ":", 2)
		if len(arr) != 2 {
			return nil, false
		}
		cipher := arr[0]
		password := arr[1]
		var sip003 Sip003
		plugin := u.Query().Get("plugin")
		if len(plugin) > 0 {
			sip003 = ParseSip003(plugin)
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

func (s *Shadowsocks) ConfigurationMC(info PriorInfo) (c Configuration, err error) {
	v2rayServer := coreObj.Server{
		Address:  s.Server,
		Port:     s.Port,
		Method:   s.Cipher,
		Password: s.Password,
	}
	udpSupport := false
	var proxySettings *coreObj.ProxySettings
	var extraOutbounds []coreObj.OutboundObject
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
				"obfs": []string{s.Plugin.Opts.Obfs},
				"host": []string{s.Plugin.Opts.Host},
				"uri":  []string{s.Plugin.Opts.Path},
			}.Encode(),
		}
		chain = append(chain, tcp.String(), simpleObfs.String())
		switch s.Plugin.Opts.Obfs {
		case "http", "tls":
		default:
			return c, fmt.Errorf("unsupported obfs %v of plugin %v", s.Plugin.Opts.Obfs, s.Plugin.Name)
		}
	case "v2ray-plugin":
		dialerTag := info.Tag + "-dialer"
		proxySettings = &coreObj.ProxySettings{
			Tag: dialerTag,
		}
		streamSettings := &coreObj.StreamSettings{}
		host := s.Plugin.Opts.Host
		if host == "" {
			host = "cloudflare.com"
		}
		path := s.Plugin.Opts.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if s.Plugin.Opts.Tls == "tls" {
			streamSettings.Security = "tls"
			streamSettings.TLSSettings = &coreObj.TLSSettings{}
			// SNI
			streamSettings.TLSSettings.ServerName = host
		}
		switch s.Plugin.Opts.Obfs {
		case "quic":
			return c, fmt.Errorf("quic is not yet supported")
		default:
			// "websocket" or ""
			streamSettings.Network = "ws"
			streamSettings.WsSettings = &coreObj.WsSettings{
				Path: path,
				Headers: coreObj.Headers{
					Host: host,
				},
			}
		}
		extraOutbounds = append(extraOutbounds, coreObj.OutboundObject{
			Tag:      dialerTag,
			Protocol: "freedom",
			Settings: coreObj.Settings{
				DomainStrategy: "AsIs",
				Redirect:       net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
			},
			StreamSettings: streamSettings,
			Mux: &coreObj.Mux{
				Enabled:     true,
				Concurrency: 1,
			},
		})
	case "":
		// no plugin
		udpSupport = true
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
			ProxySettings: proxySettings,
		},
		ExtraOutbounds: extraOutbounds,
		PluginChain:    strings.Join(chain, ","),
		UDPSupport:     udpSupport,
	}, nil
}

func (s *Shadowsocks) ConfigurationMT(info PriorInfo) (c Configuration, err error) {
	v2rayServer := coreObj.Server{
		Address:  s.Server,
		Port:     s.Port,
		Method:   s.Cipher,
		Password: s.Password,
	}
	udpSupport := true
	var v2StreamSettings *coreObj.StreamSettings
	var v2Mux *coreObj.Mux
	switch s.Plugin.Name {
	case "simple-obfs":
		host := s.Plugin.Opts.Host
		if host == "" {
			host = "cloudflare.com"
		}
		path := s.Plugin.Opts.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		v2StreamSettings = &coreObj.StreamSettings{}
		tcpSettings := coreObj.TCPSettings{
			Header: coreObj.TCPHeader{
				Type: "http",
				Request: coreObj.HTTPRequest{
					Version: "1.1",
					Method: "GET",
					Path: strings.Split(path, ","),
					Headers: coreObj.HTTPReqHeader{
						Host: strings.Split(host, ","),
						UserAgent: []string{
							"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/527.36 (KHTML, like Gecko) Chrome/55.2883.75 Safari/537.36",
							"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.46",
						},
						AcceptEncoding: []string{"gzip, deflate"},
						Connection: []string{"keep-alive"},
						Pragma: "no-cache",
					},
				},
				Response: coreObj.HTTPResponse{
					Version: "1.1",
					Status: "200",
					Reason: "OK",
					Headers: coreObj.HTTPRespHeader {
						ContentType: []string{"application/octet-stream", "video/mpeg"},
						TransferEncoding: []string{"chunked"},
						Connection: []string{"keep-alive"},
						Pragma: "no-cache",
					},
				},
			},
		}
		switch s.Plugin.Opts.Obfs {
		case "http":
			v2StreamSettings.Network = "tcp"
			v2StreamSettings.TCPSettings = &tcpSettings
		case "tls":
			v2StreamSettings.Security = "tls"
			v2StreamSettings.TLSSettings = &coreObj.TLSSettings{}
			// SNI
			v2StreamSettings.TLSSettings.ServerName = host
		default:
			return c, fmt.Errorf("unsupported obfs %v of plugin %v", s.Plugin.Opts.Obfs, s.Plugin.Name)
		}
	case "v2ray-plugin":
		v2StreamSettings = &coreObj.StreamSettings{}
		host := s.Plugin.Opts.Host
		if host == "" {
			host = "cloudflare.com"
		}
		path := s.Plugin.Opts.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if s.Plugin.Opts.Tls == "tls" {
			v2StreamSettings.Security = "tls"
			v2StreamSettings.TLSSettings = &coreObj.TLSSettings{}
			// SNI
			v2StreamSettings.TLSSettings.ServerName = host
		}
		v2Mux = &coreObj.Mux{
			Enabled: true,
			Concurrency: 1,
		}
		switch s.Plugin.Opts.Obfs {
		case "quic":
			return c, fmt.Errorf("quic is not yet supported")
		default:
			// "websocket" or ""
			v2StreamSettings.Network = "ws"
			v2StreamSettings.WsSettings = &coreObj.WsSettings{
				Path: path,
				Headers: coreObj.Headers{
					Host: host,
				},
			}
		}
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
			StreamSettings: v2StreamSettings,
			Mux: v2Mux,
		},
		UDPSupport:     udpSupport,
	}, nil
}

func (s *Shadowsocks) Configuration(info PriorInfo) (c Configuration, err error) {
	switch s.Cipher {
	case "aes-256-gcm", "aes-128-gcm", "chacha20-poly1305", "chacha20-ietf-poly1305", "plain", "none":
	default:
		return c, fmt.Errorf("unsupported shadowsocks encryption method: %v", s.Cipher)
	}
	// check shadowsocks plugin implementation
	ssPluginImpl := s.Plugin.Opts.Impl
	if ssPluginImpl == "" {
		switch s.Plugin.Name {
		case "simple-obfs":
			ssPluginImpl = "transport"
		default:
			ssPluginImpl = "chained"
		}
	}
	switch ssPluginImpl {
	case "chained":
		return s.ConfigurationMC(info)
	case "transport":
		return s.ConfigurationMT(info)
	default:
		return c, fmt.Errorf("unsupported shadowsocks plugin implementation: %v", ssPluginImpl)
	}
}

func (s *Shadowsocks) ExportToURL() string {
	// sip002
	u := &url.URL{
		Scheme:   "ss",
		User:     url.User(strings.TrimSuffix(base64.URLEncoding.EncodeToString([]byte(s.Cipher+":"+s.Password)), "=")),
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

func (s *Shadowsocks) NeedPluginPort() bool {
	return len(s.Plugin.Name) > 0
}

func (s *Shadowsocks) GetProtocol() string {
	return s.Protocol
}

func (s *Shadowsocks) ProtoToShow() string {
	ciph := s.Cipher
	if ciph == "chacha20-ietf-poly1305" || ciph == "chacha20-poly1305" {
		ciph = "c20p1305"
	}
	if s.Plugin.Name != "" {
		return fmt.Sprintf("SS(%v+%v)", ciph, s.Plugin.Name)
	}
	return fmt.Sprintf("SS(%v)", ciph)
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
	Name string     `json:"name"`
	Opts Sip003Opts `json:"opts"`
}
type Sip003Opts struct {
	Tls  string `json:"tls"`
	Obfs string `json:"obfs"`
	Host string `json:"host"`
	Path string `json:"uri"`
	Impl string `json:"impl"`
}

func ParseSip003Opts(opts string) Sip003Opts {
	var sip003Opts Sip003Opts
	fields := strings.Split(opts, ";")
	for i := range fields {
		a := strings.Split(fields[i], "=")
		if len(a) == 1 {
			// to avoid panic
			a = append(a, "")
		}
		switch a[0] {
		case "tls":
			sip003Opts.Tls = "tls"
		case "obfs", "mode":
			sip003Opts.Obfs = a[1]
		case "obfs-path", "obfs-uri", "path":
			if !strings.HasPrefix(a[1], "/") {
				a[1] += "/"
			}
			sip003Opts.Path = a[1]
		case "obfs-host", "host":
			sip003Opts.Host = a[1]
		case "impl":
			sip003Opts.Impl = a[1]
		}
	}
	return sip003Opts
}
func ParseSip003(plugin string) Sip003 {
	var sip003 Sip003
	fields := strings.SplitN(plugin, ";", 2)
	switch fields[0] {
	case "obfs-local", "simpleobfs":
		sip003.Name = "simple-obfs"
	default:
		sip003.Name = fields[0]
	}
	sip003.Opts = ParseSip003Opts(fields[1])
	return sip003
}

func (s *Sip003) String() string {
	list := []string{s.Name}
	switch s.Name {
	case "simple-obfs":
		if s.Opts.Obfs != "" {
			list = append(list, "obfs="+s.Opts.Obfs)
		}
		if s.Opts.Host != "" {
			list = append(list, "obfs-host="+s.Opts.Host)
		}
		if s.Opts.Path != "" {
			list = append(list, "obfs-uri="+s.Opts.Path)
		}
		if s.Opts.Impl != "" {
			list = append(list, "impl="+s.Opts.Impl)
		}
	case "v2ray-plugin":
		if s.Opts.Tls != "" {
			list = append(list, "tls")
		}
		if s.Opts.Obfs != "" {
			list = append(list, "mode="+s.Opts.Obfs)
		}
		if s.Opts.Host != "" {
			list = append(list, "host="+s.Opts.Host)
		}
		if s.Opts.Path != "" {
			list = append(list, "path="+s.Opts.Path)
		}
		if s.Opts.Impl != "" {
			list = append(list, "impl="+s.Opts.Impl)
		}
	}
	return strings.Join(list, ";")
}
