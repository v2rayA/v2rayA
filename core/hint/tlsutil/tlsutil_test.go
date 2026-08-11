package tlsutil

import "testing"

func TestParsePinnedChain(t *testing.T) {
	cases := []struct {
		name string
		pin  string
		ok   bool
	}{
		{"empty", "", true},
		{"std base64 padded", "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf+w4EA+c=", true},
		{"base64url unpadded", "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf-w4EA-c", true},
		{"base64url padded", "YAeLpzeMUkvQyVF5zpGov2ZMfdNfnfqkFHEf-w4EA-c=", true},
		{"hex", "601ceb02bd8a1e929519a9db9d00c1424c0948cd679bd06f407c913bc68a9da1", true},
		{"hex with colons", "60:1c:eb:02:bd:8a:1e:92:95:19:a9:db:9d:00:c1:42:4c:09:48:cd:67:9b:d0:6f:40:7c:91:3b:c6:8a:9d:a1", true},
		{"sha256 prefix hex", "sha256:601ceb02bd8a1e929519a9db9d00c1424c0948cd679bd06f407c913bc68a9da1", true},
		{"invalid", "not-a-pin", false},
		{"short base64", "YWJj", false},
	}
	for _, c := range cases {
		hash, err := ParsePinnedChain(c.pin)
		if c.ok {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", c.name, err)
				continue
			}
			if c.pin == "" {
				if len(hash) != 0 {
					t.Errorf("%s: empty pin must yield empty hash, got %d", c.name, len(hash))
				}
				continue
			}
			if len(hash) != 32 {
				t.Errorf("%s: expected 32-byte hash, got %d", c.name, len(hash))
			}
		} else if err == nil {
			t.Errorf("%s: expected error, got nil (hash len %d)", c.name, len(hash))
		}
	}
}

func TestGenerateChainHash(t *testing.T) {
	cert1 := []byte("cert-1")
	cert2 := []byte("cert-2")
	h1 := GenerateChainHash([][]byte{cert1})
	h2 := GenerateChainHash([][]byte{cert2})
	h12 := GenerateChainHash([][]byte{cert1, cert2})
	if len(h1) != 32 || len(h2) != 32 || len(h12) != 32 {
		t.Fatalf("expected 32-byte hashes")
	}
	if string(h1) == string(h2) {
		t.Fatalf("different certs must not collide")
	}
	if string(h12) == string(h1) || string(h12) == string(h2) {
		t.Fatalf("chained hash must differ from single-cert hashes")
	}
}

func TestVerifyPeerChain(t *testing.T) {
	hash := GenerateChainHash([][]byte{[]byte("cert")})
	if err := VerifyPeerChain([][]byte{[]byte("cert")}, [][]byte{hash}); err != nil {
		t.Fatalf("matching chain should verify: %v", err)
	}
	if err := VerifyPeerChain([][]byte{[]byte("other")}, [][]byte{hash}); err == nil {
		t.Fatalf("mismatched chain must fail")
	}
}
