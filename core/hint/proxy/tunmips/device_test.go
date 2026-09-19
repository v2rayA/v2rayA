// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"net/netip"
	"testing"
)

func TestPeerOf(t *testing.T) {
	for in, want := range map[string]string{
		"10.0.85.2/30":          "10.0.85.1",
		"10.0.85.1/30":          "10.0.85.2",
		"fdfe:dcba:9876::2/126": "fdfe:dcba:9876::1",
		"169.254.10.2/30":       "169.254.10.1",
	} {
		if got := peerOf(netip.MustParsePrefix(in)); got != netip.MustParseAddr(want) {
			t.Errorf("peerOf(%s) = %s, want %s", in, got, want)
		}
	}
}
