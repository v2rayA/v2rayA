package ssr

import (
	"fmt"
	shadowsocksr "github.com/v2rayA/shadowsocksR"
	"github.com/v2rayA/shadowsocksR/obfs"
	"github.com/v2rayA/shadowsocksR/protocol"
	"github.com/v2rayA/shadowsocksR/ssr"
	"github.com/v2rayA/shadowsocksR/streamCipher"
	"github.com/v2rayA/shadowsocksR/tools/socks"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"net/url"
	"strconv"
)

// SSR struct.
type SSR struct {
	dialer plugin.Dialer
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
	plugin.RegisterDialer("ssr", NewSSRDialer)
}

// NewSSR returns a shadowsocksr proxy, ssr://method:pass@host:port/query
func NewSSR(s string, d plugin.Dialer) (*SSR, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Warn("parse err: %s", err)
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
	p.Protocol = query.Get("proto")
	p.ProtocolParam = query.Get("protoParam")
	p.Obfs = query.Get("obfs")
	p.ObfsParam = query.Get("obfsParam")

	p.ProtocolData = new(protocol.AuthData)

	return p, nil
}

// NewSSRDialer returns a ssr proxy dialer.
func NewSSRDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
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
		return nil, fmt.Errorf("[ssr] unable to parse address: " + addr)
	}

	cipher, err := streamCipher.NewStreamCipher(s.EncryptMethod, s.EncryptPassword)
	if err != nil {
		return nil, err
	}

	c, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		log.Warn("[ssr] dial to %s error: %s", s.addr, err)
		return nil, err
	}

	ssrconn := shadowsocksr.NewSSTCPConn(c, cipher)
	if ssrconn.Conn == nil || ssrconn.RemoteAddr() == nil {
		return nil, fmt.Errorf("[ssr] nil connection")
	}

	// should initialize obfs/protocol now
	h, p, _ := net.SplitHostPort(ssrconn.RemoteAddr().String())
	port, _ := strconv.Atoi(p)

	ssrconn.IObfs = obfs.NewObfs(s.Obfs)
	if ssrconn.IObfs == nil {
		return nil, fmt.Errorf("[ssr] unsupported obfs type: " + s.Obfs)
	}

	obfsServerInfo := &ssr.ServerInfo{
		Host:   h,
		Port:   uint16(port),
		TcpMss: 1460,
		Param:  s.ObfsParam,
	}
	ssrconn.IObfs.SetServerInfo(obfsServerInfo)

	ssrconn.IProtocol = protocol.NewProtocol(s.Protocol)
	if ssrconn.IProtocol == nil {
		return nil, fmt.Errorf("[ssr] unsupported protocol type: " + s.Protocol)
	}

	protocolServerInfo := &ssr.ServerInfo{
		Host:   h,
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
	return nil, nil, fmt.Errorf("[ssr] udp not supported now")
}
