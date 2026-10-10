package v2ray

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestNodeDNSDiscoveryPriority(t *testing.T) {
	rules := []configure.DnsRule{
		{Server: "1.1.1.1", Outbound: "proxy"},
		{Server: "tls://1.1.1.1", Outbound: "custom"},
		{Server: "udp://1.1.1.1:53", Outbound: "direct"},
		{Upstream: "tcp://1.1.1.1", Server: "8.8.8.8", Outbound: "direct"},
		{Server: "https://[2001:db8::1]/a?x=1", Outbound: "proxy"},
		{Server: "https://[2001:db8::1]/b?x=1", Outbound: "proxy"},
		{Server: "9.9.9.9", Action: "reject", Outbound: "direct"},
		{Server: "8.8.8.8", Outbound: "block"},
		{Server: "tls://dns.google", Outbound: "direct"},
		{Server: "localhost", Outbound: "proxy"},
	}
	options := collectNodeDNSOptions(rules, []string{"1.1.1.1:53", "127.0.0.53:53", "127.2.0.17:53"}, []string{"127.2.0.17:53"}, []string{"udp://1.1.1.1:53", "tcp://119.29.29.29:53"})
	want := []NodeDNSOption{
		{"auto", "udp://1.1.1.1:53", "auto"},
		{"udp://1.1.1.1:53", "udp://1.1.1.1:53", "direct"},
		{"tcp://1.1.1.1:53", "tcp://1.1.1.1:53", "direct"},
		{"udp://127.0.0.53:53", "udp://127.0.0.53:53", "localhost"},
		{"tls://1.1.1.1:853", "tls://1.1.1.1:853", "proxy"},
		{"https://[2001:db8::1]:443/a?x=1", "https://[2001:db8::1]:443/a?x=1", "proxy"},
		{"https://[2001:db8::1]:443/b?x=1", "https://[2001:db8::1]:443/b?x=1", "proxy"},
		{"tcp://119.29.29.29:53", "tcp://119.29.29.29:53", "fallback"},
	}
	if !reflect.DeepEqual(options.Options, want) {
		t.Fatalf("got %+v; want %+v", options.Options, want)
	}
	for i := 0; i < 5; i++ {
		e, err := options.Select("auto")
		if err != nil || e.URL != want[0].URL {
			t.Fatalf("auto=%+v, %v", e, err)
		}
	}
	if _, err := options.Select("udp://9.9.9.9:53"); err == nil {
		t.Fatal("stale source accepted")
	}
	// Removing direct only changes the category if a system source survives.
	options = collectNodeDNSOptions(nil, []string{"1.1.1.1:53"}, nil, nil)
	if _, err := options.Select("1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	options = collectNodeDNSOptions([]configure.DnsRule{{Server: "119.29.29.29", Outbound: "proxy"}}, nil, nil, []string{"udp://119.29.29.29:53", "tcp://119.29.29.29:53"})
	if options.Options[0].URL != "tcp://119.29.29.29:53" {
		t.Fatalf("auto selected proxy duplicate: %+v", options)
	}
	options = collectNodeDNSOptions([]configure.DnsRule{{Server: "119.29.29.29", Outbound: "proxy"}}, nil, nil, []string{"udp://119.29.29.29:53"})
	if len(options.Options) != 1 || options.Options[0].Category != "proxy" {
		t.Fatal("unavailable auto returned")
	}
}

func TestNodeDNSExcludesListeners(t *testing.T) {
	for _, tc := range []struct {
		address   string
		listeners []string
		want      bool
	}{
		{"127.0.0.53:53", []string{"127.0.0.1:52353", "127.2.0.17:53"}, false},
		{"127.2.0.17:53", []string{"127.2.0.17:53"}, true},
		{"[::1]:5353", []string{"[::]:5353"}, true},
		{"127.0.0.1:5353", []string{"0.0.0.0:5353"}, true},
		{"127.0.0.1:53", []string{"0.0.0.0:5353"}, false},
		{"1.1.1.1:5353", []string{"0.0.0.0:5353"}, false},
	} {
		e, _ := resolv.ParseIPDNS(tc.address)
		if got := isOwnDNS(e, tc.listeners); got != tc.want {
			t.Errorf("%s %v: got %v", tc.address, tc.listeners, got)
		}
	}
}

func TestNodeDNSLatencyConfigPreservesProtocolAndPriority(t *testing.T) {
	for _, scheme := range []string{"udp", "tcp", "tls", "https"} {
		t.Run(scheme, func(t *testing.T) {
			e, _ := resolv.ParseIPDNS(scheme + "://1.1.1.1")
			tmpl := &Template{Setting: configure.NewSetting(), NodeDNS: e, DnsModuleConfig: json.RawMessage(`{"upstreams":[{"id":"normal","addr":"8.8.8.8:53"}],"rules":[],"default_upstream":"normal"}`)}
			servers := []serverObj.ServerObj{&serverObj.V2Ray{Add: "new-node.example"}, &serverObj.V2Ray{Add: "1.2.3.4"}}
			if err := tmpl.AddNodeDNSDomains(servers); err != nil {
				t.Fatal(err)
			}
			var cfg struct {
				Upstreams []struct {
					ID, Addr, Protocol string
					ProxyTag           string `json:"proxy_tag"`
				}
				Rules []struct {
					ID      string
					Domains []string `json:"domain_suffix"`
				}
				Default string `json:"default_upstream"`
			}
			if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.Upstreams[0].ID != "upstream-node" || cfg.Upstreams[0].Addr != e.UpstreamAddress() || cfg.Upstreams[0].Protocol != e.Protocol() || cfg.Upstreams[0].ProxyTag != "direct" || cfg.Default != "normal" {
				t.Fatalf("wrong node upstream: %+v", cfg)
			}
			if !reflect.DeepEqual(cfg.Rules[0].Domains, []string{"new-node.example"}) {
				t.Fatalf("wrong node domains: %+v", cfg.Rules)
			}
			if len(tmpl.DNS.Servers) != 2 {
				t.Fatalf("unexpected Xray DNS servers: %+v", tmpl.DNS.Servers)
			}
			node := tmpl.DNS.Servers[0].(coreObj.DnsServer)
			if node.Address != "tcp+local://"+dnsModuleListenAddr(tmpl.Setting) || !node.FinalQuery || !node.SkipFallback || !node.DisableCache || !reflect.DeepEqual(node.Domains, []string{"full:new-node.example"}) {
				t.Fatalf("Xray does not query the module exclusively for nodes: %+v", node)
			}
			if servers[0].GetHostname() != "new-node.example" || net.ParseIP(servers[1].GetHostname()) == nil {
				t.Fatal("server address rewritten")
			}
		})
	}
}

func TestNodeDNSAppliesWithoutTakeover(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	wasLite := env.Lite
	t.Cleanup(func() { env.Lite = wasLite })
	selected, _ := resolv.ParseIPDNS("tls://192.0.2.53")
	for _, lite := range []bool{false, true} {
		env.Lite = lite
		for _, transparent := range []configure.TransparentMode{configure.TransparentClose, configure.TransparentProxy} {
			setting := configure.NewSetting()
			setting.Transparent = transparent
			setting.TransparentType = configure.TransparentTun
			setting.TunAutoRoute = false
			tmpl := &Template{Setting: setting, NodeDNS: selected}
			if err := tmpl.generateDnsModuleConfig(nil); err != nil {
				t.Fatal(err)
			}
			var cfg struct {
				Upstreams []struct{ ID, Addr, Protocol, ProxyTag string }
			}
			if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
				t.Fatal(err)
			}
			node := cfg.Upstreams[0]
			if node.ID != "upstream-node" || node.Addr != selected.Address() || node.Protocol != "tcp-tls" {
				t.Fatalf("lite=%v transparent=%v: selected endpoint lost: %+v", lite, transparent, cfg)
			}
		}
	}
}

