// Package tlsutil provides shared certificate-chain pinning helpers used by the
// native outbound protocol handlers (anytls, tuic, juicity).
//
// SPDX-License-Identifier: MPL-2.0
package tlsutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// ParsePinnedChain accepts a pin value (hex or base64/base64url, optional
// "sha256:" prefix, optional openssl-style colon separators) and returns the
// parsed 32-byte chain hash. An empty pin yields (nil, nil). A non-empty but
// invalid pin returns an error so callers can fail loudly instead of silently
// dropping the pin.
func ParsePinnedChain(pin string) ([]byte, error) {
	pin = strings.TrimSpace(pin)
	pin = strings.TrimPrefix(pin, "sha256:")
	pin = strings.TrimPrefix(pin, "SHA256:")
	// Strip openssl-style colons: AB:CD:EF:...
	pin = strings.ReplaceAll(pin, ":", "")
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return nil, nil
	}
	if raw, err := hex.DecodeString(pin); err == nil && len(raw) == 32 {
		return raw, nil
	}
	// Standard and URL-safe (base64url) alphabets, with and without padding.
	for _, dec := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if raw, err := dec.DecodeString(pin); err == nil && len(raw) == 32 {
			return raw, nil
		}
	}
	return nil, fmt.Errorf("unrecognized certificate chain pin: %q", pin)
}

// GenerateChainHash mirrors xray-core's tls.GenerateCertChainHash so that pin
// values computed for v2rayA / xray are interchangeable across protocols.
func GenerateChainHash(rawCerts [][]byte) []byte {
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

// VerifyPeerChain verifies that the peer certificate chain matches one of the
// expected chain hashes.
func VerifyPeerChain(rawCerts [][]byte, expected [][]byte) error {
	chainHash := GenerateChainHash(rawCerts)
	for _, h := range expected {
		if hmac.Equal(chainHash, h) {
			return nil
		}
	}
	return fmt.Errorf("peer certificate chain is unrecognized (hash %s)",
		base64.StdEncoding.EncodeToString(chainHash))
}

// PinVerifier returns a tls.Config.VerifyPeerCertificate callback that enforces
// the pinned certificate chain hash against the peer certificate chain.
func PinVerifier(hash []byte) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		return VerifyPeerChain(rawCerts, [][]byte{hash})
	}
}
