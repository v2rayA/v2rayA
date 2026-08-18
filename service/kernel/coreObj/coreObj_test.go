package coreObj

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
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
