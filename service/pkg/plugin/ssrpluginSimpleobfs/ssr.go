package ssrpluginSimpleobfs

import (
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/plugin/infra"
	"github.com/v2rayA/v2rayA/pkg/plugin/infra/simpleobfs"
	"github.com/v2rayA/v2rayA/pkg/plugin/infra/ssr"
	"github.com/v2rayA/v2rayA/pkg/plugin/shadowsocksr"
	simpleobfs2 "github.com/v2rayA/v2rayA/pkg/plugin/simpleobfs"
)

type SSRSimpleobfs struct {
	s *plugin.Server
}
type Params struct {
	Cipher, Passwd, Address, Port, Obfs, ObfsParam, Protocol, ProtocolParam string
}

func init() {
	plugin.RegisterPlugin("ssrplugin-simpleobfs", NewSsrSimpleobfsPlugin)
}

func (self *SSRSimpleobfs) LocalPort() int {
	return self.s.LocalPort
}

func NewSsrSimpleobfsPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(SSRSimpleobfs)
	err = plugin.Serve(localPort, v)
	return
}

func (self *SSRSimpleobfs) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	self.s = plugin.NewServer(localPort)
	ss := v
	plugin.RazorSS(&ss)
	sss, _ := shadowsocksr.ParseVmess(ss)
	sos, _ := simpleobfs2.ParseVmess(v)
	d, _ := infra.NewDirect("")
	simpleobfsDialer, err := simpleobfs.NewSimpleObfsDialer(sos, d)
	if err != nil {
		return
	}
	dialer, err := ssr.NewSSR(sss, simpleobfsDialer)
	if err != nil {
		return
	}
	return self.s.Serve(ssr.Proxy{SSR: *dialer}, "socks5")
}

func (self *SSRSimpleobfs) Close() error {
	return self.s.Close()
}

func (self *SSRSimpleobfs) SupportUDP() bool {
	return false
}
