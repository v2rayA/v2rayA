package juicity

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/protocol"
	_ "github.com/daeuniverse/softwind/protocol/juicity" // register juicity protocol

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
	dialer     netproxy.Dialer
	newDialer  func(internet.Dialer) (netproxy.Dialer, error)
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

	newDialer := func(dialer internet.Dialer) (netproxy.Dialer, error) {
		return protocol.NewDialer("juicity", xrayDialer{dialer: dialer}, protocol.Header{
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

func (d xrayDialer) Dial(network, addr string) (netproxy.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d xrayDialer) DialContext(ctx context.Context, network, addr string) (netproxy.Conn, error) {
	destination, err := xrayDestination(network, addr)
	if err != nil {
		return nil, err
	}
	return d.dialer.Dial(ctx, destination)
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
	if strings.HasPrefix(network, "udp") {
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

	conn, err := c.dialer.Dial("tcp", destAddr)
	if err != nil {
		return errors.New("juicity: failed to dial destination").Base(err)
	}
	defer conn.Close()

	postRequest := func() error {
		return xray_buf.Copy(link.Reader, xray_buf.NewWriter(netproxyConnAsWriter(conn)))
	}
	getResponse := func() error {
		return xray_buf.Copy(xray_buf.NewReader(netproxyConnAsReader(conn)), link.Writer)
	}

	responseDoneAndCloseWriter := task.OnSuccess(getResponse, task.Close(link.Writer))
	if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
		return errors.New("juicity connection ends").Base(err)
	}

	return nil
}

// netproxyConn wraps a netproxy.Conn to expose io.Reader/io.Writer for xray's buf.
type netproxyConn struct {
	c netproxy.Conn
}

func netproxyConnAsReader(c netproxy.Conn) io.Reader { return &netproxyConn{c} }
func netproxyConnAsWriter(c netproxy.Conn) io.Writer { return &netproxyConn{c} }
func (n *netproxyConn) Read(b []byte) (int, error)   { return n.c.Read(b) }
func (n *netproxyConn) Write(b []byte) (int, error)  { return n.c.Write(b) }

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
