package serverObj

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestParseSip003PrependsSlashToPath(t *testing.T) {
	if got := ParseSip003Opts("obfs-uri=abc").Path; got != "/abc" {
		t.Fatalf("path = %q, want /abc", got)
	}
}

func TestParseSSURLAcceptsPluginWithoutOptions(t *testing.T) {
	u := &url.URL{
		Scheme: "ss",
		User:   url.User(base64.RawURLEncoding.EncodeToString([]byte("aes-128-gcm:secret"))),
		Host:   "example.com:8388",
	}
	query := u.Query()
	query.Set("plugin", "v2ray-plugin")
	u.RawQuery = query.Encode()
	server, err := ParseSSURL(u.String())
	if err != nil {
		t.Fatal(err)
	}
	if server.Plugin.Name != "v2ray-plugin" {
		t.Fatalf("plugin = %+v", server.Plugin)
	}
}

func TestParseVlessURLRejectsMissingOrInvalidPort(t *testing.T) {
	for _, link := range []string{"vless://uuid@host", "vless://uuid@host:abc"} {
		t.Run(link, func(t *testing.T) {
			if _, err := ParseVlessURL(link); err == nil {
				t.Fatalf("ParseVlessURL(%q) succeeded", link)
			}
		})
	}
}

func TestParseVmessURLRejectsMissingOrInvalidPort(t *testing.T) {
	for _, port := range []string{"", "abc"} {
		t.Run(port, func(t *testing.T) {
			payload, err := json.Marshal(map[string]string{"add": "host", "port": port, "id": "uuid"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseVmessURL("vmess://" + base64.StdEncoding.EncodeToString(payload)); err == nil {
				t.Fatalf("VMess JSON with port %q succeeded", port)
			}
		})
	}
}

func TestHTTPConfigurationUsesNativeOutbound(t *testing.T) {
	obj, err := ParseHttpURL("http://user:pass@1.2.3.4:8080")
	if err != nil {
		t.Fatal(err)
	}
	config, err := obj.Configuration(PriorInfo{Tag: "proxy", PluginPort: 12345})
	if err != nil {
		t.Fatal(err)
	}
	servers, ok := config.CoreOutbound.Settings.Servers.([]coreObj.Server)
	if !ok || len(servers) != 1 {
		t.Fatalf("servers = %#v", config.CoreOutbound.Settings.Servers)
	}
	if config.CoreOutbound.Protocol != "http" || servers[0].Address != "1.2.3.4" || servers[0].Port != 8080 || len(servers[0].Users) != 1 || servers[0].Users[0].User != "user" || servers[0].Users[0].Pass != "pass" {
		t.Fatalf("configuration = %+v", config)
	}
	raw, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "127.0.0.1") {
		t.Fatalf("configuration still uses plugin loopback: %s", raw)
	}
	if obj.NeedPluginPort() {
		t.Fatal("native HTTP outbound requested a plugin port")
	}
}

func TestHTTPSConfigurationUsesTLS(t *testing.T) {
	obj, err := ParseHttpURL("https://proxy.example:8443")
	if err != nil {
		t.Fatal(err)
	}
	config, err := obj.Configuration(PriorInfo{Tag: "proxy"})
	if err != nil {
		t.Fatal(err)
	}
	stream := config.CoreOutbound.StreamSettings
	if stream == nil || stream.Security != "tls" || stream.TLSSettings == nil || stream.TLSSettings.ServerName != obj.Server {
		t.Fatalf("stream settings = %+v", stream)
	}
}

func TestSSRLinksAreRefusedAndStoredNodesStayUnsupported(t *testing.T) {
	if _, err := NewFromLink("ssr", "ssr://ZXhhbXBsZS5jb206ODM4ODpvcmlnaW46YWVzLTI1Ni1jZmI6cGxhaW46YzJWamNtVjA"); err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("ssr link import error = %v", err)
	}
	stored, err := New("shadowsocksr")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stored.Configuration(PriorInfo{}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("stored SSR node configuration error = %v", err)
	}
}

func TestParseSSURLUserinfo(t *testing.T) {
	const (
		cipher   = "aes-256-gcm"
		password = "丿x"
	)
	credentials := []byte(cipher + ":" + password)
	standard := base64.StdEncoding.EncodeToString(credentials)
	if !strings.Contains(standard, "/") {
		t.Fatalf("standard base64 credential %q does not contain /", standard)
	}
	for _, tc := range []struct {
		name     string
		userinfo string
	}{
		{"standard padded", standard},
		{"standard unpadded", strings.TrimRight(standard, "=")},
		{"URL-safe padded", base64.URLEncoding.EncodeToString(credentials)},
		{"URL-safe unpadded", strings.TrimRight(base64.URLEncoding.EncodeToString(credentials), "=")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			link := (&url.URL{
				Scheme:   "ss",
				User:     url.User(tc.userinfo),
				Host:     "example.com:8388",
				Fragment: "standard",
			}).String()
			got, err := ParseSSURL(link)
			if err != nil {
				t.Fatal(err)
			}
			if got.Cipher != cipher || got.Password != password || got.Server != "example.com" || got.Port != 8388 || got.Name != "standard" {
				t.Fatalf("got %#v", got)
			}
		})
	}

	got, err := ParseSSURL("ss://aes-256-gcm:plain@example.com:8388")
	if err != nil {
		t.Fatal(err)
	}
	if got.Cipher != "aes-256-gcm" || got.Password != "plain" {
		t.Fatalf("raw userinfo parsed as %#v", got)
	}
}

