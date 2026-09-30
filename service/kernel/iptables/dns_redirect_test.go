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
	for _, tc := range []struct {
		name, port, family string
		ipv6               bool
	}{
		{"IPv4", "52353", "ipv4", false},
		{"dual stack custom port", "5354", "{ ipv4, ipv6 }", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, deletes := 0, 0
			tableExists := true // Simulate a table left by a crashed process.
			executeCommands = func(command string, _ bool) error {
				if command != "nft delete table inet v2raya_dns 2>/dev/null || true" {
					t.Fatalf("unexpected setup command: %s", command)
				}
				deletes++
				tableExists = false
				return nil
			}
			executeCommandWithInput = func(bin string, args []string, rules string) error {
				calls++
				if tableExists {
					t.Fatal("existing DNS table was not removed before creation")
				}
				tableExists = true
				if bin != "nft" || !slices.Equal(args, []string{"-f", "-"}) {
					t.Fatalf("expected an atomic nft batch, got %s %v", bin, args)
				}
				_, chain, found := strings.Cut(rules, "chain output {")
				if !found || strings.Contains(rules, "chain prerouting {") {
					t.Fatalf("only local OUTPUT DNS should be redirected: %s", rules)
				}
				chain, _, _ = strings.Cut(chain, "\n    }")
				if !strings.Contains(chain, "type nat hook output priority dstnat - 1") {
					t.Fatalf("DNS needs NAT before the generic redirect: %s", chain)
				}
				bypass := strings.Index(chain, "meta mark & 0x80 == 0x80 return")
				redirect := strings.Index(chain, "meta nfproto "+tc.family+" meta l4proto { tcp, udp } th dport 53 redirect to :"+tc.port)
				if bypass < 0 || redirect < bypass {
					t.Fatalf("upstream bypass must precede TCP/UDP DNS interception: %s", chain)
				}
				return nil
			}
			for range 2 {
				if err := NftDNSRedirect(tc.port, tc.ipv6).Run(true); err != nil {
					t.Fatal(err)
				}
			}
			if calls != 2 || deletes != 2 {
				t.Fatalf("expected two idempotent installs, got %d batches and %d deletes", calls, deletes)
			}
		})
	}
}

func TestNftDNSRedirectReportsFailure(t *testing.T) {
	originalInput, originalCommands := executeCommandWithInput, executeCommands
	t.Cleanup(func() {
		executeCommandWithInput, executeCommands = originalInput, originalCommands
	})
	want := errors.New("nft NAT setup failed")
	deletes := 0
	executeCommands = func(command string, _ bool) error {
		if command != "nft delete table inet v2raya_dns 2>/dev/null || true" {
			t.Fatalf("unexpected cleanup command: %s", command)
		}
		deletes++
		return nil
	}
	executeCommandWithInput = func(string, []string, string) error { return want }
	if err := NftDNSRedirect("52353", false).Run(true); !errors.Is(err, want) {
		t.Fatalf("setup failure was lost: got %v, want %v", err, want)
	}
	if deletes != 2 {
		t.Fatalf("failed batch must be cleaned up, got %d deletes", deletes)
	}
}
