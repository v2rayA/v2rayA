// Package certpin probes upstream TLS/QUIC endpoints and computes the
// certificate hashes required by xray-core and v2rayA's native protocols.
//
// It is used by the cert-fix job to decide, per node, whether the upstream
// certificate is trusted by the system store (no pinning needed) or whether
// a pinnedPeerCertSha256 / certificate-chain hash must be computed and saved.
package certpin

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

// ProtocolFamily identifies the transport layer used for the probe.
type ProtocolFamily string

const (
	ProtocolTCP  ProtocolFamily = "tcp"
	ProtocolQUIC ProtocolFamily = "quic"
)

// HashType identifies which kind of pin the consuming protocol expects.
type HashType string

const (
	// HashTypeLeaf is the SHA-256 hash of the leaf certificate (or an
	// intermediate CA) used by xray-core pinnedPeerCertSha256.
	HashTypeLeaf HashType = "leaf"
	// HashTypeChain is the cumulative SHA-256 hash of the whole peer
	// certificate chain used by native protocols such as tuic/juicity/anytls.
	HashTypeChain HashType = "chain"
)

// Target describes what to probe.
type Target struct {
	// ProtocolFamily is the transport used for the probe.
	ProtocolFamily ProtocolFamily
	// Address is the host:port of the upstream server.
	Address string
	// SNI is the ServerName sent in the TLS handshake.
	SNI string
	// ALPN is the list of application protocols sent in the TLS handshake.
	ALPN []string
	// HashType selects the hash format returned in Result.
	HashType HashType
}

// Result reports the outcome of a probe.
type Result struct {
	// Trusted is true when the system trust store accepted the certificate.
	Trusted bool
	// LeafHash is the hex-encoded SHA-256 of the leaf certificate. It is
	// populated for both HashTypeLeaf and HashTypeChain probes.
	LeafHash string
	// ChainHash is the hex-encoded cumulative hash of the peer certificate
	// chain. It is only populated when HashType is HashTypeChain.
	ChainHash string
	// IntermediateHash is the hex-encoded SHA-256 of the first intermediate
	// CA found in the chain (if any). It can be used as a more stable pin.
	IntermediateHash string
	// Error holds a human-readable error string for failures that are not
	// certificate-authority errors (e.g. timeout, connection refused).
	Error string
}

// ExtractTarget builds a probe Target from a parsed server object.
// The second return value is false when the server object does not represent
// a TLS-bearing outbound (e.g. plain shadowsocks) or when its protocol is not
// yet supported by the probe.
func ExtractTarget(obj serverObj.ServerObj) (*Target, bool) {
	switch s := obj.(type) {
	case *serverObj.V2Ray:
		return extractV2Ray(s)
	case *serverObj.Trojan:
		return extractTrojan(s)
	case *serverObj.Hysteria2:
		return extractHysteria2(s)
	case *serverObj.Tuic:
		return extractTuic(s)
	case *serverObj.Juicity:
		return extractJuicity(s)
	case *serverObj.AnyTLS:
		return extractAnyTLS(s)
	default:
		return nil, false
	}
}

// Probe performs a TLS/QUIC handshake against the target and returns the
// trust state plus certificate hashes. It honours ctx cancellation/deadline.
func Probe(ctx context.Context, target *Target) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if target == nil {
		return nil, errors.New("certpin: target is nil")
	}

	switch target.ProtocolFamily {
	case ProtocolTCP:
		return probeTCP(ctx, target)
	case ProtocolQUIC:
		return probeQUIC(ctx, target)
	default:
		return nil, fmt.Errorf("certpin: unsupported protocol family %q", target.ProtocolFamily)
	}
}

// extractV2Ray builds a target for VLESS/VMess outbounds that use TLS/XTLS.
func extractV2Ray(s *serverObj.V2Ray) (*Target, bool) {
	if s == nil {
		return nil, false
	}
	security := strings.ToLower(s.TLS)
	if security != "tls" && security != "xtls" {
		return nil, false
	}
	port, err := strconv.Atoi(s.Port)
	if err != nil || port <= 0 {
		return nil, false
	}
	addr := net.JoinHostPort(s.Add, strconv.Itoa(port))
	sni := s.SNI
	if sni == "" {
		sni = s.Host
	}
	if sni == "" {
		sni = s.Add
	}
	return &Target{
		ProtocolFamily: ProtocolTCP,
		Address:        addr,
		SNI:            sni,
		ALPN:           splitALPN(s.Alpn),
		HashType:       HashTypeLeaf,
	}, true
}