func TestShadowsocksURLRoundTripChinesePassword(t *testing.T) {
	original := &Shadowsocks{
		Cipher:   "aes-256-gcm",
		Password: "密码",
		Server:   "example.com",
		Port:     8388,
		Name:     "测试",
	}
	got, err := ParseSSURL(original.ExportToURL())
	if err != nil {
		t.Fatal(err)
	}
	if got.Cipher != original.Cipher || got.Password != original.Password || got.Server != original.Server || got.Port != original.Port || got.Name != original.Name {
		t.Fatalf("got %#v", got)
	}
}

func TestParseSocksURLDefaultPort(t *testing.T) {
	got, err := ParseSocksURL("socks5://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.Protocol != "socks5" || got.Server != "example.com" || got.Port != 1080 {
		t.Fatalf("got %#v", got)
	}

	for _, bad := range []string{"socks5://", "socks5://user:pass@"} {
		if _, err := ParseSocksURL(bad); err == nil {
			t.Errorf("%s should be rejected", bad)
		}
	}
}

func TestVlessURLRoundTrip(t *testing.T) {
	original := &V2Ray{
		Ps:                  "name",
		Add:                 "example.com",
		Port:                "443",
		ID:                  "id /: 密",
		Net:                 "grpc",
		Path:                "service",
		TLS:                 "none",
		MultiMode:           "gun",
		IdleTimeout:         "30",
		HealthCheckTimeout:  "5",
		PermitWithoutStream: "true",
		InitialWindowsSize:  "65535",
		Protocol:            "vless",
	}
	link := original.ExportToURL()
	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	if u.User == nil || u.User.String() != url.User(original.ID).String() {
		t.Fatalf("VLESS ID is not encoded once in %q", link)
	}
	for key, want := range map[string]string{
		"serviceName":         original.Path,
		"multiMode":           original.MultiMode,
		"idleTimeout":         original.IdleTimeout,
		"healthCheckTimeout":  original.HealthCheckTimeout,
		"permitWithoutStream": original.PermitWithoutStream,
		"initialWindowsSize":  original.InitialWindowsSize,
	} {
		if got := u.Query().Get(key); got != want {
			t.Errorf("%s: got %q, want %q", key, got, want)
		}
	}
	got, err := ParseVlessURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != original.ID || got.Path != original.Path || got.MultiMode != original.MultiMode || got.IdleTimeout != original.IdleTimeout || got.HealthCheckTimeout != original.HealthCheckTimeout || got.PermitWithoutStream != original.PermitWithoutStream || got.InitialWindowsSize != original.InitialWindowsSize {
		t.Fatalf("got %#v", got)
	}
}

// A share link must carry base64 a strict decoder accepts: the exporters used
// to trim a single "=" from a value that needed two.
func TestExportedBase64IsUnpadded(t *testing.T) {
	ss := &Shadowsocks{Cipher: "aes-128-gcm", Password: "test", Server: "1.2.3.4", Port: 8388, Name: "n"}
	link := ss.ExportToURL()
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("%s: %v", link, err)
	}
	userinfo := u.User.Username()
	if strings.Contains(userinfo, "=") {
		t.Errorf("shadowsocks userinfo %q still carries padding", userinfo)
	}
	if _, err := base64.RawURLEncoding.DecodeString(userinfo); err != nil {
		t.Errorf("shadowsocks userinfo %q is not valid unpadded base64: %v", userinfo, err)
	}
	back, err := ParseSSURL(link)
	if err != nil {
		t.Fatalf("%s: %v", link, err)
	}
	if back.Password != ss.Password || back.Cipher != ss.Cipher {
		t.Errorf("round trip changed the credentials: %q/%q", back.Cipher, back.Password)
	}
}

func TestDegenerateLinksDoNotPanic(t *testing.T) {
	for _, link := range []string{"ss:", "ss:x", "vmess:", "vmess:x", "ssr:", "trojan:", "vless:"} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%q panicked: %v", link, r)
				}
			}()
			scheme := link[:strings.Index(link, ":")]
			if _, err := NewFromLink(scheme, link); err == nil {
				t.Fatalf("%q parsed", link)
			}
		}()
	}
}

func TestVlessRawTransportIsTCP(t *testing.T) {
	obj, err := NewFromLink("vless", "vless://b831381d-6324-4d53-ad4f-8cda48b30811@1.2.3.4:443?type=raw&security=reality&sni=example.com&pbk=key&fp=chrome#raw")
	if err != nil {
		t.Fatal(err)
	}
	c, err := obj.Configuration(PriorInfo{Tag: "t"})
	if err != nil {
		t.Fatalf("raw transport rejected: %v", err)
	}
	if c.CoreOutbound.StreamSettings == nil || c.CoreOutbound.StreamSettings.Network != "tcp" {
		t.Fatalf("stream settings = %+v", c.CoreOutbound.StreamSettings)
	}
}

func TestVmessNumericPortAndAid(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte(`{"add":"1.2.3.4","aid":0,"id":"93637105-bcea-4a68-b089-1bb6091f0b16","net":"tcp","port":27467,"ps":"n","tls":"none","type":"http","v":2}`))
	obj, err := NewFromLink("vmess", "vmess://"+payload)
	if err != nil {
		t.Fatalf("numeric port/aid rejected: %v", err)
	}
	if obj.GetPort() != 27467 {
		t.Fatalf("port = %d", obj.GetPort())
	}
}
