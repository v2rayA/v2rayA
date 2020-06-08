package ssr

import (
	shadowsocksr "github.com/mzz2017/shadowsocksR"
	"github.com/mzz2017/shadowsocksR/ssr"
	"github.com/mzz2017/shadowsocksR/streamCipher"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/mzz2017/shadowsocksR/obfs"
	"github.com/mzz2017/shadowsocksR/protocol"
	"github.com/mzz2017/v2rayA/extra/proxy"
	"github.com/nadoo/glider/common/socks"
	"log"
)

// SSR struct.
type SSR struct {
	dialer proxy.Dialer
	addr   string

	EncryptMethod   string
	EncryptPassword string
	Obfs            string
	ObfsParam       string
	ObfsData        interface{}
	Protocol        string
	ProtocolParam   string
	ProtocolData    interface{}
}

func init() {
	proxy.RegisterDialer("ssr", NewSSRDialer)
}

// NewSSR returns a shadowsocksr proxy, ssr://method:pass@host:port/query
func NewSSR(s string, d proxy.Dialer) (*SSR, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Printf("parse err: %s\n", err)
		return nil, err
	}

	addr := u.Host
	method := u.User.Username()
	pass, _ := u.User.Password()

	p := &SSR{
		dialer:          d,
		addr:            addr,
		EncryptMethod:   method,
		EncryptPassword: pass,
	}

	query := u.Query()
	p.Protocol = query.Get("protocol")
	p.ProtocolParam = query.Get("protocol_param")
	p.Obfs = query.Get("obfs")
	p.ObfsParam = query.Get("obfs_param")

	p.ProtocolData = new(protocol.AuthData)

	return p, nil
}

// NewSSRDialer returns a ssr proxy dialer.
func NewSSRDialer(s string, d proxy.Dialer) (proxy.Dialer, error) {
	return NewSSR(s, d)
}

// Addr returns forwarder's address
func (s *SSR) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the proxy.
func (s *SSR) Dial(network, addr string) (net.Conn, error) {
	target := socks.ParseAddr(addr)
	if target == nil {
		return nil, newError("[ssr] unable to parse address: " + addr)
	}

	cipher, err := streamCipher.NewStreamCipher(s.EncryptMethod, s.EncryptPassword)
	if err != nil {
		return nil, err
	}

	c, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		log.Printf("[ssr] dial to %s error: %s\n", s.addr, err)
		return nil, err
	}

	ssrconn := shadowsocksr.NewSSTCPConn(c, cipher)
	if ssrconn.Conn == nil || ssrconn.RemoteAddr() == nil {
		return nil, newError("[ssr] nil connection")
	}

	// should initialize obfs/protocol now
	rs := strings.Split(ssrconn.RemoteAddr().String(), ":")
	port, _ := strconv.Atoi(rs[1])

	ssrconn.IObfs = obfs.NewObfs(s.Obfs)
	if ssrconn.IObfs == nil {
		return nil, newError("[ssr] unsupported obfs type: " + s.Obfs)
	}

	obfsServerInfo := &ssr.ServerInfoForObfs{
		Host:   rs[0],
		Port:   uint16(port),
		TcpMss: 1460,
		Param:  s.ObfsParam,
	}
	ssrconn.IObfs.SetServerInfo(obfsServerInfo)

	ssrconn.IProtocol = protocol.NewProtocol(s.Protocol)
	if ssrconn.IProtocol == nil {
		return nil, newError("[ssr] unsupported protocol type: " + s.Protocol)
	}

	protocolServerInfo := &ssr.ServerInfoForObfs{
		Host:   rs[0],
		Port:   uint16(port),
		TcpMss: 1460,
		Param:  s.ProtocolParam,
	}
	ssrconn.IProtocol.SetServerInfo(protocolServerInfo)

	if s.ObfsData == nil {
		s.ObfsData = ssrconn.IObfs.GetData()
	}
	ssrconn.IObfs.SetData(s.ObfsData)

	if s.ProtocolData == nil {
		s.ProtocolData = ssrconn.IProtocol.GetData()
	}
	ssrconn.IProtocol.SetData(s.ProtocolData)

	if _, err := ssrconn.Write(target); err != nil {
		ssrconn.Close()
		return nil, err
	}

	return ssrconn, err
}

// DialUDP connects to the given address via the proxy.
func (s *SSR) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	return nil, nil, newError("[ssr] udp not supported now")
}