// extractTrojan builds a target for Trojan/Trojan-go TLS outbounds.
func extractTrojan(s *serverObj.Trojan) (*Target, bool) {
	if s == nil || s.Port <= 0 {
		return nil, false
	}
	sni := s.Sni
	if sni == "" {
		sni = s.Server
	}
	return &Target{
		ProtocolFamily: ProtocolTCP,
		Address:        net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		SNI:            sni,
		ALPN:           splitALPN(s.Alpn),
		HashType:       HashTypeLeaf,
	}, true
}

// extractHysteria2 builds a target for Hysteria2 (QUIC) outbounds.
func extractHysteria2(s *serverObj.Hysteria2) (*Target, bool) {
	if s == nil || s.Port <= 0 {
		return nil, false
	}
	p := parseHysteria2Link(s.Link)
	sni := p.sni
	if sni == "" {
		sni = s.Server
	}
	return &Target{
		ProtocolFamily: ProtocolQUIC,
		Address:        net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		SNI:            sni,
		ALPN:           nil, // hysteria2 does not negotiate ALPN for cert trust
		HashType:       HashTypeLeaf,
	}, true
}

// extractTuic builds a target for TUIC (QUIC) outbounds.
func extractTuic(s *serverObj.Tuic) (*Target, bool) {
	if s == nil || s.Port <= 0 {
		return nil, false
	}
	sni := s.Sni
	if sni == "" {
		sni = s.Server
	}
	return &Target{
		ProtocolFamily: ProtocolQUIC,
		Address:        net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		SNI:            sni,
		ALPN:           splitALPN(s.Alpn),
		HashType:       HashTypeChain,
	}, true
}

// extractJuicity builds a target for Juicity (QUIC) outbounds.
func extractJuicity(s *serverObj.Juicity) (*Target, bool) {
	if s == nil || s.Port <= 0 {
		return nil, false
	}
	p := parseJuicityLink(s.Link)
	sni := p.sni
	if sni == "" {
		sni = s.Server
	}
	return &Target{
		ProtocolFamily: ProtocolQUIC,
		Address:        net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		SNI:            sni,
		ALPN:           nil,
		HashType:       HashTypeChain,
	}, true
}

// extractAnyTLS builds a target for AnyTLS (TCP+TLS) outbounds.
func extractAnyTLS(s *serverObj.AnyTLS) (*Target, bool) {
	if s == nil || s.Port <= 0 {
		return nil, false
	}
	p := parseAnyTLSLink(s.Link)
	sni := p.sni
	if sni == "" {
		sni = s.Server
	}
	return &Target{
		ProtocolFamily: ProtocolTCP,
		Address:        net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		SNI:            sni,
		ALPN:           splitALPN(p.alpn),
		HashType:       HashTypeChain,
	}, true
}

// linkParams mirrors the query parameters parsed from hysteria2/juicity/anytls
// links. Keeping the structs unexported avoids extending the public API of the
// serverObj package while still letting certpin access the parsed values.
type hysteria2Params struct {
	sni                  string
	pinnedPeerCertSha256 string
	verifyPeerCertByName string
}

type juicityParams struct {
	sni                  string
	pinnedCertChainSha256 string
	allowInsecure        bool
}

type anytlsParams struct {
	sni                  string
	pinnedPeerCertSha256 string
	verifyPeerCertByName string
	alpn                 string
	allowInsecure        bool
}

func parseHysteria2Link(link string) hysteria2Params {
	var p hysteria2Params
	u, err := url.Parse(link)
	if err != nil {
		return p
	}
	q := u.Query()
	p.sni = q.Get("sni")
	p.pinnedPeerCertSha256 = q.Get("pinSHA256")
	if p.pinnedPeerCertSha256 == "" {
		p.pinnedPeerCertSha256 = q.Get("pinSha256")
	}
	if p.pinnedPeerCertSha256 == "" {
		p.pinnedPeerCertSha256 = q.Get("pin_sha256")
	}
	if p.pinnedPeerCertSha256 == "" {
		p.pinnedPeerCertSha256 = q.Get("pinned_peer_cert_sha256")
	}
	p.verifyPeerCertByName = q.Get("verify_peer_cert_by_name")
	return p
}

