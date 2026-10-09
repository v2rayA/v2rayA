package iptables

import (
	"os"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
)

// legacyRules renders the legacy iptables rule text. The setter's own Run
// would execute it; the whitelist steps are plain command text, so the
// interception can be read without a kernel.
func legacyRules(t *testing.T, dnsHijack bool, run func(bool) Setter) string {
	t.Helper()
	oldAvailable := commandAvailable
	oldCommands := executeCommands
	oldIPv6 := legacyIPv6Supported
	commandAvailable = func(string) bool { return true }
	legacyIPv6Supported = func() bool { return true }
	var text strings.Builder
	executeCommands = func(commands string, _ bool) error {
		text.WriteString(commands)
		text.WriteString("\n")
		return nil
	}
	t.Cleanup(func() {
		commandAvailable = oldAvailable
		executeCommands = oldCommands
		legacyIPv6Supported = oldIPv6
	})
	// The whitelist markers are not commands; without them the setter would
	// still run, and the rule text is what the assertions read.
	if err := run(dnsHijack).Run(false); err != nil {
		t.Fatal(err)
	}
	return text.String()
}

func legacyRedirectRules(t *testing.T, dnsHijack bool) string {
	t.Helper()
	return legacyRules(t, dnsHijack, (&legacyRedirect{}).GetSetupCommands)
}

func legacyTproxyRules(t *testing.T, dnsHijack bool) string {
	t.Helper()
	return legacyRules(t, dnsHijack, (&legacyTproxy{}).GetSetupCommands)
}

