// from https://github.com/nadoo/glider/blob/master/proxy/trojan/trojan.go

// protocol spec:
// https://trojan-gfw.github.io/trojan/protocol

package trojanc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/v2rayA/v2rayA/plugin/infra"
	"github.com/v2rayA/v2rayA/plugin/infra/socks"
	"net"
	"net/url"
)

// Trojan is a base trojan struct
type Trojan struct {
	dialer infra.Dialer
	addr   string
	pass   [56]byte
}

// NewTrojanc returns a trojan-cleartext infra.
func NewTrojanc(s string, d infra.Dialer) (*Trojan, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, newError("[trojan]").Base(err)
	}

	t := &Trojan{
		dialer: d,
		addr:   u.Host,
	}

	// pass
	hash := sha256.New224()
	hash.Write([]byte(u.User.Username()))
	hex.Encode(t.pass[:], hash.Sum(nil))

	return t, nil
}

// Addr returns forwarder's address.
func (s *Trojan) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the infra.
func (s *Trojan) Dial(network, addr string) (net.Conn, error) {
	return s.dial(network, addr)
}

func (s *Trojan) dial(network, addr string) (net.Conn, error) {
	rc, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		return nil, newError(fmt.Sprintf("[trojan]: dial to %s", s.addr)).Base(err)
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
	_, err = rc.Write(buf.Bytes())
	return rc, err
}

// DialUDP connects to the given address via the infra.
func (s *Trojan) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	//TODO
	return nil, nil, nil
}
