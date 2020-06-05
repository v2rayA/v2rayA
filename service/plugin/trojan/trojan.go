// protocol spec:
// https://trojan-gfw.github.io/trojan/protocol

package trojan

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"
	"v2rayA/common"
	"v2rayA/common/netTools/ports"
	"v2rayA/core/vmessInfo"
	"v2rayA/extra/proxy/socks5"
	"v2rayA/extra/proxy/trojan"
	"v2rayA/plugin"
)

type Trojan struct {
	c         chan struct{}
	localPort int
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
	tro.c = make(chan struct{}, 0)
	tro.localPort = localPort
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
	local, err := socks5.NewSocks5Server("socks5://127.0.0.1:"+strconv.Itoa(localPort), p)
	if err != nil {
		return
	}
	go func() {
		go func() {
			e := local.ListenAndServe()
			if e != nil {
				err = e
			}
		}()
		<-tro.c
		if local.(*socks5.Socks5).TcpListener != nil {
			_ = local.(*socks5.Socks5).TcpListener.Close()
		}
	}()
	//等待100ms的error
	time.Sleep(100 * time.Millisecond)
	return err
}

func (tro *Trojan) Close() error {
	if tro.c == nil {
		return newError("close fail: trojan not running")
	}
	if len(tro.c) > 0 {
		return newError("close fail: duplicate close")
	}
	tro.c <- struct{}{}
	tro.c = nil
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	for {
		var o bool
		o, _, err := ports.IsPortOccupied([]string{strconv.Itoa(tro.localPort) + ":tcp"})
		if err != nil {
			return err
		}
		if !o {
			break
		}
		if time.Since(start) > 5*time.Second {
			log.Println("Trojan.Close: timeout", tro.localPort)
			return newError("Trojan.Close: timeout")
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (tro *Trojan) SupportUDP() bool {
	return false
}
