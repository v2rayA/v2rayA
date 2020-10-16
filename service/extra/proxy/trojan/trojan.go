// from https://github.com/nadoo/glider/blob/master/proxy/trojan/trojan.go

// protocol spec:
// https://trojan-gfw.github.io/trojan/protocol

package trojan

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/nadoo/glider/common/socks"
	"net"
	"net/url"
	"strings"
	"github.com/v2rayA/v2rayA/extra/proxy"
)

// Trojan is a base trojan struct
type Trojan struct {
	dialer     proxy.Dialer
	proxy      proxy.Proxy
	addr       string
	pass       [56]byte
	serverName string
	skipVerify bool
	tlsConfig  *tls.Config
}

// NewTrojan returns a trojan proxy.
func NewTrojan(s string, d proxy.Dialer, p proxy.Proxy) (*Trojan, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, newError("[trojan]").Base(err)
	}

	t := &Trojan{
		dialer: d,
		proxy:  p,
		addr:   u.Host,
	}

	// pass
	hash := sha256.New224()
	hash.Write([]byte(u.User.Username()))
	hex.Encode(t.pass[:], hash.Sum(nil))

	query := u.Query()
	t.serverName = query.Get("peer")
	if t.serverName == "" {
		colonPos := strings.LastIndex(t.addr, ":")
		if colonPos == -1 {
			colonPos = len(t.addr)
		}
		t.serverName = t.addr[:colonPos]
	}

	// skipVerify
	if query.Get("skipVerify") == "true" {
		t.skipVerify = true
	}

	t.tlsConfig = &tls.Config{
		ServerName:         t.serverName,
		InsecureSkipVerify: t.skipVerify,
		NextProtos:         []string{"http/1.1", "h2"},
		ClientSessionCache: tls.NewLRUClientSessionCache(64),
		MinVersion:         tls.VersionTLS10,
	}

	return t, nil
}

// Addr returns forwarder's address.
func (s *Trojan) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the proxy.
func (s *Trojan) Dial(network, addr string) (net.Conn, error) {
	return s.dial(network, addr)
}

func (s *Trojan) dial(network, addr string) (net.Conn, error) {
	rc, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		return nil, newError(fmt.Sprintf("[trojan]: dial to %s", s.addr)).Base(err)
	}

	tlsConn := tls.Client(rc, s.tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write(s.pass[:])
	buf.WriteString("\r\n")

	cmd := socks.CmdConnect
	if network == "udp" {
		cmd = socks.CmdUDPAssociate
	}
	buf.WriteByte(byte(cmd))

	buf.Write(socks.ParseAddr(addr))
	buf.WriteString("\r\n")
	_, err = tlsConn.Write(buf.Bytes())

	return tlsConn, err
}

// DialUDP connects to the given address via the proxy.
func (s *Trojan) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	//TODO
	return nil, nil, nil
}
