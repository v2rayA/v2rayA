package coreObj

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestPinnedPeerCertSha256Hex(t *testing.T) {
	const canonicalHex = "601ceb02bd8a1e929519a9db9d00c1424c0948cd679bd06f407c913bc68a9da1"
	cases := []struct {
		name string
		pin  string
		ok   bool
		want string
	}{
		{"empty", "", true, ""},
		{"std base64 padded", "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf+w4EA+c=", true, ""},
		{"base64url unpadded", "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf-w4EA-c", true, ""},
		{"base64url padded", "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf-w4EA-c=", true, ""},
		{"hex", canonicalHex, true, canonicalHex},
		{"hex uppercase", strings.ToUpper(canonicalHex), true, canonicalHex},
		{"hex with colons", "60:1c:eb:02:bd:8a:1e:92:95:19:a9:db:9d:00:c1:42:4c:09:48:cd:67:9b:d0:6f:40:7c:91:3b:c6:8a:9d:a1", true, canonicalHex},
		{"sha256 prefix hex", "sha256:" + canonicalHex, true, canonicalHex},
		{"comma separated", canonicalHex + "," + canonicalHex, true, canonicalHex + "," + canonicalHex},
		{"invalid", "not-a-pin", false, ""},
	}
	for _, c := range cases {
		got, err := PinnedPeerCertSha256Hex(c.pin)
		if c.ok {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", c.name, err)
				continue
			}
			if c.pin == "" {
				if got != "" {
					t.Errorf("%s: empty pin must yield empty output, got %q", c.name, got)
				}
				continue
			}
			if c.want != "" {
				if got != c.want {
					t.Errorf("%s: expected %q, got %q", c.name, c.want, got)
				}
				continue
			}
			// base64 inputs: output must be canonical lowercase hex of 32 bytes
			raw, derr := hex.DecodeString(got)
			if derr != nil || len(raw) != 32 {
				t.Errorf("%s: output %q is not hex of 32 bytes", c.name, got)
				continue
			}
			if strings.ToLower(got) != got {
				t.Errorf("%s: output %q must be lowercase", c.name, got)
			}
		} else if err == nil {
			t.Errorf("%s: expected error, got nil (%q)", c.name, got)
		}
	}
}

func TestPinnedPeerCertSha256HexBase64RoundTrip(t *testing.T) {
	// base64 input must be converted to the equivalent hex representation
	b64 := "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf+w4EA+c="
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	want := hex.EncodeToString(raw)
	got, err := PinnedPeerCertSha256Hex(b64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestPinnedPeerCertSha256HexNormalizesURLSafe(t *testing.T) {
	// - and _ must be mapped back to + and / before decoding
	got, err := PinnedPeerCertSha256Hex("YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf-w4EA-c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, _ := PinnedPeerCertSha256Hex("YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf+w4EA+c=")
	if got != want {
		t.Fatalf("URL-safe pin must normalize to the same hex, got %q want %q", got, want)
	}
}

func TestXHTTPRangeConfigMarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   XHTTPRangeConfig
		want string
	}{
		{"single value", XHTTPRangeConfig{From: 123, To: 123}, "123"},
		{"range", XHTTPRangeConfig{From: 123, To: 456}, `"123-456"`},
		{"zero", XHTTPRangeConfig{}, "0"},
		{"negative range", XHTTPRangeConfig{From: -1919, To: -810}, `"-1919--810"`},
	}
	for _, c := range cases {
		got, err := jsoniter.Marshal(c.in)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.name, err)
		}
		if string(got) != c.want {
			t.Errorf("%s: expected %s, got %s", c.name, c.want, got)
		}
	}
}

func TestXHTTPSettingsRangeJSONShape(t *testing.T) {
	s := XHTTPSettings{
		Path:          "/path",
		XPaddingBytes: &XHTTPRangeConfig{From: 123, To: 456},
		Xmux: &XHTTPXmux{
			MaxConcurrency:   &XHTTPRangeConfig{From: 8, To: 8},
			HKeepAlivePeriod: 30,
		},
	}
	got, err := jsoniter.Marshal(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"path":"/path","xPaddingBytes":"123-456","xmux":{"maxConcurrency":8,"hKeepAlivePeriod":30}}`
	if string(got) != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestXHTTPRangeConfigUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want XHTTPRangeConfig
	}{
		{"plain int", "123", XHTTPRangeConfig{From: 123, To: 123}},
		{"range string", `"123-456"`, XHTTPRangeConfig{From: 123, To: 456}},
		{"number string", `"123"`, XHTTPRangeConfig{From: 123, To: 123}},
		{"negative range", `"-1919--810"`, XHTTPRangeConfig{From: -1919, To: -810}},
	}
	for _, c := range cases {
		var got XHTTPRangeConfig
		if err := json.Unmarshal([]byte(c.in), &got); err != nil {
			t.Fatalf("%s: unexpected error: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s: expected %+v, got %+v", c.name, c.want, got)
		}
	}
}

func TestXHTTPRangeConfigUnmarshalJSONInvalid(t *testing.T) {
	for _, in := range []string{`"abc"`, `{"from":1,"to":2}`, "true"} {
		var got XHTTPRangeConfig
		if err := json.Unmarshal([]byte(in), &got); err == nil {
			t.Errorf("%s: expected error, got %+v", in, got)
		}
	}
}
