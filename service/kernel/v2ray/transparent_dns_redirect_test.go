package v2ray

import (
	"errors"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/iptables"
)

func TestDNSRedirectPolicyByPlatformAndMode(t *testing.T) {
	for _, tc := range []struct {
		name      string
		goos      string
		mode      configure.TransparentType
		autoRoute bool
		install   bool
		required  bool
	}{
		{"linux tproxy", "linux", configure.TransparentTproxy, false, true, true},
		{"linux redirect", "linux", configure.TransparentRedirect, false, true, true},
		{"linux system proxy", "linux", configure.TransparentSystemProxy, false, true, false},
		{"linux automatic tun", "linux", configure.TransparentTun, true, true, false},
		{"linux manual tun", "linux", configure.TransparentTun, false, false, false},
		{"macOS system proxy", "darwin", configure.TransparentSystemProxy, false, false, false},
		{"Windows system proxy", "windows", configure.TransparentSystemProxy, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setting := &configure.Setting{TransparentType: tc.mode, TunAutoRoute: tc.autoRoute}
			install, required := dnsRedirectPolicy(tc.goos, setting)
			if install != tc.install || required != tc.required {
				t.Fatalf("got install=%v required=%v, want install=%v required=%v", install, required, tc.install, tc.required)
			}
		})
	}
}

// With the opt-out on no platform installs the DNS REDIRECT rules; there is
// no DNS module for them to feed.
func TestDNSRedirectPolicyIsOffWithoutHijack(t *testing.T) {
	for _, tc := range []struct {
		goos      string
		mode      configure.TransparentType
		autoRoute bool
	}{
		{"linux", configure.TransparentTproxy, false},
		{"linux", configure.TransparentRedirect, false},
		{"linux", configure.TransparentSystemProxy, false},
		{"linux", configure.TransparentTun, true},
		{"linux", configure.TransparentTun, false},
		{"darwin", configure.TransparentTun, true},
		{"windows", configure.TransparentTun, true},
	} {
		setting := &configure.Setting{
			TransparentType: tc.mode,
			TunAutoRoute:    tc.autoRoute,
			DnsHijack:       configure.No,
		}
		if install, required := dnsRedirectPolicy(tc.goos, setting); install || required {
			t.Fatalf("%s %s: install=%v required=%v, want neither", tc.goos, tc.mode, install, required)
		}
	}
}

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
	if err := runDNSRedirect(setter, false, true, func() { cleaned = true }); err != nil {
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
	if err := runDNSRedirect(setter, true, true, func() { cleaned = true }); !errors.Is(err, want) {
		t.Fatalf("nft failure must abort startup: %v", err)
	}
	if continued || cleaned {
		t.Fatalf("nft startup continued after failure: continued=%v cleaned=%v", continued, cleaned)
	}
}

func TestRunDNSRedirectKeepsSystemProxyStartupBestEffort(t *testing.T) {
	want := errors.New("nft DNS setup failed")
	cleaned := false
	setter := iptables.Setter{PreFunc: func() error { return want }}
	if err := runDNSRedirect(setter, true, false, func() { cleaned = true }); err != nil {
		t.Fatalf("system proxy must start without nft DNS redirect: %v", err)
	}
	if cleaned {
		t.Fatal("nft setup already cleans its own table; legacy cleanup must not run")
	}
}
