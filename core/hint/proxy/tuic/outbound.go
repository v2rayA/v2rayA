package tuic

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/netip"
	"strings"

	outbound_netproxy "github.com/daeuniverse/outbound/netproxy"
	outbound_protocol "github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/direct"
	_ "github.com/daeuniverse/outbound/protocol/tuic" // register tuic protocol

	"github.com/xtls/xray-core/common"
	xray_buf "github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/errors"
	xray_session "github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/transport"
	"github.com/xtls/xray-core/transport/internet"
)

// Client is the tuic outbound handler.
type Client struct {
	config *ClientConfig
	dialer outbound_netproxy.Dialer
}

// NewClient creates a new tuic outbound handler.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	if config.Address == "" {
		return nil, errors.New("tuic: no server address")
	}
	if config.Uuid == "" {
		return nil, errors.New("tuic: no UUID")
	}

	sni := config.Sni
	if sni == "" && !config.DisableSni {
		// extract host part from address (may be "host:port")
		host := config.Address
		if h, _, err := splitHostPort(host); err == nil {
			sni = h
		} else {
			sni = host
		}
	}

	alpn := config.Alpn
	if len(alpn) == 0 {
		alpn = []string{"h3"}
	}

	tlsCfg := &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: config.AllowInsecure, // #nosec G402 -- user-configurable
		NextProtos:         alpn,
		MinVersion:         tls.VersionTLS13,
	}
	if config.DisableSni {
		tlsCfg.ServerName = ""
	}

	congestion := config.CongestionControl
	if congestion == "" {
		congestion = "bbr"
	}

	// TUIC uses QUIC/UDP underneath; use a direct UDP dialer as the underlying transport.
	nextDialer := direct.NewDirectDialerLaddr(netip.Addr{}, direct.Option{FullCone: false})

	dialer, err := outbound_protocol.NewDialer("tuic", nextDialer, outbound_protocol.Header{
		ProxyAddress: config.Address,
		TlsConfig:    tlsCfg,
		User:         config.Uuid,
		Password:     config.Password,
		Feature1:     congestion,
		IsClient:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("tuic: failed to create dialer: %w", err)
	}

	return &Client{
		config: config,
		dialer: dialer,
	}, nil
}

// splitHostPort splits a host:port string, returning host and port separately.
// It handles IPv6 addresses in brackets.
func splitHostPort(addr string) (host, port string, err error) {
	// Try standard net.SplitHostPort first.
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return addr, "", nil
	}
	host = addr[:i]
	port = addr[i+1:]
	// Strip brackets from IPv6.
	if len(host) >= 2 && host[0] == '[' && host[len(host)-1] == ']' {
		host = host[1 : len(host)-1]
	}
	return host, port, nil
}

// Process implements proxy.Outbound.
func (c *Client) Process(ctx context.Context, link *transport.Link, _ internet.Dialer) error {
	outbounds := xray_session.OutboundsFromContext(ctx)
	ob := outbounds[len(outbounds)-1]
	if !ob.Target.IsValid() {
		return errors.New("target not specified")
	}
	destination := ob.Target

	destAddr := fmt.Sprintf("%s:%d", destination.Address.String(), destination.Port.Value())

	conn, err := c.dialer.DialContext(ctx, "tcp", destAddr)
	if err != nil {
		return errors.New("tuic: failed to dial destination").Base(err)
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
		return errors.New("tuic connection ends").Base(err)
	}

	return nil
}

// outboundConnReader/Writer wraps outbound_netproxy.Conn to expose io.Reader/Writer for xray's buf.
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
