// protocol spec:
// https://trojan-gfw.github.io/trojan/protocol

package trojan

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/extra/proxy/trojan"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
	"net/url"
)

type Trojan struct {
	s *plugin.Server
}
type Params struct {
	Address    string
	Port       string
	Passwd     string
	SkipVerify bool
	Peer       string
}

func init() {
	plugin.RegisterPlugin("trojan", NewTrojanPlugin)
}

func NewTrojanPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(Trojan)
	err = plugin.Serve(localPort, v)
	return
}

func (tro *Trojan) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	tro.s = plugin.NewServer(localPort)
	params := Params{
		Passwd:     v.ID,
		Address:    v.Add,
		Port:       v.Port,
		SkipVerify: v.AllowInsecure,
		Peer:       v.Host,
	}
	u, err := url.Parse(fmt.Sprintf(
		"trojan://%v@%v:%v",
		url.PathEscape(params.Passwd),
		url.PathEscape(params.Address),
		url.PathEscape(params.Port),
	))
	if err != nil {
		log.Println(err)
		return
	}
	q := u.Query()
	q.Set("skipVerify", common.BoolToString(params.SkipVerify))
	q.Set("peer", params.Peer)
	u.RawQuery = q.Encode()
	p, _ := trojan.NewProxy(u.String())
	return tro.s.Serve(p, "socks5")
}

func (tro *Trojan) Close() error {
	return tro.s.Close()
}

func (tro *Trojan) SupportUDP() bool {
	return false
}