func TestNodeDNSOffDoesNotConfigureModule(t *testing.T) {
	setting := configure.NewSetting()
	setting.DnsMode = configure.DnsModeOff
	endpoint, _ := resolv.ParseIPDNS("tls://192.0.2.53")
	tmpl := &Template{Setting: setting, NodeDNS: endpoint}
	servers := []serverObj.ServerObj{&serverObj.V2Ray{Add: "node.example"}}
	if err := tmpl.setDNS([]serverInfo{{Info: servers[0]}}); err != nil {
		t.Fatal(err)
	}
	if tmpl.DnsModuleConfig != nil || tmpl.NodeDNS != nil || tmpl.DNS != nil {
		t.Fatal("off mode configured node DNS")
	}
	tmpl.NodeDNS = endpoint
	if err := tmpl.AddNodeDNSDomains(servers); err != nil {
		t.Fatal(err)
	}
	if tmpl.DnsModuleConfig != nil || tmpl.NodeDNS != nil || tmpl.DNS != nil {
		t.Fatal("latency configuration enabled DNS in off mode")
	}
}

func TestNodeDNSSystemSource(t *testing.T) {
	original := "nameserver 127.0.0.53\nnameserver 2001:db8::53\n"
	for _, tc := range []struct {
		name, content, backup string
		missing               bool
	}{
		{name: "original", content: original},
		{name: "hijacked", content: HijackFlag + "\nnameserver 127.2.0.17\nnameserver 223.5.5.5\n", backup: original},
		{name: "missing", missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path, backup := filepath.Join(dir, "resolv.conf"), filepath.Join(dir, "backup")
			withPaths(t, path, backup)
			if !tc.missing {
				if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.backup != "" {
				if err := os.WriteFile(backup, []byte(tc.backup), 0600); err != nil {
					t.Fatal(err)
				}
			}
			options := CollectNodeDNSOptions(configure.NewSetting(), []configure.DnsRule{{Server: "192.0.2.53", Outbound: "direct"}})
			var system []string
			for _, option := range options.Options {
				if option.Category == "localhost" {
					system = append(system, option.URL)
				}
			}
			var wantSystem, wantWarnings []string
			if tc.missing {
				wantWarnings = []string{"NODE_DNS_SYSTEM_READ_FAILED"}
			} else {
				wantSystem = []string{"udp://127.0.0.53:53", "udp://[2001:db8::53]:53"}
			}
			if !reflect.DeepEqual(system, wantSystem) || !reflect.DeepEqual(options.Warnings, wantWarnings) {
				t.Fatalf("system=%v warnings=%v; want %v %v", system, options.Warnings, wantSystem, wantWarnings)
			}
			selected, err := options.Select("auto")
			if err != nil || selected.URL != "udp://192.0.2.53:53" {
				t.Fatalf("auto=%v err=%v", selected, err)
			}
		})
	}
}

