package v2ray

import (
	"errors"
	"testing"

	"github.com/v2rayA/v2rayA/kernel/iptables"
)

func TestRunDNSRedirectKeepsLegacyStartupBestEffort(t *testing.T) {
	want := errors.New("legacy NAT command failed")
	steps, cleaned := 0, false
	setter := iptables.Setter{
		PreFunc: func() error {
			steps++
			return want
		},
		AfterFunc: func() error {
			steps++
			return nil
		},
	}
	if err := runDNSRedirect(setter, false, func() { cleaned = true }); err != nil {
		t.Fatalf("legacy DNS failure must not abort core startup: %v", err)
	}
	if steps != 2 || !cleaned {
		t.Fatalf("legacy batch did not complete and clean partial state: steps=%d cleaned=%v", steps, cleaned)
	}
}

func TestRunDNSRedirectFailsNftStartup(t *testing.T) {
	want := errors.New("nft DNS setup failed")
	continued, cleaned := false, false
	setter := iptables.Setter{
		PreFunc: func() error { return want },
		AfterFunc: func() error {
			continued = true
			return nil
		},
	}
	if err := runDNSRedirect(setter, true, func() { cleaned = true }); !errors.Is(err, want) {
		t.Fatalf("nft failure must abort startup: %v", err)
	}
	if continued || cleaned {
		t.Fatalf("nft startup continued after failure: continued=%v cleaned=%v", continued, cleaned)
	}
}
