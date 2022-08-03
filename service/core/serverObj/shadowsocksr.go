package serverObj

import (
	"encoding/base64"
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	FromLinkRegister("shadowsocksr", NewShadowsocksR)
	FromLinkRegister("ssr", NewShadowsocksR)
	EmptyRegister("shadowsocksr", func() (ServerObj, error) {
		return new(ShadowsocksR), nil
	})
	EmptyRegister("ssr", func() (ServerObj, error) {
		return new(ShadowsocksR), nil
	})
}

type ShadowsocksR struct {
	Name       string `json:"name"`
	Server     string `json:"server"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
	Cipher     string `json:"cipher"`
	Proto      string `json:"proto"`
	ProtoParam string `json:"protoParam"`
	Obfs       string `json:"obfs"`
	ObfsParam  string `json:"obfsParam"`
	Protocol   string `json:"protocol"`
}

func NewShadowsocksR(link string) (ServerObj, error) {
	return ParseSSRURL(link)
}

func ParseSSRURL(u string) (data *ShadowsocksR, err error) {
	// parse attempts to parse ss:// links
	parse := func(content string) (v ShadowsocksR, ok bool) {
		arr := strings.Split(content, "/?")
		if strings.Contains(content, ":") && len(arr) < 2 {
			content += "/?remarks=&protoparam=&obfsparam="
			arr = strings.Split(content, "/?")
		} else if len(arr) != 2 {
			return v, false
		}
		pre := strings.Split(arr[0], ":")
		if len(pre) > 6 {
			//if the length is more than 6, it means that the host contains the characters:,
			//re-merge the first few groups into the host
			pre[len(pre)-6] = strings.Join(pre[:len(pre)-5], ":")
			pre = pre[len(pre)-6:]
		} else if len(pre) < 6 {
			return v, false
		}
		q, err := url.ParseQuery(arr[1])
		if err != nil {
			return v, false
		}
		pswd, _ := common.Base64URLDecode(pre[5])
		add, _ := common.Base64URLDecode(pre[0])
		remarks, _ := common.Base64URLDecode(q.Get("remarks"))
		protoparam, _ := common.Base64URLDecode(q.Get("protoparam"))
		obfsparam, _ := common.Base64URLDecode(q.Get("obfsparam"))
		port, err := strconv.Atoi(pre[1])
		if err != nil {
			return v, false
		}
		v = ShadowsocksR{
			Name:       remarks,
			Server:     add,
			Port:       port,
			Password:   pswd,
			Cipher:     pre[3],
			Proto:      pre[2],
			ProtoParam: protoparam,
			Obfs:       pre[4],
			ObfsParam:  obfsparam,
			Protocol:   "shadowsocksr",
		}
		return v, true
	}
	content := u[6:]
	var (
		info ShadowsocksR
		ok   bool
	)
	// try parsing the ssr:// link, if it fails, base64 decode first
	if info, ok = parse(content); !ok {
		// perform base64 decoding and parse again
		content, err = common.Base64StdDecode(content)
		if err != nil {
			content, err = common.Base64URLDecode(content)
			if err != nil {
				return
			}
		}
		info, ok = parse(content)
	}
	if !ok {
		err = fmt.Errorf("%w: unrecognized ssr address", InvalidParameterErr)
		return
	}
	return &info, nil
}

func (s *ShadowsocksR) Configuration(info PriorInfo) (c Configuration, err error) {
	socks5 := url.URL{
		Scheme: "socks5",
		Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(info.PluginPort)),
	}
	ssr := url.URL{
		Scheme: "ssr",
		Host:   net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		User:   url.UserPassword(s.Cipher, s.Password),
		RawQuery: url.Values{
			"proto":      []string{s.Proto},
			"protoParam": []string{s.ProtoParam},
			"obfs":       []string{s.Obfs},
			"obfsParam":  []string{s.ObfsParam},
		}.Encode(),
	}
	chain := []string{socks5.String(), ssr.String()}
	return Configuration{
		CoreOutbound: info.PluginObj(),
		PluginChain:  strings.Join(chain, ","),
		UDPSupport:   false,
	}, nil
}

func (s *ShadowsocksR) ExportToURL() string {
	/* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
	return fmt.Sprintf("ssr://%v", strings.TrimSuffix(base64.URLEncoding.EncodeToString([]byte(
		fmt.Sprintf(
			"%v:%v:%v:%v:%v/?remarks=%v&protoparam=%v&obfsparam=%v",
			net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
			s.Proto,
			s.Cipher,
			s.Obfs,
			base64.URLEncoding.EncodeToString([]byte(s.Password)),
			base64.URLEncoding.EncodeToString([]byte(s.Name)),
			base64.URLEncoding.EncodeToString([]byte(s.ProtoParam)),
			base64.URLEncoding.EncodeToString([]byte(s.ObfsParam)),
		),
	)), "="))
}

func (s *ShadowsocksR) NeedPluginPort() bool {
	return true
}

func (s *ShadowsocksR) ProtoToShow() string {
	obfs := s.Obfs
	if obfs == "tls1.2_ticket_auth" {
		obfs = "tls1.2"
	}
	return fmt.Sprintf("SSR(%v+%v)", s.Proto, obfs)
}

func (s *ShadowsocksR) GetProtocol() string {
	return s.Protocol
}

func (s *ShadowsocksR) GetHostname() string {
	return s.Server
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
