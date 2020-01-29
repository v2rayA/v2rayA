package shadowsocksr

import (
	"V2RayA/extra/proxy/socks5"
	"V2RayA/extra/proxy/ssr"
	"V2RayA/tools/ports"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SSR struct {
	c    chan struct{}
	port int
}

func (self *SSR) Serve(localPort int, cipher, passwd, address, port, obfs, obfsParam, protocol, protocolParam string) (err error) {
	self.c = make(chan struct{}, 0)

	u, err := url.Parse(fmt.Sprintf(
		"ssr://%v:%v@%v:%v",
		url.PathEscape(cipher),
		url.PathEscape(passwd),
		url.PathEscape(address),
		url.PathEscape(port),
	))
	if err != nil {
		log.Println(err)
		return
	}
	q := u.Query()
	if len(strings.TrimSpace(obfs)) <= 0 {
		obfs = "plain"
	}
	if len(strings.TrimSpace(protocol)) <= 0 {
		protocol = "origin"
	}
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
	self.port = localPort
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
		return errors.New("close fail: shadowsocksr not running")
	}
	if len(self.c) > 0 {
		return errors.New("close fail: duplicate close")
	}
	self.c <- struct{}{}
	self.c = nil
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	port := strconv.Itoa(self.port)
	for {
		o, who := ports.IsPortOccupied(port, "tcp", true)
		if !o {
			break
		}
		if time.Since(start) > 3*time.Second {
			log.Println("SSR.Close: 关闭SSR超时", port+"/"+who)
			return errors.New("SSR.Close: 关闭SSR超时")
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (self *SSR) IsRunning() bool {
	return self.c != nil
}
