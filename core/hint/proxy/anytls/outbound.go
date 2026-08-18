package anytls

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"time"

	anytls_padding "anytls/proxy/padding"
	"anytls/proxy/session"
	anytls_util "anytls/util"

	"github.com/sagernet/sing/common/buf"
	M "github.com/sagernet/sing/common/metadata"

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

// Client is the anytls outbound handler.
type Client struct {
	config        *ClientConfig
	sessionClient *session.Client
}

// NewClient creates a new anytls outbound handler.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	if config.Address == "" {
		return nil, errors.New("anytls: no server address")
	}
	if config.Port == 0 || config.Port > 65535 {
		return nil, errors.New("anytls: invalid server port")
	}

	hash := sha256.Sum256([]byte(config.Password))
	passwordHash := make([]byte, 32)
	copy(passwordHash, hash[:])

	serverAddr := fmt.Sprintf("%s:%d", config.Address, config.Port)
	sni := config.Sni
	if sni == "" {
		sni = config.Address
	}

	tlsCfg := &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: config.AllowInsecure, // #nosec G402 -- user-configurable
	}
	if config.PinnedPeerCertificateChainSha256 != "" {
		hash, err := tlsutil.ParsePinnedChain(config.PinnedPeerCertificateChainSha256)
		if err != nil {
			return nil, fmt.Errorf("anytls: invalid pinned peer certificate chain hash: %w", err)
		}
		tlsCfg.VerifyPeerCertificate = tlsutil.PinVerifier(hash)
		// Pin verification is self-contained against rawCerts, so the chain / host
		// checks can be bypassed. Without this a self-signed cert would always fail.
		tlsCfg.InsecureSkipVerify = true // #nosec G402 -- guarded by pinning
	}

	minIdle := int(config.MinIdleSessions)
	if minIdle <= 0 {
		minIdle = 5
	}

	serverDest := xray_net.TCPDestination(xray_net.ParseAddress(config.Address), xray_net.Port(config.Port))
	_ = serverAddr // kept for readability

	dialOut := anytls_util.DialOutFunc(func(ctx context.Context) (net.Conn, error) {
		rawConn, err := internet.DialSystem(ctx, serverDest, nil)
		if err != nil {
			// Fallback: plain TCP dial
			d := &net.Dialer{Timeout: 10 * time.Second}
			rawConn, err = d.DialContext(ctx, "tcp", serverAddr)
			if err != nil {
				return nil, fmt.Errorf("anytls: failed to dial server: %w", err)
			}
		}

		tlsConn := tls.Client(rawConn, tlsCfg)

		// Write password SHA256 + padding (protocol handshake)
		b := buf.NewPacket()
		defer b.Release()

		b.Write(passwordHash)
		var paddingLen int
		if pad := anytls_padding.DefaultPaddingFactory.Load().GenerateRecordPayloadSizes(0); len(pad) > 0 {
			paddingLen = pad[0]
		}
		binary.BigEndian.PutUint16(b.Extend(2), uint16(paddingLen))
		if paddingLen > 0 {
			b.WriteZeroN(paddingLen)
		}
		if _, err = b.WriteTo(tlsConn); err != nil {
			tlsConn.Close()
			return nil, fmt.Errorf("anytls: failed to write handshake: %w", err)
		}

		return tlsConn, nil
	})

	bgCtx := context.Background()
	sessionClient := session.NewClient(
		bgCtx,
		dialOut,
		&anytls_padding.DefaultPaddingFactory,
		30*time.Second,
		30*time.Second,
		minIdle,
	)

	return &Client{
		config:        config,
		sessionClient: sessionClient,
	}, nil
}

// Process implements proxy.Outbound.
func (c *Client) Process(ctx context.Context, link *transport.Link, _ internet.Dialer) error {
	outbounds := xray_session.OutboundsFromContext(ctx)
	ob := outbounds[len(outbounds)-1]
	if !ob.Target.IsValid() {
		return errors.New("target not specified")
	}
	destination := ob.Target

	// Build sing Socksaddr from xray Destination
	dest := xrayDestToSocksaddr(destination)

	stream, err := c.sessionClient.CreateStream(ctx)
	if err != nil {
		return errors.New("failed to create anytls stream").Base(err)
	}
	defer stream.Close()

	if err := M.SocksaddrSerializer.WriteAddrPort(stream, dest); err != nil {
		return errors.New("failed to write destination to anytls stream").Base(err)
	}

	postRequest := func() error {
		return xray_buf.Copy(link.Reader, xray_buf.NewWriter(stream))
	}
	getResponse := func() error {
		return xray_buf.Copy(xray_buf.NewReader(stream), link.Writer)
	}

	responseDoneAndCloseWriter := task.OnSuccess(getResponse, task.Close(link.Writer))
	if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
		return errors.New("anytls connection ends").Base(err)
	}

	return nil
}

// xrayDestToSocksaddr converts an xray net.Destination to a sing M.Socksaddr.
func xrayDestToSocksaddr(d xray_net.Destination) M.Socksaddr {
	var sa M.Socksaddr
	sa.Port = uint16(d.Port.Value())
	addr := d.Address
	if addr.Family().IsDomain() {
		sa.Fqdn = addr.Domain()
	} else {
		ip := addr.IP()
		if ip4 := ip.To4(); ip4 != nil {
			var arr [4]byte
			copy(arr[:], ip4)
			sa.Addr = netip.AddrFrom4(arr)
		} else {
			var arr [16]byte
			copy(arr[:], ip.To16())
			sa.Addr = netip.AddrFrom16(arr)
		}
	}
	return sa
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
