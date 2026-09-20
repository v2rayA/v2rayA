package dns

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

// dohRootCAs, when set, replaces the system roots; tests point it at their
// server's certificate.
var dohRootCAs *x509.CertPool

// dialFunc opens the TCP connection an https or tls upstream is spoken over:
// the marked dialer for a direct upstream, the xray dispatcher for a proxied
// one. The address is host:port; the host may be the upstream's name.
type dialFunc func(ctx context.Context, addr string) (net.Conn, error)

// SetBootstrapIPs records what bootstrap resolved for the upstream hostnames,
// so an https upstream is dialled by IP while TLS still verifies its name.
func (m *UpstreamManager) SetBootstrapIPs(ips map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bootstrapIPs = ips
}

// dialTarget replaces a hostname by its bootstrapped IP; a query for the
// upstream's own name would otherwise come back to this module.
func (m *UpstreamManager) dialTarget(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	m.mu.RLock()
	ip, ok := m.bootstrapIPs[host]
	m.mu.RUnlock()
	if ok && ip != "" {
		return net.JoinHostPort(ip, port)
	}
	return addr
}

// dohClient builds the upstream's HTTP client once: TLS is done here so the
// server name is the upstream's hostname whatever address was dialled, and
// the transport keeps the connection for the next query.
func (m *UpstreamManager) dohClient(upstream *UpstreamInstance, dial dialFunc) *http.Client {
	upstream.dohOnce.Do(func() {
		transport := &http.Transport{
			ForceAttemptHTTP2: true,
			IdleConnTimeout:   90 * time.Second,
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				raw, err := dial(ctx, m.dialTarget(addr))
				if err != nil {
					return nil, err
				}
				host, _, _ := net.SplitHostPort(addr)
				conn := tls.Client(raw, &tls.Config{ServerName: host, RootCAs: dohRootCAs, NextProtos: []string{"h2", "http/1.1"}})
				if err := conn.HandshakeContext(ctx); err != nil {
					raw.Close()
					return nil, err
				}
				return conn, nil
			},
		}
		upstream.doh = &http.Client{Transport: transport, Timeout: 5 * time.Second}
	})
	return upstream.doh
}

// exchangeDoH sends msg to an https upstream as RFC 8484 POST and returns
// the answer. upstream.Addr is the URL.
func (m *UpstreamManager) exchangeDoH(upstream *UpstreamInstance, msg *dns.Msg, dial dialFunc) (*dns.Msg, time.Duration, error) {
	packed, err := msg.Pack()
	if err != nil {
		return nil, 0, fmt.Errorf("dns pack: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upstream.Addr, bytes.NewReader(packed))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")
	start := time.Now()
	resp, err := m.dohClient(upstream, dial).Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("doh %s: HTTP %d", upstream.Addr, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, dns.MaxMsgSize+1))
	if err != nil {
		return nil, 0, err
	}
	if len(body) > dns.MaxMsgSize {
		return nil, 0, fmt.Errorf("doh %s: response over %d bytes", upstream.Addr, dns.MaxMsgSize)
	}
	answer := new(dns.Msg)
	if err := answer.Unpack(body); err != nil {
		return nil, 0, fmt.Errorf("dns unpack: %w", err)
	}
	return answer, time.Since(start), nil
}

// directDial is the dialFunc of a direct upstream: a marked socket, so the
// transparent-proxy rules let the connection out.
func directDial(ctx context.Context, addr string) (net.Conn, error) {
	return markedDialer().DialContext(ctx, "tcp", addr)
}

// tlsOver wraps a connection to a tls upstream and verifies the upstream's
// name; used where miekg/dns does not own the connection.
func tlsOver(ctx context.Context, raw net.Conn, serverName string) (net.Conn, error) {
	conn := tls.Client(raw, &tls.Config{ServerName: serverName})
	if err := conn.HandshakeContext(ctx); err != nil {
		raw.Close()
		return nil, err
	}
	return conn, nil
}
