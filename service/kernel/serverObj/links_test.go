package serverObj

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
)

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
