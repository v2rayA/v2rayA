package shadowsocksr

import (
	"V2RayA/common/netTools/ports"
	"V2RayA/core/vmessInfo"
	"V2RayA/extra/proxy/socks5"
	"V2RayA/extra/proxy/ssr"
	"V2RayA/plugins"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SSR struct {
	c         chan struct{}
	localPort int
}
type Params struct {
	Cipher, Passwd, Address, Port, Obfs, ObfsParam, Protocol, ProtocolParam string
}

func init() {
	plugins.RegisterPlugin("ss", newSSRPlugin)
	plugins.RegisterPlugin("ssr", newSSRPlugin)
	plugins.RegisterPlugin("shadowsocks", newSSRPlugin)
	plugins.RegisterPlugin("shadowsocksr", newSSRPlugin)
}

func newSSRPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugins.Plugin, err error) {
	plugin = new(SSR)
	err = plugin.Serve(localPort, v)
	return
}

func (self *SSR) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	self.c = make(chan struct{}, 0)
	self.localPort = localPort
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
		"ssr://%v:%v@%v:%v",
		url.PathEscape(params.Cipher),
		url.PathEscape(params.Passwd),
		url.PathEscape(params.Address),
		url.PathEscape(params.Port),
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
	p, _ := ssr.NewProxy(u.String())
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
		<-self.c
		if local.(*socks5.Socks5).TcpListener != nil {
			_ = local.(*socks5.Socks5).TcpListener.Close()
		}
	}()
	//等待100ms的error
	time.Sleep(100 * time.Millisecond)
	return err
}

func (self *SSR) Close() error {
	if self.c == nil {
		return newError("close fail: shadowsocksr not running")
	}
	if len(self.c) > 0 {
		return newError("close fail: duplicate close")
	}
	self.c <- struct{}{}
	self.c = nil
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	for {
		var o bool
		o, _, err := ports.IsPortOccupied([]string{strconv.Itoa(self.localPort) + ":tcp"})
		if err != nil {
			return err
		}
		if !o {
			break
		}
		if time.Since(start) > 5*time.Second {
			log.Println("SSR.Close: timeout", self.localPort)
			return newError("SSR.Close: timeout")
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (self *SSR) SupportUDP() bool {
	return false
}