func TestNodeDNSStartFallsBackWithoutChangingSelection(t *testing.T) {
	previous := configure.GetSettingNotNil()
	t.Cleanup(func() { _ = configure.SetSetting(previous) })
	dir := t.TempDir()
	path := filepath.Join(dir, "resolv.conf")
	withPaths(t, path, filepath.Join(dir, "backup"))
	if err := os.WriteFile(path, []byte("nameserver 192.0.2.54\n"), 0600); err != nil {
		t.Fatal(err)
	}
	rules := []configure.DnsRule{{Server: "localhost", Outbound: "direct"}}
	setting := configure.NewSetting()
	setting.DnsMode = configure.DnsModeService
	setting.NodeDns = "udp://192.0.2.53:53"
	setting.LogLevel = "debug"
	if err := configure.SetSetting(setting); err != nil {
		t.Fatal(err)
	}
	setting = configure.GetSettingNotNil()
	snapshot := *setting
	if _, err := SelectNodeDNS(setting, rules); err == nil {
		t.Fatal("settings validation accepted a removed source")
	}
	endpoint, err := SelectNodeDNSForStart(setting, rules)
	if err != nil || endpoint.URL != "udp://192.0.2.54:53" {
		t.Fatalf("startup did not use the new system source: %v, %v", endpoint, err)
	}
	stored := configure.GetSettingNotNil()
	if !reflect.DeepEqual(setting, &snapshot) || !reflect.DeepEqual(stored, &snapshot) {
		t.Fatalf("startup changed the saved selection: %+v", stored)
	}
	if err := os.WriteFile(path, []byte("nameserver 192.0.2.55\n"), 0600); err != nil {
		t.Fatal(err)
	}
	endpoint, err = SelectNodeDNSForStart(stored, rules)
	if err != nil || endpoint.URL != "udp://192.0.2.55:53" || stored.NodeDns != "udp://192.0.2.53:53" {
		t.Fatalf("auto did not follow the next system change: %v, %v", endpoint, err)
	}
	// A surviving source from another category still permits the explicit URL.
	endpoint, err = SelectNodeDNSForStart(stored, []configure.DnsRule{{Server: "192.0.2.53", Outbound: "proxy"}})
	if err != nil || endpoint.URL != stored.NodeDns {
		t.Fatalf("surviving explicit source was replaced: %v, %v", endpoint, err)
	}
	if err := os.WriteFile(path, []byte("nameserver 192.0.2.53\n"), 0600); err != nil {
		t.Fatal(err)
	}
	endpoint, err = SelectNodeDNSForStart(stored, rules)
	if err != nil || endpoint.URL != stored.NodeDns || !reflect.DeepEqual(configure.GetSettingNotNil(), stored) {
		t.Fatalf("returning explicit source was not restored: %v, %v", endpoint, err)
	}
	stored.DnsMode = configure.DnsModeOff
	endpoint, err = SelectNodeDNSForStart(stored, rules)
	if err != nil || endpoint != nil || stored.NodeDns != "udp://192.0.2.53:53" {
		t.Fatal("off mode changed the unused selection")
	}
}
