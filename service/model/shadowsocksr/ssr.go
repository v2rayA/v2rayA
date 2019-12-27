package shadowsocksr

import (
	"V2RayA/extra/proxy/socks5"
	"V2RayA/extra/proxy/ssr"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

type SSR struct {
	c chan struct{}
}

func (self *SSR) Serve(localPort int, cipher, passwd, address, port, obfs, obfsParam, protocol, protocolParam string) (err error) {
	self.c = make(chan struct{}, 0)
	u, _ := url.Parse(fmt.Sprintf("ssr://%v:%v@%v:%v", cipher, passwd, address, port))
	q := u.Query()
	q.Set("obfs", obfs)
	q.Set("obfs_param", obfsParam)
	q.Set("protocol", protocol)
	q.Set("protocol_param", protocolParam)
	u.RawQuery = q.Encode()
	p, _ := ssr.NewProxy(u.String())
	local, err := socks5.NewSocks5Server("socks5://127.0.0.1:"+strconv.Itoa(localPort), p)
	if err != nil {
		return
	}
	go func() {
		go local.ListenAndServe()
		<-self.c
		_ = local.(*socks5.Socks5).TcpListener.Close()
	}()
	return nil
}

func (self *SSR) Close() error {
	if self.c == nil {
		return errors.New("close fail: shadowsocksr not running")
	}
	if len(self.c) > 0 {
		return errors.New("close fail: duplicate close")
	}
	self.c <- struct{}{}
	self.c = nil
	return nil
}

func (self *SSR) IsRunning() bool {
	return self.c != nil
}
