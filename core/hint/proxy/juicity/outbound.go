package juicity

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"time"

	outbound_netproxy "github.com/daeuniverse/outbound/netproxy"
	outbound_protocol "github.com/daeuniverse/outbound/protocol"
	_ "github.com/daeuniverse/outbound/protocol/juicity" // register juicity protocol

	"github.com/xtls/xray-core/common"
	xray_buf "github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/errors"
	xray_net "github.com/xtls/xray-core/common/net"
	xray_session "github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/transport"
	"github.com/xtls/xray-core/transport/internet"

	"github.com/v2rayA/v2raya-core/hint/tlsutil"
)

// Client is the juicity outbound handler.
type Client struct {
	config     *ClientConfig
	dialer     outbound_netproxy.Dialer
	newDialer  func(internet.Dialer) (outbound_netproxy.Dialer, error)
	dialerOnce sync.Once
	dialerErr  error
}

// NewClient creates a new juicity outbound handler.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	if config.Address == "" {
		return nil, errors.New("juicity: no server address")
	}
	if config.Uuid == "" {
		return nil, errors.New("juicity: no UUID")
	}

	sni := config.Sni
	if sni == "" {
		// extract host from address (may be host:port or just host)
		host, _, err := net.SplitHostPort(config.Address)
		if err != nil {
			host = config.Address
		}
		sni = host
	}

	tlsCfg := &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: config.AllowInsecure, // #nosec G402 -- user-configurable
		NextProtos:         []string{"h3"},
		MinVersion:         tls.VersionTLS13,
	}
	if config.PinnedCertchainSha256 != "" {
		hash, err := tlsutil.ParsePinnedChain(config.PinnedCertchainSha256)
		if err != nil {
			return nil, fmt.Errorf("juicity: invalid pinned certificate chain hash: %w", err)
		}
		tlsCfg.VerifyPeerCertificate = tlsutil.PinVerifier(hash)
		// Pin verification is self-contained against rawCerts, so the chain / host
		// checks can be bypassed. Without this a self-signed cert would always fail.
		tlsCfg.InsecureSkipVerify = true // #nosec G402 -- guarded by pinning
	}

	congestion := config.CongestionControl
	if congestion == "" {
		congestion = "bbr"
	}

	newDialer := func(dialer internet.Dialer) (outbound_netproxy.Dialer, error) {
		return outbound_protocol.NewDialer("juicity", xrayDialer{dialer: dialer}, outbound_protocol.Header{
			ProxyAddress: config.Address,
			Feature1:     congestion,
			TlsConfig:    tlsCfg,
			User:         config.Uuid,
			Password:     config.Password,
			IsClient:     true,
		})
	}

	return &Client{
		config:    config,
		newDialer: newDialer,
	}, nil
}

type xrayDialer struct {
	dialer internet.Dialer
}

func (d xrayDialer) DialContext(ctx context.Context, network, addr string) (outbound_netproxy.Conn, error) {
	destination, err := xrayDestination(network, addr)
	if err != nil {
		return nil, err
	}
	conn, err := d.dialer.Dial(ctx, destination)
	if err != nil {
		return nil, err
	}
	// QUIC-based protocols assert the underlay conn to netproxy.PacketConn,
	// whose ReadFrom/WriteTo signatures differ from net.PacketConn. xray's UDP
	// dial returns *internet.PacketConnWrapper, so adapt it here.
	if destination.Network == xray_net.Network_UDP {
		if pc, ok := conn.(net.PacketConn); ok {
			return &packetConnAdapter{conn: conn, pc: pc}, nil
		}
		conn.Close()
		return nil, fmt.Errorf("juicity: UDP dial returned %T, which is not a net.PacketConn", conn)
	}
	return conn, nil
}

// packetConnAdapter adapts xray's UDP conn (net.Conn + net.PacketConn) to
// outbound's netproxy.PacketConn interface.
type packetConnAdapter struct {
	conn net.Conn
	pc   net.PacketConn
}

func (a *packetConnAdapter) Read(b []byte) (int, error)  { return a.conn.Read(b) }
func (a *packetConnAdapter) Write(b []byte) (int, error) { return a.conn.Write(b) }
func (a *packetConnAdapter) Close() error                { return a.conn.Close() }
func (a *packetConnAdapter) SetDeadline(t time.Time) error {
	return a.conn.SetDeadline(t)
}
func (a *packetConnAdapter) SetReadDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}
func (a *packetConnAdapter) SetWriteDeadline(t time.Time) error {
	return a.conn.SetWriteDeadline(t)
}

func (a *packetConnAdapter) ReadFrom(p []byte) (int, netip.AddrPort, error) {
	n, addr, err := a.pc.ReadFrom(p)
	if err != nil {
		return n, netip.AddrPort{}, err
	}
	ap, err := netip.ParseAddrPort(addr.String())
	if err != nil {
		return n, netip.AddrPort{}, err
	}
	return n, ap, nil
}

func (a *packetConnAdapter) WriteTo(p []byte, addr string) (int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return 0, err
	}
	return a.pc.WriteTo(p, udpAddr)
}

func xrayDestination(network, addr string) (xray_net.Destination, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return xray_net.Destination{}, fmt.Errorf("juicity: invalid server address %q: %w", addr, err)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return xray_net.Destination{}, fmt.Errorf("juicity: invalid server port %q: %w", port, err)
	}
	address := xray_net.ParseAddress(host)
	mn, err := outbound_netproxy.ParseMagicNetwork(network)
	if err == nil && mn.Network == "udp" {
		return xray_net.UDPDestination(address, xray_net.Port(portNumber)), nil
	}
	return xray_net.TCPDestination(address, xray_net.Port(portNumber)), nil
}

// Process implements proxy.Outbound.
func (c *Client) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	c.dialerOnce.Do(func() {
		c.dialer, c.dialerErr = c.newDialer(dialer)
	})
	if c.dialerErr != nil {
		return errors.New("juicity: failed to create dialer").Base(c.dialerErr)
	}
	outbounds := xray_session.OutboundsFromContext(ctx)
	ob := outbounds[len(outbounds)-1]
	if !ob.Target.IsValid() {
		return errors.New("target not specified")
	}
	destination := ob.Target

	// Build destination address string
	destAddr := fmt.Sprintf("%s:%d", destination.Address.String(), destination.Port.Value())

	conn, err := c.dialer.DialContext(ctx, "tcp", destAddr)
	if err != nil {
		return errors.New("juicity: failed to dial destination").Base(err)
	}
	defer conn.Close()

	postRequest := func() error {
		return xray_buf.Copy(link.Reader, xray_buf.NewWriter(outboundConnWriter(conn)))
	}
	getResponse := func() error {
		return xray_buf.Copy(xray_buf.NewReader(outboundConnReader(conn)), link.Writer)
	}

	responseDoneAndCloseWriter := task.OnSuccess(getResponse, task.Close(link.Writer))
	if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
		return errors.New("juicity connection ends").Base(err)
	}

	return nil
}

// outboundConn wraps outbound_netproxy.Conn to expose io.Reader/io.Writer for xray's buf.
type outboundConn struct {
	c outbound_netproxy.Conn
}

func outboundConnReader(c outbound_netproxy.Conn) io.Reader { return &outboundConn{c} }
func outboundConnWriter(c outbound_netproxy.Conn) io.Writer { return &outboundConn{c} }
func (n *outboundConn) Read(b []byte) (int, error)          { return n.c.Read(b) }
func (n *outboundConn) Write(b []byte) (int, error)         { return n.c.Write(b) }

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