func parseJuicityLink(link string) juicityParams {
	var p juicityParams
	u, err := url.Parse(link)
	if err != nil {
		return p
	}
	q := u.Query()
	p.sni = q.Get("sni")
	p.pinnedCertChainSha256 = q.Get("pinned_certchain_sha256")
	p.allowInsecure = q.Get("allow_insecure") == "true" || q.Get("allowInsecure") == "true"
	return p
}

func parseAnyTLSLink(link string) anytlsParams {
	var p anytlsParams
	u, err := url.Parse(link)
	if err != nil {
		return p
	}
	q := u.Query()
	p.sni = q.Get("sni")
	p.pinnedPeerCertSha256 = q.Get("pinnedPeerCertSha256")
	p.verifyPeerCertByName = q.Get("verifyPeerCertByName")
	p.alpn = q.Get("alpn")
	p.allowInsecure = q.Get("allow_insecure") == "true" || q.Get("allow_insecure") == "1"
	return p
}

// HasPin reports whether the server object already carries an explicit
// certificate pin. Nodes that already have a pin are skipped by detection.
func HasPin(obj serverObj.ServerObj) bool {
	switch s := obj.(type) {
	case *serverObj.V2Ray:
		return strings.TrimSpace(s.PinnedPeerCertSha256) != ""
	case *serverObj.Trojan:
		return strings.TrimSpace(s.PinnedPeerCertSha256) != ""
	case *serverObj.Hysteria2:
		return strings.TrimSpace(parseHysteria2Link(s.Link).pinnedPeerCertSha256) != ""
	case *serverObj.Tuic:
		return strings.TrimSpace(s.PinnedPeerCertSha256) != ""
	case *serverObj.Juicity:
		return strings.TrimSpace(parseJuicityLink(s.Link).pinnedCertChainSha256) != ""
	case *serverObj.AnyTLS:
		return strings.TrimSpace(parseAnyTLSLink(s.Link).pinnedPeerCertSha256) != ""
	default:
		return false
	}
}

func ApplyResult(obj serverObj.ServerObj, res *Result) error {
	if res.Trusted {
		return clearInsecure(obj)
	}
	switch s := obj.(type) {
	case *serverObj.V2Ray:
		s.PinnedPeerCertSha256 = res.LeafHash
		return nil
	case *serverObj.Trojan:
		s.PinnedPeerCertSha256 = res.LeafHash
		return nil
	case *serverObj.Hysteria2:
		return setHysteria2Pin(s, res.LeafHash)
	case *serverObj.Tuic:
		s.AllowInsecure = false
		s.PinnedPeerCertSha256 = res.ChainHash
		return nil
	case *serverObj.Juicity:
		return setJuicityPin(s, res.ChainHash)
	case *serverObj.AnyTLS:
		s.AllowInsecure = false
		return setAnyTLSPin(s, res.ChainHash)
	default:
		return fmt.Errorf("certpin: unsupported server object type %T", obj)
	}
}

// IsAtRisk reports whether the server object should be offered to the user for
// certificate fixing. It is true for TLS-bearing outbounds that do not
// already carry a valid pin, or for native protocols that still have
// allow_insecure enabled.
func IsAtRisk(obj serverObj.ServerObj) bool {
	if _, ok := ExtractTarget(obj); !ok {
		return false
	}
	if HasPin(obj) {
		return false
	}
	return true
}

func clearInsecure(obj serverObj.ServerObj) error {
	switch s := obj.(type) {
	case *serverObj.V2Ray:
		s.PinnedPeerCertSha256 = ""
		return nil
	case *serverObj.Trojan:
		s.PinnedPeerCertSha256 = ""
		return nil
	case *serverObj.Hysteria2:
		return setHysteria2Pin(s, "")
	case *serverObj.Tuic:
		s.AllowInsecure = false
		s.PinnedPeerCertSha256 = ""
		return nil
	case *serverObj.Juicity:
		return setJuicityPin(s, "")
	case *serverObj.AnyTLS:
		s.AllowInsecure = false
		return setAnyTLSPin(s, "")
	default:
		return fmt.Errorf("certpin: unsupported server object type %T", obj)
	}
}

func setHysteria2Pin(s *serverObj.Hysteria2, pin string) error {
	u, err := url.Parse(s.Link)
	if err != nil {
		return err
	}
	q := u.Query()
	if pin == "" {
		q.Del("pinned_peer_cert_sha256")
		q.Del("pinSHA256")
		q.Del("pinSha256")
		q.Del("pin_sha256")
	} else {
		q.Set("pinned_peer_cert_sha256", pin)
	}
	u.RawQuery = q.Encode()
	s.Link = u.String()
	return nil
}

