package iptables

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestNftDNSRedirectWithoutIptables(t *testing.T) {
	originalInput, originalCommands := executeCommandWithInput, executeCommands
	t.Cleanup(func() {
		executeCommandWithInput, executeCommands = originalInput, originalCommands
	})
	executeCommands = func(string, bool) error {
		t.Fatal("native DNS setup must not execute iptables commands")
		return nil
	}
	for _, tc := range []struct {
		name, port, family string
		ipv6               bool
	}{
		{"IPv4", "52353", "ipv4", false},
		{"dual stack custom port", "5354", "{ ipv4, ipv6 }", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			executeCommandWithInput = func(bin string, args []string, rules string) error {
				calls++
				if bin != "nft" || !slices.Equal(args, []string{"-f", "-"}) {
					t.Fatalf("expected an atomic nft batch, got %s %v", bin, args)
				}
				for _, hook := range []string{"output", "prerouting"} {
					_, chain, found := strings.Cut(rules, "chain "+hook+" {")
					if !found {
						t.Fatalf("missing %s DNS hook", hook)
					}
					chain, _, _ = strings.Cut(chain, "\n    }")
					if !strings.Contains(chain, "type nat hook "+hook+" priority dstnat - 1") {
						t.Fatalf("DNS needs NAT before the generic redirect: %s", chain)
					}
					bypass := strings.Index(chain, "meta mark & 0x80 == 0x80 return")
					redirect := strings.Index(chain, "meta nfproto "+tc.family+" meta l4proto { tcp, udp } th dport 53 redirect to :"+tc.port)
					if bypass < 0 || redirect < bypass {
						t.Fatalf("upstream bypass must precede TCP/UDP DNS interception: %s", chain)
					}
				}
				return nil
			}
			if err := NftDNSRedirect(tc.port, tc.ipv6).Run(true); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("expected one atomic batch, got %d", calls)
			}
		})
	}
}

func TestNftDNSRedirectReportsFailure(t *testing.T) {
	original := executeCommandWithInput
	t.Cleanup(func() { executeCommandWithInput = original })
	want := errors.New("nft NAT setup failed")
	executeCommandWithInput = func(string, []string, string) error { return want }
	if err := NftDNSRedirect("52353", false).Run(true); !errors.Is(err, want) {
		t.Fatalf("setup failure was lost: got %v, want %v", err, want)
	}
}