// nftTable returns the table an nft setter wrote: it hands `nft -f` a file
// path, so the rules are read back from there instead of executed.
func nftTable(t *testing.T, run func(bool) Setter, dnsHijack bool) string {
	t.Helper()
	oldAvailable := commandAvailable
	oldCommands := executeCommands
	oldConfig := conf.GetEnvironmentConfig().Config
	commandAvailable = func(string) bool { return true }
	executeCommands = func(string, bool) error { return nil }
	conf.GetEnvironmentConfig().Config = t.TempDir()
	t.Cleanup(func() {
		commandAvailable = oldAvailable
		executeCommands = oldCommands
		conf.GetEnvironmentConfig().Config = oldConfig
	})
	if err := run(dnsHijack).Run(false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(asset.GetNftablesConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func nftRedirectTable(t *testing.T, dnsHijack bool) string {
	t.Helper()
	return nftTable(t, (&nftRedirect{}).GetSetupCommands, dnsHijack)
}

func nftTproxyTable(t *testing.T, dnsHijack bool) string {
	t.Helper()
	return nftTable(t, (&nftTproxy{}).GetSetupCommands, dnsHijack)
}

// With interception on, port 53 is redirected to the DNS module for both
// families and nothing returns it from the transparent chain first.
func TestLegacyRedirectInterceptsDNSByDefault(t *testing.T) {
	rules := legacyRedirectRules(t, true)
	for _, want := range []string{
		"iptables -w 2 -t nat -N DNS_REDIRECT",
		"iptables -w 2 -t nat -A DNS_REDIRECT -p udp -j REDIRECT --to-port 52353",
		"iptables -w 2 -t nat -I OUTPUT -p udp --dport 53 -j DNS_REDIRECT",
		"ip6tables -w 2 -t nat -I PREROUTING -p tcp --dport 53 -j DNS_REDIRECT",
	} {
		if !strings.Contains(rules, want) {
			t.Fatalf("DNS interception is missing %q:\n%s", want, rules)
		}
	}
	if strings.Contains(rules, "TP_RULE -p udp --dport 53 -j RETURN") {
		t.Fatalf("port 53 must not bypass the transparent chain while it is intercepted:\n%s", rules)
	}
}

// With the opt-out off the chain, its hooks and the module target are gone
// and port 53 returns from TP_RULE rather than being redirected to the
// transparent proxy's port.
func TestLegacyRedirectLetsDNSPassWhenHijackIsOff(t *testing.T) {
	rules := legacyRedirectRules(t, false)
	if strings.Contains(rules, "DNS_REDIRECT") {
		t.Fatalf("the DNS redirect chain must not be installed:\n%s", rules)
	}
	for _, want := range []string{
		"iptables -w 2 -t nat -A TP_RULE -p udp --dport 53 -j RETURN",
		"iptables -w 2 -t nat -A TP_RULE -p tcp --dport 53 -j RETURN",
		"ip6tables -w 2 -t nat -A TP_RULE -p udp --dport 53 -j RETURN",
		"ip6tables -w 2 -t nat -A TP_RULE -p tcp --dport 53 -j RETURN",
	} {
		if !strings.Contains(rules, want) {
			t.Fatalf("the DNS bypass is missing %q:\n%s", want, rules)
		}
	}
	// Everything else, the generic TCP redirect included, is unchanged.
	if !strings.Contains(rules, "iptables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 52345") {
		t.Fatalf("the transparent proxy itself was disabled:\n%s", rules)
	}
}

func TestNftRedirectInterceptsDNSByDefault(t *testing.T) {
	table := nftRedirectTable(t, true)
	for _, want := range []string{
		"chain dns_redirect",
		"meta l4proto { tcp, udp } th dport 53 redirect to :52353",
		"meta l4proto { tcp, udp } th dport 53 jump dns_redirect",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("DNS interception is missing %q:\n%s", want, table)
		}
	}
	if strings.Contains(table, "th dport 53 return") {
		t.Fatalf("port 53 must not bypass the transparent chain while it is intercepted:\n%s", table)
	}
}

func TestNftRedirectLetsDNSPassWhenHijackIsOff(t *testing.T) {
	table := nftRedirectTable(t, false)
	if strings.Contains(table, "dns_redirect") || strings.Contains(table, "redirect to :52353") {
		t.Fatalf("port 53 must not be redirected to the DNS module:\n%s", table)
	}
	for _, want := range []string{
		"meta l4proto { tcp, udp } th dport 53 return",
		"meta l4proto tcp redirect to :52345",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("the nft redirect is missing %q:\n%s", want, table)
		}
	}
}

func TestLegacyTproxyInterceptsDNSByDefault(t *testing.T) {
	rules := legacyTproxyRules(t, true)
	for _, want := range []string{
		"iptables -w 2 -t mangle -N DNS_MARK",
		"iptables -w 2 -t mangle -I OUTPUT -p udp --dport 53 -j DNS_MARK",
		"iptables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j TP_MARK",
		"ip6tables -w 2 -t mangle -A TP_PRE -p udp -m mark --mark 0x40/0xc0 --dport 53 -j TPROXY --on-port 52353 --on-ip ::1",
	} {
		if !strings.Contains(rules, want) {
			t.Fatalf("DNS interception is missing %q:\n%s", want, rules)
		}
	}
}

func TestLegacyTproxyLetsDNSPassWhenHijackIsOff(t *testing.T) {
	rules := legacyTproxyRules(t, false)
	for _, gone := range []string{"DNS_MARK", "--on-port 52353", "--dport 52353"} {
		if strings.Contains(rules, gone) {
			t.Fatalf("%q must not be installed while interception is off:\n%s", gone, rules)
		}
	}
	for _, want := range []string{
		"iptables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j RETURN",
		"iptables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j RETURN",
		"ip6tables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j RETURN",
		"ip6tables -w 2 -t mangle -A TP_RULE -p tcp --dport 53 -j RETURN",
	} {
		if !strings.Contains(rules, want) {
			t.Fatalf("the DNS bypass is missing %q:\n%s", want, rules)
		}
	}
	if !strings.Contains(rules, "iptables -w 2 -t mangle -A TP_PRE -p tcp -m mark --mark 0x40/0xc0 -j TPROXY --on-port 52345") {
		t.Fatalf("the transparent proxy itself was disabled:\n%s", rules)
	}
	bypass := strings.Index(rules, "iptables -w 2 -t mangle -A TP_RULE -p udp --dport 53 -j RETURN")
	restore := strings.Index(rules, "iptables -w 2 -t mangle -A TP_RULE -j CONNMARK --restore-mark")
	if bypass < 0 || restore < 0 || bypass > restore {
		t.Fatalf("the DNS bypass must precede connmark restoration:\n%s", rules)
	}
}

func TestNftTproxyInterceptsDNSByDefault(t *testing.T) {
	table := nftTproxyTable(t, true)
	for _, want := range []string{
		"chain dns_mark",
		"meta l4proto { tcp, udp } th dport 53 jump dns_mark",
		"mark & 0xc0 == 0x40 th dport 53 tproxy ip to 127.2.0.17:52353",
		"meta l4proto { tcp, udp } th dport 53 jump tp_mark",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("DNS interception is missing %q:\n%s", want, table)
		}
	}
	start := strings.Index(table, "chain dns_mark {")
	if start < 0 {
		t.Fatalf("could not find the DNS marking chain:\n%s", table)
	}
	end := strings.Index(table[start:], "\n    chain output")
	if end < 0 {
		t.Fatalf("could not isolate the DNS marking chain:\n%s", table)
	}
	markChain := table[start : start+end]
	if !strings.Contains(markChain, "meta mark set mark | 0x40\n        return") || strings.Contains(markChain, "accept") {
		t.Fatalf("DNS marking must return to the base chain so forwarded queries reach the DNS TPROXY rule:\n%s", table)
	}
}

func TestNftTproxyLetsDNSPassWhenHijackIsOff(t *testing.T) {
	table := nftTproxyTable(t, false)
	for _, gone := range []string{"dns_mark", "52353"} {
		if strings.Contains(table, gone) {
			t.Fatalf("%q must not be installed while interception is off:\n%s", gone, table)
		}
	}
	for _, want := range []string{
		"meta l4proto { tcp, udp } th dport 53 return",
		"tproxy ip to 127.0.0.1:52345",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("the nft tproxy is missing %q:\n%s", want, table)
		}
	}
	bypass := strings.Index(table, "meta l4proto { tcp, udp } th dport 53 return")
	restore := strings.Index(table, "meta mark set ct mark")
	if bypass < 0 || restore < 0 || bypass > restore {
		t.Fatalf("the DNS bypass must precede conntrack mark restoration:\n%s", table)
	}
}
