// from https://github.com/nadoo/glider/blob/master/proxy/trojan/trojan.go

// protocol spec:
// https://trojan-gfw.github.io/trojan/protocol

package trojanc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"

	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/plugin/socks"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Trojan is a base trojan struct
type Trojan struct {
	dialer plugin.Dialer
	addr   string
	pass   [56]byte
}

func init() {
	log.Trace("[trojanc] registering dialer")
	plugin.RegisterDialer("trojanc", NewTrojancDialer)
}

// NewTrojanc returns a trojan-cleartext infra.
func NewTrojanc(s string, d plugin.Dialer) (*Trojan, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("NewTrojanc: %w", err)
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

func NewTrojancDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	return NewTrojanc(s, d)
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
	return s.DialContext(context.Background(), network, addr)
}

func (s *Trojan) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	log.Info("[trojanc] dialing %s via %s", addr, s.addr)
	rc, err := s.dialer.DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return nil, fmt.Errorf("[trojan]: dial to %s: %w", s.addr, err)
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
func (s *Trojan) DialUDP(network string) (pc plugin.FakeNetPacketConn, err error) {
	//TODO
	return nil, nil
}
