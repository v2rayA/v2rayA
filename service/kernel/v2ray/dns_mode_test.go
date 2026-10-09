package v2ray

import (
	"encoding/json"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

// The mode decides two separate things: whether the DNS module is configured
// at all, and whether the host is rewired to send its queries there. Every
// mode has to land on its own pair, so they are checked side by side.

func dnsModeSetting(mode configure.DnsMode) *configure.Setting {
	setting := tunSetting()
	setting.DnsMode = mode
	return setting
}

func dnsModuleListener(t *testing.T, setting *configure.Setting) (string, []string, bool) {
	t.Helper()
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setDNS(nil); err != nil {
		t.Fatal(err)
	}
	if tmpl.DnsModuleConfig == nil {
		return "", nil, false
	}
	var cfg struct {
		Listener struct {
			ListenAddr string   `json:"listen_addr"`
			Extra      []string `json:"extra_listen_addrs"`
		} `json:"listener"`
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg.Listener.ListenAddr, cfg.Listener.Extra, true
}

// Service mode runs the module without hijack: the module must be there, and
// the resolver-facing addresses the REDIRECT rules land on must not be, since
// nothing redirects to them.
func TestDnsServiceModeRunsTheModuleWithoutTheHijackAddresses(t *testing.T) {
	setting := dnsModeSetting(configure.DnsModeService)
	listen, extra, ok := dnsModuleListener(t, setting)
	if !ok {
		t.Fatal("service mode must configure the DNS module")
	}
	if listen == "" {
		t.Fatal("service mode must leave the module an address to answer on")
	}
	if len(extra) != 0 {
		t.Fatalf("service mode binds the redirect landing addresses %v", extra)
	}
}

// Hijack keeps every address the historical behaviour installed.
func TestDnsHijackModeKeepsTheRedirectLandingAddresses(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	oldLite := env.Lite
	env.Lite = false
	t.Cleanup(func() { env.Lite = oldLite })

	setting := dnsModeSetting(configure.DnsModeHijack)
	_, extra, ok := dnsModuleListener(t, setting)
	if !ok {
		t.Fatal("hijack mode must configure the DNS module")
	}
	if !contains(extra, "127.2.0.17:53") {
		t.Fatalf("hijack mode lost the hijacked resolver's address: %v", extra)
	}
}

// Off is the only mode that leaves the core without a resolver at all.
func TestDnsOffModeConfiguresNoModule(t *testing.T) {
	if _, _, ok := dnsModuleListener(t, dnsModeSetting(configure.DnsModeOff)); ok {
		t.Fatal("off mode must not configure the DNS module")
	}
}

// The TUN tells the core where to relay an intercepted query. Only hijack
// mode has a relay, so service mode leaves the field empty exactly as off
// does; the module's own listener is a different thing.
func TestOnlyHijackModeGivesTheTunADnsTarget(t *testing.T) {
	for _, tc := range []struct {
		mode configure.DnsMode
		want string
	}{
		{configure.DnsModeOff, ""},
		{configure.DnsModeService, ""},
		{configure.DnsModeHijack, "127.0.0.1:52353"},
	} {
		if got := tunDnsTarget(dnsModeSetting(tc.mode)); got != tc.want {
			t.Fatalf("%s mode dns target = %q, want %q", tc.mode, got, tc.want)
		}
	}
}

// Without a relay the queries that still reach the TUN device must leave the
// host directly instead of joining the transparent proxy chain and landing on
// a proxy's port 53. That holds for every mode but hijack, so the rule keys
// on the interception and not on whether the module happens to be running.
func TestPort53LeavesTheHostDirectlyInEveryNonHijackMode(t *testing.T) {
	for _, mode := range []configure.DnsMode{configure.DnsModeOff, configure.DnsModeService} {
		tmpl := baseTemplate(t, dnsModeSetting(mode))
		if err := tmpl.setTransparentRouting(); err != nil {
			t.Fatal(err)
		}
		rule, ok := findRoutingRule(t, tmpl, "transparent", "53")
		if !ok {
			t.Fatalf("%s mode: no port-53 rule for the tun inbound: %+v", mode, tmpl.Routing.Rules)
		}
		if rule.OutboundTag != "direct" {
			t.Fatalf("%s mode: port 53 goes to %q, want direct", mode, rule.OutboundTag)
		}
		if first := tmpl.Routing.Rules[0]; first.Port != "53" || first.OutboundTag != "direct" {
			t.Fatalf("%s mode: the port-53 rule must come first: %+v", mode, first)
		}
	}

	hijacked := baseTemplate(t, dnsModeSetting(configure.DnsModeHijack))
	if err := hijacked.setTransparentRouting(); err != nil {
		t.Fatal(err)
	}
	if _, ok := findRoutingRule(t, hijacked, "transparent", "53"); ok {
		t.Fatalf("hijack mode gained a port-53 rule: %+v", hijacked.Routing.Rules)
	}
}

// Service mode leaves the firewall alone: there is nothing for the module to
// catch, and installing the rules would redirect the host's queries to a
// listener the mode was chosen not to expose.
func TestServiceModeInstallsNoDnsRedirectRules(t *testing.T) {
	for _, mode := range []configure.TransparentType{
		configure.TransparentRedirect,
		configure.TransparentTproxy,
		configure.TransparentTun,
		configure.TransparentSystemProxy,
	} {
		setting := dnsModeSetting(configure.DnsModeService)
		setting.TransparentType = mode
		setting.TunAutoRoute = true
		if install, required := dnsRedirectPolicy("linux", setting); install || required {
			t.Fatalf("%s mode: install=%v required=%v, want neither", mode, install, required)
		}
	}
}

// The resolver hijack is what sends the system's queries at the module in the
// first place, so only hijack mode may rewrite /etc/resolv.conf. Lite mode is
// turned off because it disables the hijack for its own reasons.
func TestOnlyHijackModeHijacksTheResolver(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	oldLite := env.Lite
	env.Lite = false
	t.Cleanup(func() { env.Lite = oldLite })

	stored := configure.NewSetting()
	stored.Transparent = configure.TransparentProxy
	if err := configure.SetSetting(stored); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.SetSetting(configure.NewSetting()) })

	for _, tc := range []struct {
		mode configure.DnsMode
		want bool
	}{
		{configure.DnsModeOff, false},
		{configure.DnsModeService, false},
		{configure.DnsModeHijack, true},
	} {
		stored.DnsMode = tc.mode
		if err := configure.SetSetting(stored); err != nil {
			t.Fatal(err)
		}
		if got := ShouldLocalDnsListen(); got != tc.want {
			t.Fatalf("%s mode: ShouldLocalDnsListen = %v, want %v", tc.mode, got, tc.want)
		}
	}
}

// A setting stored before the mode existed keeps whatever the opt-out said,
// whichever way the core is reached: the module predicate and the
// interception predicate must both read the same resolved mode.
func TestLegacyOptOutStillDecidesBothPredicates(t *testing.T) {
	off := tunSetting()
	off.DnsMode = ""
	off.DnsHijack = configure.No
	if dnsServiceEnabled(off) || dnsInterceptionEnabled(off) {
		t.Fatal("a legacy opt-out must leave the module and the interception off")
	}

	on := tunSetting()
	on.DnsMode = ""
	on.DnsHijack = configure.Yes
	if !dnsServiceEnabled(on) || !dnsInterceptionEnabled(on) {
		t.Fatal("a legacy opt-in must keep the module and the interception on")
	}
}
