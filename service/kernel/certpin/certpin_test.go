package certpin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestExtractTargetV2Ray(t *testing.T) {
	obj := &serverObj.V2Ray{
		Add:  "1.2.3.4",
		Port: "443",
		SNI:  "example.com",
		TLS:  "tls",
		Alpn: "h2,h3",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, ProtocolTCP, target.ProtocolFamily)
	assert.Equal(t, "1.2.3.4:443", target.Address)
	assert.Equal(t, "example.com", target.SNI)
	assert.Equal(t, []string{"h2", "h3"}, target.ALPN)
	assert.Equal(t, HashTypeLeaf, target.HashType)
}

func TestExtractTargetV2RayFallbackSNI(t *testing.T) {
	obj := &serverObj.V2Ray{
		Add:  "example.com",
		Port: "443",
		TLS:  "tls",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, "example.com", target.SNI)
}

func TestExtractTargetV2RayNoTLS(t *testing.T) {
	obj := &serverObj.V2Ray{
		Add:  "1.2.3.4",
		Port: "80",
	}
	_, ok := ExtractTarget(obj)
	assert.False(t, ok)
}

func TestExtractTargetTrojan(t *testing.T) {
	obj := &serverObj.Trojan{
		Server: "example.com",
		Port:   443,
		Sni:    "cdn.example.com",
		Alpn:   "h2",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, ProtocolTCP, target.ProtocolFamily)
	assert.Equal(t, "cdn.example.com", target.SNI)
}

func TestExtractTargetHysteria2(t *testing.T) {
	obj := &serverObj.Hysteria2{
		Server: "hy.example.com",
		Port:   443,
		Link:   "hysteria2://user:pass@hy.example.com:443?sni=hy.example.com#name",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, ProtocolQUIC, target.ProtocolFamily)
	assert.Equal(t, "hy.example.com:443", target.Address)
	assert.Equal(t, "hy.example.com", target.SNI)
}

func TestExtractTargetTuic(t *testing.T) {
	obj := &serverObj.Tuic{
		Server: "tuic.example.com",
		Port:   443,
		Sni:    "sni.example.com",
		Alpn:   "h3",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, ProtocolQUIC, target.ProtocolFamily)
	assert.Equal(t, HashTypeChain, target.HashType)
	assert.Equal(t, "sni.example.com", target.SNI)
	assert.Equal(t, []string{"h3"}, target.ALPN)
}

func TestExtractTargetJuicity(t *testing.T) {
	obj := &serverObj.Juicity{
		Server: "j.example.com",
		Port:   443,
		Link:   "juicity://uuid:pass@j.example.com:443?sni=sni.example.com#name",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, ProtocolQUIC, target.ProtocolFamily)
	assert.Equal(t, HashTypeChain, target.HashType)
	assert.Equal(t, "sni.example.com", target.SNI)
}

func TestExtractTargetAnyTLS(t *testing.T) {
	obj := &serverObj.AnyTLS{
		Server: "a.example.com",
		Port:   443,
		Link:   "anytls://pass@a.example.com:443?sni=sni.example.com&alpn=h2#name",
	}
	target, ok := ExtractTarget(obj)
	require.True(t, ok)
	assert.Equal(t, ProtocolTCP, target.ProtocolFamily)
	assert.Equal(t, HashTypeChain, target.HashType)
	assert.Equal(t, "sni.example.com", target.SNI)
	assert.Equal(t, []string{"h2"}, target.ALPN)
}

func TestHasPin(t *testing.T) {
	assert.False(t, HasPin(&serverObj.V2Ray{TLS: "tls"}))
	assert.True(t, HasPin(&serverObj.V2Ray{TLS: "tls", PinnedPeerCertSha256: "deadbeef"}))
	assert.True(t, HasPin(&serverObj.Trojan{PinnedPeerCertSha256: "abc"}))
	assert.True(t, HasPin(&serverObj.Hysteria2{
		Link: "hysteria2://x@y:443?pinned_peer_cert_sha256=abcd#name",
	}))
	assert.True(t, HasPin(&serverObj.Juicity{
		Link: "juicity://x@y:443?pinned_certchain_sha256=abcd#name",
	}))
}

func TestProbeTCPSelfSigned(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	parsed, err := url.Parse(ts.URL)
	require.NoError(t, err)

	target := &Target{
		ProtocolFamily: ProtocolTCP,
		Address:        parsed.Host,
		SNI:            "localhost",
		HashType:       HashTypeLeaf,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := Probe(ctx, target)
	require.NoError(t, err)
	assert.False(t, res.Trusted)
	assert.Empty(t, res.Error)
	assert.NotEmpty(t, res.LeafHash)
	assert.Len(t, res.LeafHash, 64)
}

func TestGenerateChainHash(t *testing.T) {
	data := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	h := generateChainHash(data)
	assert.Len(t, h, 32)
}

func TestBuildResultNoCerts(t *testing.T) {
	res := buildResult(nil, HashTypeLeaf)
	assert.NotEmpty(t, res.Error)
}

func TestSplitALPN(t *testing.T) {
	assert.Nil(t, splitALPN(""))
	assert.Equal(t, []string{"h2", "h3"}, splitALPN("h2,h3"))
	assert.Equal(t, []string{"h2"}, splitALPN(" h2 "))
}