func setJuicityPin(s *serverObj.Juicity, pin string) error {
	u, err := url.Parse(s.Link)
	if err != nil {
		return err
	}
	q := u.Query()
	if pin == "" {
		q.Del("pinned_certchain_sha256")
	} else {
		q.Set("pinned_certchain_sha256", pin)
	}
	u.RawQuery = q.Encode()
	s.Link = u.String()
	return nil
}

func setAnyTLSPin(s *serverObj.AnyTLS, pin string) error {
	u, err := url.Parse(s.Link)
	if err != nil {
		return err
	}
	q := u.Query()
	if pin == "" {
		q.Del("pinnedPeerCertSha256")
	} else {
		q.Set("pinnedPeerCertSha256", pin)
	}
	u.RawQuery = q.Encode()
	s.Link = u.String()
	return nil
}

// probeTCP fetches the peer certificate chain with InsecureSkipVerify and then
// verifies the leaf certificate against the system trust store.
func probeTCP(ctx context.Context, target *Target) (*Result, error) {
	fetchConf := &tls.Config{
		ServerName:         target.SNI,
		NextProtos:         target.ALPN,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    fetchConf,
	}
	conn, err := dialer.DialContext(ctx, "tcp", target.Address)
	if err != nil {
		return &Result{Error: err.Error()}, nil
	}
	defer conn.Close()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return &Result{Error: "tls.Dialer returned a non-TLS connection"}, nil
	}

	peerCerts := tlsConn.ConnectionState().PeerCertificates
	res := buildResult(peerCerts, target.HashType)
	if res.Error != "" {
		return res, nil
	}

	res.Trusted = verifySystemTrust(peerCerts, target.SNI)
	return res, nil
}

// probeQUIC fetches the peer certificate chain with InsecureSkipVerify and then
// verifies the leaf certificate against the system trust store.
func probeQUIC(ctx context.Context, target *Target) (*Result, error) {
	fetchConf := &tls.Config{
		ServerName:         target.SNI,
		NextProtos:         target.ALPN,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	conn, err := quic.DialAddr(ctx, target.Address, fetchConf, &quic.Config{})
	if err != nil {
		return &Result{Error: err.Error()}, nil
	}
	defer conn.CloseWithError(0, "")

	peerCerts := conn.ConnectionState().TLS.PeerCertificates
	res := buildResult(peerCerts, target.HashType)
	if res.Error != "" {
		return res, nil
	}

	res.Trusted = verifySystemTrust(peerCerts, target.SNI)
	return res, nil
}

// verifySystemTrust reports whether the leaf certificate is trusted by the
// system store for the given server name.
func verifySystemTrust(certs []*x509.Certificate, serverName string) bool {
	if len(certs) == 0 {
		return false
	}
	opts := x509.VerifyOptions{
		DNSName:       serverName,
		Intermediates: x509.NewCertPool(),
	}
	for i := 1; i < len(certs); i++ {
		opts.Intermediates.AddCert(certs[i])
	}
	if _, err := certs[0].Verify(opts); err != nil {
		return false
	}
	return true
}

// buildResult computes hashes from the peer certificate chain.
func buildResult(certs []*x509.Certificate, hashType HashType) *Result {
	if len(certs) == 0 {
		return &Result{Error: "no peer certificates received"}
	}
	r := &Result{
		LeafHash: hex.EncodeToString(hashOf(certs[0].Raw)),
	}
	if len(certs) > 1 {
		r.IntermediateHash = hex.EncodeToString(hashOf(certs[1].Raw))
	}
	if hashType == HashTypeChain {
		raw := make([][]byte, len(certs))
		for i, c := range certs {
			raw[i] = c.Raw
		}
		r.ChainHash = hex.EncodeToString(generateChainHash(raw))
	}
	return r
}

// hashOf returns the SHA-256 hash of b.
func hashOf(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

// generateChainHash computes the cumulative chain hash used by native
// protocols. It mirrors core/hint/tlsutil/tlsutil.go without importing the
// core module into the service module.
func generateChainHash(rawCerts [][]byte) []byte {
	var hashValue []byte
	for _, certValue := range rawCerts {
		out := sha256.Sum256(certValue)
		if hashValue == nil {
			hashValue = out[:]
		} else {
			newHash := sha256.Sum256(append(hashValue, out[:]...))
			hashValue = newHash[:]
		}
	}
	return hashValue
}

// splitALPN converts a comma-separated ALPN string into a slice.
func splitALPN(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
