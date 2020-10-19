package plugin

import (
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/extra/proxy"
	"github.com/v2rayA/v2rayA/extra/proxy/socks5"
	"github.com/v2rayA/v2rayA/extra/proxy/tcp"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	C         chan interface{}
	LocalPort int
	closed    chan interface{}
}

func NewServer(localPort int) *Server {
	s := new(Server)
	s.C = make(chan interface{}, 0)
	s.closed = make(chan interface{}, 0)
	s.LocalPort = localPort
	return s
}

// protocol:
// socks5
// tcp->192.168.0.5:80
func (s *Server) Serve(p proxy.Proxy, protocol string) error {
	var local proxy.Server
	var err error
	switch {
	case protocol == "socks5":
		local, err = socks5.NewSocks5Server("socks5://127.0.0.1:"+strconv.Itoa(s.LocalPort), p)
	case strings.HasPrefix(protocol, "tcp"):
		arr := strings.Split(protocol, "->")
		if len(arr) != 2 {
			return newError("func Serve: wrong format of tcp")
		}
		local, err = tcp.NewTcpServer("tcp://127.0.0.1:"+strconv.Itoa(s.LocalPort)+"/?target="+url.PathEscape(arr[1]), p)
	}
	if err != nil {
		return err
	}
	go func() {
		go func() {
			e := local.ListenAndServe()
			if e != nil {
				err = e
			}
		}()
		<-s.C
		if closer, ok := local.(io.Closer); ok {
			close(s.closed)
			_ = closer.Close()
		}
	}()
	//等待100ms的error
	time.Sleep(100 * time.Millisecond)
	return err
}

func (s *Server) Close() error {
	if s.C == nil {
		return newError("close fail: server not running")
	}
	if len(s.C) > 0 {
		return newError("close fail: duplicate close")
	}
	s.C <- nil
	s.C = nil
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	port := strconv.Itoa(s.LocalPort)
out:
	for {
		select {
		case <-s.closed:
			break out
		default:
		}
		var o bool
		o, _, err := ports.IsPortOccupied([]string{port + ":tcp"})
		if err != nil {
			return err
		}
		if !o {
			break
		}
		conn, e := net.Dial("tcp", ":"+port)
		if e == nil {
			conn.Close()
		}
		if time.Since(start) > 3*time.Second {
			log.Println("plugin.Server.Close: timeout", s.LocalPort)
			return newError("Server.Close: timeout")
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}
