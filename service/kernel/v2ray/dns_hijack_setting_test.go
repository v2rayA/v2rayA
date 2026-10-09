package v2ray

import (
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func tunNoHijackSetting() *configure.Setting {
	s := tunSetting()
	s.DnsHijack = configure.No
	return s
}

// The tun inbound tells the core where to relay intercepted queries. With the
// opt-out on there is nothing to relay to, and the core builds no relay for
// an empty target.
func TestTunInboundHasNoDnsTargetWhenHijackIsOff(t *testing.T) {
	if got := tunDnsTarget(tunSetting()); got != "127.0.0.1:52353" {
		t.Fatalf("dns target with interception on: %q", got)
	}
	if got := tunDnsTarget(tunNoHijackSetting()); got != "" {
		t.Fatalf("dns target with interception off: %q", got)
	}
	setting := tunNoHijackSetting()
	if got := tunInboundSettings(setting).DnsTarget; got != "" {
		t.Fatalf("tun inbound settings carry a dns target: %q", got)
	}
}

// An older configuration has no dnsHijack key; the zero value must still mean
// interception on, so the target and the DNS module stay as they were.
func TestDnsHijackZeroValueKeepsInterception(t *testing.T) {
	setting := tunSetting()
	setting.DnsHijack = ""
	if !dnsInterceptionEnabled(setting) {
		t.Fatal("an unset dnsHijack must keep DNS interception on")
	}
	if got := tunDnsTarget(setting); got != "127.0.0.1:52353" {
		t.Fatalf("dns target for an unset dnsHijack: %q", got)
	}
}

// The DNS module is only generated while interception is on; nothing must
// reach the core's resolver when it is off.
func TestSetDNSGeneratesNothingWhenHijackIsOff(t *testing.T) {
	tmpl := baseTemplate(t, tunNoHijackSetting())
	if err := tmpl.setDNS(nil); err != nil {
		t.Fatal(err)
	}
	if tmpl.DnsModuleConfig != nil {
		t.Fatalf("the DNS module was configured anyway: %s", tmpl.DnsModuleConfig)
	}
}

// Without a relay the queries that still reach the device must leave the
// host, not join the transparent proxy chain and land on a proxy's port 53.
func TestTunRoutesPort53DirectlyWhenHijackIsOff(t *testing.T) {
	setting := tunNoHijackSetting()
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setTransparentRouting(); err != nil {
		t.Fatal(err)
	}
	rule, ok := findRoutingRule(t, tmpl, "transparent", "53")
	if !ok {
		t.Fatalf("no port-53 rule for the tun inbound: %+v", tmpl.Routing.Rules)
	}
	if rule.OutboundTag != "direct" {
		t.Fatalf("port 53 goes to %q, want direct", rule.OutboundTag)
	}
	if first := tmpl.Routing.Rules[0]; first.Port != "53" || first.OutboundTag != "direct" {
		t.Fatalf("the port-53 rule must come first so it wins over the mode's rules: %+v", first)
	}

	// With interception on the module answers port 53 itself, so the ordinary
	// routing rules are left alone.
	on := baseTemplate(t, tunSetting())
	if err := on.setTransparentRouting(); err != nil {
		t.Fatal(err)
	}
	if _, ok := findRoutingRule(t, on, "transparent", "53"); ok {
		t.Fatalf("a port-53 direct rule was added while interception is on: %+v", on.Routing.Rules)
	}
}

// The rule is specific to the tun inbound: redirect and tproxy never let port
// 53 reach the transparent chain at all, so they need nothing extra.
func TestPort53DirectRuleIsTunOnly(t *testing.T) {
	for _, mode := range []configure.TransparentType{
		configure.TransparentRedirect,
		configure.TransparentTproxy,
		configure.TransparentSystemProxy,
	} {
		setting := tunNoHijackSetting()
		setting.TransparentType = mode
		tmpl := baseTemplate(t, setting)
		if err := tmpl.setTransparentRouting(); err != nil {
			t.Fatal(err)
		}
		if _, ok := findRoutingRule(t, tmpl, "transparent", "53"); ok {
			t.Fatalf("%s mode gained a port-53 routing rule", mode)
		}
	}
}

func findRoutingRule(t *testing.T, tmpl *Template, inboundTag, port string) (coreObj.RoutingRule, bool) {
	t.Helper()
	for _, r := range tmpl.Routing.Rules {
		if r.Port != port || r.OutboundTag == "" {
			continue
		}
		for _, tag := range r.InboundTag {
			if tag == inboundTag {
				return r, true
			}
		}
	}
	return coreObj.RoutingRule{}, false
}

// The resolver hijack is what sends the system's queries at the DNS module
// in the first place, so with the opt-out off it must not run. Lite mode is
// turned off here because it disables the hijack for its own reasons, which
// would make the assertion pass without proving anything.
func TestResolverIsNotHijackedWhenHijackIsOff(t *testing.T) {
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

	if !ShouldLocalDnsListen() {
		t.Fatal("the resolver is not hijacked with the default setting")
	}
	stored.DnsHijack = configure.No
	if err := configure.SetSetting(stored); err != nil {
		t.Fatal(err)
	}
	if ShouldLocalDnsListen() {
		t.Fatal("the resolver is still hijacked with interception off")
	}
}
