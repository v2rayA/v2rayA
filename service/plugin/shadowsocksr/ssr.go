package shadowsocksr

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/plugin"
	"github.com/v2rayA/v2rayA/plugin/infra/ssr"
	"log"
	"net"
	"net/url"
	"strings"
)

type SSR struct {
	s *plugin.Server
}
type Params struct {
	Cipher, Passwd, Address, Port, Obfs, ObfsParam, Protocol, ProtocolParam string
}

func init() {
	plugin.RegisterPlugin("ssr", NewSSRPlugin)
	plugin.RegisterPlugin("shadowsocksr", NewSSRPlugin)
}

func (self *SSR) LocalPort() int {
	return self.s.LocalPort
}
func NewSSRPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(SSR)
	err = plugin.Serve(localPort, v)
	return
}

func ParseVmess(v vmessInfo.VmessInfo) (s string, err error) {
	params := Params{
		Cipher:        v.Net,
		Passwd:        v.ID,
		Address:       v.Add,
		Port:          v.Port,
		Obfs:          v.TLS,
		ObfsParam:     v.Path,
		Protocol:      v.Type,
		ProtocolParam: v.Host,
	}
	u, err := url.Parse(fmt.Sprintf(
		"ssr://%v:%v@%v",
		url.PathEscape(params.Cipher),
		url.PathEscape(params.Passwd),
		net.JoinHostPort(params.Address, params.Port),
	))
	if err != nil {
		log.Println(err)
		return
	}
	q := u.Query()
	if len(strings.TrimSpace(params.Obfs)) <= 0 {
		params.Obfs = "plain"
	}
	if len(strings.TrimSpace(params.Protocol)) <= 0 {
		params.Protocol = "origin"
	}
	q.Set("obfs", params.Obfs)
	q.Set("obfs_param", params.ObfsParam)
	q.Set("protocol", params.Protocol)
	q.Set("protocol_param", params.ProtocolParam)
	u.RawQuery = q.Encode()
	s = u.String()
	return
}

func (self *SSR) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	self.s = plugin.NewServer(localPort)
	s, err := ParseVmess(v)
	if err != nil {
		return
	}
	p, _ := ssr.NewProxy(s)
	return self.s.Serve(p, "socks5")
}

func (self *SSR) Close() error {
	return self.s.Close()
}

func (self *SSR) SupportUDP() bool {
	return false
}
