package v2ray

import (
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestNormalizeNodeDNS(t *testing.T) {
	setting := dnsModeSetting(configure.DnsModeService)
	setting.IpForward = false
	for input, want := range map[string]string{
		"": "auto", "auto": "auto", " 1.1.1.1 ": "udp://1.1.1.1:53",
		"1.1.1.1:5353":                        "udp://1.1.1.1:5353",
		"2001:db8::1":                         "udp://[2001:db8::1]:53",
		"tcp://[2001:db8::1]":                 "tcp://[2001:db8::1]:53",
		"tls://1.1.1.1":                       "tls://1.1.1.1:853",
		"https://[2001:db8::1]/dns-query?x=1": "https://[2001:db8::1]:443/dns-query?x=1",
		"udp://127.0.0.53":                    "udp://127.0.0.53:53",
	} {
		got, err := NormalizeNodeDNS(input, setting)
		if err != nil || got != want {
			t.Errorf("%q = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{
		"localhost", "tls://dns.example", "quic://1.1.1.1", "ftp://1.1.1.1",
		"0.0.0.0", "::", "224.0.0.1", "ff02::1", "1.1.1.1:0",
		"1.1.1.1:65536", "1.1.1.1:", "1.1.1.1:bad",
		"udp://1.1.1.1/path", "tcp://1.1.1.1?x=1", "https://u:p@1.1.1.1/dns-query",
		"https://1.1.1.1/dns-query#fragment", "udp://127.0.0.1:52353",
		"udp://::1:53", "tls://2001:db8::1:853",
	} {
		if _, err := NormalizeNodeDNS(input, setting); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	setting.PortSharing = true
	if _, err := NormalizeNodeDNS("tcp://127.0.0.53:52353", setting); err == nil {
		t.Fatal("accepted the wildcard listener via loopback")
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && !ip.IsLoopback() {
			if _, err := NormalizeNodeDNS("udp://"+net.JoinHostPort(ip.String(), "52353"), setting); err == nil {
				t.Errorf("accepted wildcard listener via local address %s", ip)
			}
		}
	}
	setting.DnsListenAddr = "[::]:5353"
	if _, err := NormalizeNodeDNS("udp://[::1]:5353", setting); err == nil {
		t.Fatal("accepted the IPv6 wildcard listener")
	}
}

func TestNormalizeNodeDNSRejectsInterceptionListeners(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	oldLite := env.Lite
	env.Lite = false
	t.Cleanup(func() { env.Lite = oldLite })
	setting := dnsModeSetting(configure.DnsModeHijack)
	var endpoints []string
	switch runtime.GOOS {
	case "linux":
		endpoints = []string{"udp://127.2.0.17:53", "tcp://127.2.0.17:52353", "udp://[::1]:52353"}
	case "darwin":
		endpoints = []string{"udp://127.0.0.1:53"}
	}
	for _, endpoint := range endpoints {
		if _, err := NormalizeNodeDNS(endpoint, setting); err == nil {
			t.Errorf("accepted interception listener %q", endpoint)
		}
	}
}

func TestSubscriptionNodeDNSConfig(t *testing.T) {
	setting := dnsModeSetting(configure.DnsModeService)
	infos := []serverInfo{
		{Info: &serverObj.V2Ray{Add: "auto.example"}},
		{Info: &serverObj.V2Ray{Add: "udp.example"}, NodeDns: "1.1.1.1:5353"},
		{Info: &serverObj.V2Ray{Add: "tcp.example"}, NodeDns: "tcp://1.1.1.1"},
		{Info: &serverObj.V2Ray{Add: "tls.example"}, NodeDns: "tls://1.1.1.1"},
		{Info: &serverObj.V2Ray{Add: "https.example"}, NodeDns: "https://[2001:db8::1]/dns-query?x=1"},
		{Info: &serverObj.V2Ray{Add: "other.example"}, NodeDns: "tcp://1.1.1.1:53"},
		{Info: &serverObj.V2Ray{Add: "192.0.2.1"}, NodeDns: "invalid"},
	}
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setDNS(infos); err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Upstreams []struct {
			ID, Addr, Protocol string
			ProxyTag           string `json:"proxy_tag"`
		}
		Rules []struct {
			ID, Upstream string
			Domain       []string
			Suffix       []string `json:"domain_suffix"`
		}
		Default string `json:"default_upstream"`
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	for i, want := range []struct{ addr, protocol string }{
		{"1.1.1.1:5353", "udp"}, {"1.1.1.1:53", "tcp"},
		{"1.1.1.1:853", "tcp-tls"}, {"https://[2001:db8::1]:443/dns-query?x=1", "https"},
	} {
		upstream := cfg.Upstreams[i]
		rule := cfg.Rules[i]
		if upstream.Addr != want.addr || upstream.Protocol != want.protocol || upstream.ProxyTag != "direct" || rule.Upstream != upstream.ID || len(rule.Domain) == 0 || len(rule.Suffix) != 0 {
			t.Fatalf("subscription %d: upstream=%+v rule=%+v", i, upstream, rule)
		}
	}
	if len(cfg.Rules[1].Domain) != 2 {
		t.Fatal("identical endpoints were not grouped")
	}
	if cfg.Upstreams[4].ID != "upstream-node" || cfg.Upstreams[4].Protocol != "udp" || cfg.Upstreams[4].Addr != "223.5.5.5:53" || cfg.Rules[4].Suffix[0] != "auto.example" {
		t.Fatal("auto no longer follows the original direct-first selection")
	}
	if cfg.Default == cfg.Upstreams[0].ID {
		t.Fatal("node upstream became the ordinary default")
	}
}

func TestSubscriptionNodeDNSPreservesExplicitHostnames(t *testing.T) {
	setting := dnsModeSetting(configure.DnsModeService)
	infos := []serverInfo{
		{Info: &serverObj.V2Ray{Add: "Node.Example."}, NodeDns: "1.1.1.1"},
		{Info: &serverObj.V2Ray{Add: "node.example"}, NodeDns: "udp://1.1.1.1:53"},
		{Info: &serverObj.V2Ray{Add: "Node.Example."}, NodeDns: "1.1.1.1"},
		{Info: &serverObj.V2Ray{Add: "Node.Example"}, NodeDns: "1.1.1.1"},
	}
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setDNS(infos); err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Rules []struct {
			ID     string
			Domain []string
		}
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	want := []string{"Node.Example.", "node.example", "Node.Example"}
	if len(cfg.Rules) == 0 || cfg.Rules[0].ID != "rule-node-subscription-0" || !reflect.DeepEqual(cfg.Rules[0].Domain, want) {
		t.Fatalf("explicit hostnames = %+v; want %v", cfg.Rules, want)
	}
}

func TestSubscriptionNodeDNSSharedDomain(t *testing.T) {
	setting := dnsModeSetting(configure.DnsModeService)
	infos := []serverInfo{
		{Info: &serverObj.V2Ray{Add: "Shared.Example."}, NodeDns: "1.1.1.1"},
		{Info: &serverObj.V2Ray{Add: "shared.example"}, NodeDns: "tcp://1.1.1.1"},
	}
	tmpl := baseTemplate(t, setting)
	var coded *common.CodedError
	if err := tmpl.setDNS(infos); !errors.As(err, &coded) || coded.Code != "NODE_DNS_CONFLICT" {
		t.Fatalf("conflict = %v", err)
	}
	infos[1].NodeDns = "udp://1.1.1.1:53"
	if err := tmpl.setDNS(infos); err != nil {
		t.Fatal(err)
	}
	infos[1].NodeDns = "auto"
	if err := tmpl.setDNS(infos); err != nil {
		t.Fatalf("automatic endpoint must defer to the explicit one: %v", err)
	}
	var cfg struct {
		Rules []struct {
			ID, Upstream string
			Domain       []string
			Suffix       []string `json:"domain_suffix"`
		}
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Rules) < 2 || len(cfg.Rules[0].Domain) != 1 || cfg.Rules[0].Domain[0] != "Shared.Example." {
		t.Fatalf("explicit rule did not come first: %+v", cfg.Rules)
	}
	if len(cfg.Rules[1].Suffix) != 1 || cfg.Rules[1].Suffix[0] != "shared.example" {
		t.Fatalf("automatic suffix rule missing: %+v", cfg.Rules)
	}
	setting.DnsMode = configure.DnsModeOff
	if err := tmpl.setDNS(infos); err != nil || tmpl.DnsModuleConfig != nil {
		t.Fatalf("off mode changed: %v", err)
	}
}

// With only auto endpoints the generated config must stay the one every
// earlier release produced: one direct node upstream and suffix rule in front
// of the user's rules, with the ordinary default upstream unchanged.
func TestSubscriptionNodeDNSAutoUnchanged(t *testing.T) {
	setting := dnsModeSetting(configure.DnsModeService)
	infos := []serverInfo{
		{Info: &serverObj.V2Ray{Add: "Node.Example."}},
		{Info: &serverObj.V2Ray{Add: "other.example"}, NodeDns: "auto"},
		{Info: &serverObj.V2Ray{Add: "Node.Example."}},
		{Info: &serverObj.V2Ray{Add: "192.0.2.1"}},
	}
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setDNS(infos); err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Upstreams []struct {
			ID, Addr, Protocol string
			ProxyTag           string `json:"proxy_tag"`
		}
		Rules []struct {
			ID, Upstream string
			Domain       []string
			Suffix       []string `json:"domain_suffix"`
		}
		Default string `json:"default_upstream"`
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Upstreams) != 4 || cfg.Upstreams[0].ID != "upstream-node" ||
		cfg.Upstreams[0].Addr != "223.5.5.5:53" || cfg.Upstreams[0].Protocol != "udp" {
		t.Fatalf("auto upstream = %+v", cfg.Upstreams)
	}
	for i, id := range []string{"upstream-0", "upstream-1", "upstream-2"} {
		if cfg.Upstreams[i+1].ID != id {
			t.Fatalf("user upstreams were reordered or renumbered: %+v", cfg.Upstreams)
		}
	}
	if len(cfg.Rules) != 3 || cfg.Rules[0].ID != "rule-node" ||
		cfg.Rules[0].Upstream != "upstream-node" || len(cfg.Rules[0].Domain) != 0 ||
		len(cfg.Rules[0].Suffix) != 2 || cfg.Rules[0].Suffix[0] != "Node.Example." ||
		cfg.Rules[0].Suffix[1] != "other.example" {
		t.Fatalf("auto rule = %+v", cfg.Rules)
	}
	if cfg.Rules[1].Upstream != "upstream-0" || cfg.Rules[2].Upstream != "upstream-1" || cfg.Default != "upstream-2" {
		t.Fatalf("user rules or default upstream changed: %+v default=%q", cfg.Rules, cfg.Default)
	}
}

func TestNormalizeNodeDNSErrorPreservesInput(t *testing.T) {
	for _, input := range []string{"localhost", " 0.0.0.0 ", "1.1.1.1:0"} {
		_, err := NormalizeNodeDNS(input, dnsModeSetting(configure.DnsModeService))
		var coded *common.CodedError
		if !errors.As(err, &coded) || coded.Code != "NODE_DNS_INVALID" || !strings.Contains(err.Error(), "invalid node DNS "+strconv.Quote(input)+":") {
			t.Errorf("input %q: error = %v", input, err)
		}
	}
}

func TestSubscriptionNodeDNSInvalidSavedValueFallsBack(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("uses the Linux hijack listener")
	}
	env := conf.GetEnvironmentConfig()
	oldLite := env.Lite
	env.Lite = false
	t.Cleanup(func() { env.Lite = oldLite })
	setting := dnsModeSetting(configure.DnsModeService)
	saved := "udp://127.2.0.17:53"
	if _, err := NormalizeNodeDNS(saved, setting); err != nil {
		t.Fatal(err)
	}
	setting.DnsMode = configure.DnsModeHijack
	if _, err := NormalizeNodeDNS(saved, setting); err == nil {
		t.Fatal("accepted the hijack listener")
	}
	for _, endpoint := range []string{saved, "invalid"} {
		t.Run(endpoint, func(t *testing.T) {
			infos := []serverInfo{{Info: &serverObj.V2Ray{Add: "Node.Example."}, NodeDns: endpoint}}
			tmpl := baseTemplate(t, setting)
			if err := tmpl.setDNS(infos); err != nil {
				t.Fatal(err)
			}
			auto := baseTemplate(t, setting)
			if err := auto.setDNS([]serverInfo{{Info: infos[0].Info, NodeDns: "auto"}}); err != nil {
				t.Fatal(err)
			}
			var got, want map[string]interface{}
			if err := json.Unmarshal(tmpl.DnsModuleConfig, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(auto.DnsModuleConfig, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("invalid saved endpoint did not produce the auto config: %s", tmpl.DnsModuleConfig)
			}
			if infos[0].NodeDns != endpoint {
				t.Fatal("saved endpoint was overwritten")
			}
		})
	}
}
