package trojango

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/plugin/infra/dialer2proxy"
	"github.com/v2rayA/v2rayA/plugin/infra/ss"
	"github.com/v2rayA/v2rayA/plugin/infra/trojanc"
	"github.com/v2rayA/v2rayA/plugin/infra/ws"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/plugin"
	"github.com/v2rayA/v2rayA/plugin/infra"
	"github.com/v2rayA/v2rayA/plugin/infra/tls"
	"net"
	"net/url"
	"strings"
)

type TrojanGo struct {
	s *plugin.Server
}
type Params struct {
	Cipher, Passwd, Address, Port, Obfs, ObfsParam, Protocol, ProtocolParam string
}

func init() {
	plugin.RegisterPlugin("trojan-go", NewTrojanGo)
}

func NewTrojanGo(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(TrojanGo)
	err = plugin.Serve(localPort, v)
	return
}

func (self *TrojanGo) LocalPort() int {
	return self.s.LocalPort
}
func (self *TrojanGo) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	fields := strings.SplitN(v.Host, ",", 2)
	sni := fields[0]
	host := fields[1]
	self.s = plugin.NewServer(localPort)
	// tls -> ws -> ss -> trojanc
	var (
		dialer infra.Dialer
		u      url.URL
	)
	dialer = &infra.Direct{}
	if v.TLS == "tls" {
		query := url.Values{}
		query.Add("host", sni)
		query.Add("skipVerify", common.BoolToString(v.AllowInsecure))
		u = url.URL{
			Scheme:   "tls",
			Host:     net.JoinHostPort(v.Add, v.Port),
			RawQuery: query.Encode(),
		}
		dialer, err = tls.NewTls(u.String(), dialer)
		if err != nil {
			return fmt.Errorf("[trojango] failed for NewTls: %w", err)
		}
	}
	if v.Net == "ws" || v.Net == "websocket" {
		query := url.Values{}
		query.Add("host", host)
		query.Add("path", v.Path)
		u = url.URL{
			Scheme:   "ws",
			Host:     net.JoinHostPort(v.Add, v.Port),
			RawQuery: query.Encode(),
		}
		dialer, err = ws.NewWs(u.String(), dialer)
		if err != nil {
			return fmt.Errorf("[trojango] failed for NewWs: %w", err)
		}
	}
	if strings.HasPrefix(v.Type, "ss;") {
		fields := strings.SplitN(v.Type, ";", 3)
		u = url.URL{
			Scheme: "ss",
			Host:   net.JoinHostPort(v.Add, v.Port),
			User:   url.UserPassword(fields[1], fields[2]),
		}
		dialer, err = ss.NewShadowsocks(u.String(), dialer)
		if err != nil {
			return fmt.Errorf("[trojango] failed for NewShadowsocks: %w", err)
		}
	}
	u = url.URL{
		Scheme: "trojanc",
		User:   url.User(v.ID),
		Host:   net.JoinHostPort(v.Add, v.Port),
	}
	dialer, err = trojanc.NewTrojanc(u.String(), dialer)
	if err != nil {
		return fmt.Errorf("[trojango] failed for NewTrojanc: %w", err)
	}
	return self.s.Serve(dialer2proxy.From(dialer, net.JoinHostPort(v.Add, v.Port)), "socks5")
}

func (self *TrojanGo) Close() error {
	return self.s.Close()
}

func (self *TrojanGo) SupportUDP() bool {
	return false
}
