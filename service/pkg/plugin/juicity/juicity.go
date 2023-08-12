package juicity

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/protocol"
	"github.com/daeuniverse/softwind/protocol/juicity"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/plugin/internal/common"
)

// Juicity is a base juicity struct
type Juicity struct {
	dialer netproxy.Dialer

	Server                string
	Port                  int
	User                  string
	Password              string
	Sni                   string
	AllowInsecure         bool
	CongestionControl     string
	PinnedCertchainSha256 string
}

func init() {
	plugin.RegisterDialer("juicity", NewJuicityDialer)
}

func ParseJuicityURL(u string) (data *Juicity, err error) {
	t, err := url.Parse(u)
	if err != nil {
		err = fmt.Errorf("invalid trojan format")
		return
	}
	allowInsecure, _ := strconv.ParseBool(t.Query().Get("allowInsecure"))
	if !allowInsecure {
		allowInsecure, _ = strconv.ParseBool(t.Query().Get("allow_insecure"))
	}
	if !allowInsecure {
		allowInsecure, _ = strconv.ParseBool(t.Query().Get("allowinsecure"))
	}
	if !allowInsecure {
		allowInsecure, _ = strconv.ParseBool(t.Query().Get("skipVerify"))
	}
	sni := t.Query().Get("peer")
	if sni == "" {
		sni = t.Query().Get("sni")
	}
	if sni == "" {
		sni = t.Hostname()
	}
	port, err := strconv.Atoi(t.Port())
	if err != nil {
		return nil, fmt.Errorf("parse port: %w", err)
	}
	password, _ := t.User.Password()
	data = &Juicity{
		Server:                t.Hostname(),
		Port:                  port,
		User:                  t.User.Username(),
		Password:              password,
		Sni:                   sni,
		AllowInsecure:         allowInsecure,
		CongestionControl:     t.Query().Get("congestion_control"),
		PinnedCertchainSha256: t.Query().Get("pinned_certchain_sha256"),
	}
	return data, nil
}

// NewJuicity returns a juicity infra.
func NewJuicity(u string, d plugin.Dialer) (*Juicity, error) {
	s, err := ParseJuicityURL(u)
	if err != nil {
		return nil, err
	}
	var flags protocol.Flags
	tlsConfig := &tls.Config{
		NextProtos:         []string{"h3"},
		MinVersion:         tls.VersionTLS13,
		ServerName:         s.Sni,
		InsecureSkipVerify: s.AllowInsecure,
	}
	if s.PinnedCertchainSha256 != "" {
		pinnedHash, err := base64.URLEncoding.DecodeString(s.PinnedCertchainSha256)
		if err != nil {
			pinnedHash, err = base64.StdEncoding.DecodeString(s.PinnedCertchainSha256)
			if err != nil {
				pinnedHash, err = hex.DecodeString(s.PinnedCertchainSha256)
				if err != nil {
					return nil, fmt.Errorf("failed to decode PinnedCertchainSha256")
				}
			}
		}
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if !bytes.Equal(common.GenerateCertChainHash(rawCerts), pinnedHash) {
				return fmt.Errorf("pinned hash of cert chain does not match")
			}
			return nil
		}
	}
	if s.dialer, err = juicity.NewDialer(&plugin.Converter{
		Dialer: d,
	}, protocol.Header{
		ProxyAddress: net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		Feature1:     s.CongestionControl,
		TlsConfig:    tlsConfig,
		User:         s.User,
		Password:     s.Password,
		IsClient:     true,
		Flags:        flags,
	}); err != nil {
		return nil, err
	}
	return s, nil
}

func NewJuicityDialer(s string, d plugin.Dialer) (plugin.Dialer, error) {
	return NewJuicity(s, d)
}

// Addr returns forwarder's address.
func (s *Juicity) Addr() string {
	return ""
}

// Dial connects to the address addr on the network net via the infra.
func (s *Juicity) Dial(network, addr string) (net.Conn, error) {
	return s.dial(network, addr)
}

func (s *Juicity) dial(network, addr string) (net.Conn, error) {
	rc, err := s.dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[juicity]: dial to %s: %w", addr, err)
	}
	return &netproxy.FakeNetConn{
		Conn:  rc,
		LAddr: nil,
		RAddr: nil,
	}, err
}

// DialUDP connects to the given address via the infra.
func (s *Juicity) DialUDP(network, addr string) (net.PacketConn, net.Addr, error) {
	rc, err := s.dialer.Dial("udp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("[juicity]: dial to %s: %w", addr, err)
	}
	return &netproxy.FakeNetPacketConn{
		PacketConn: rc.(netproxy.PacketConn),
		LAddr:      nil,
		RAddr:      nil,
	}, nil, err
}
