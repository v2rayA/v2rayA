package v2ray

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
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
			if servers[0].GetHostname() != "new-node.example" || net.ParseIP(servers[1].GetHostname()) == nil {
				t.Fatal("server address rewritten")
			}
		})
	}
}

func TestNodeDNSOnlyAppliesToPlannedTakeover(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux DNS redirect policy")
	}
	env := conf.GetEnvironmentConfig()
	wasLite := env.Lite
	env.Lite = false
	t.Cleanup(func() { env.Lite = wasLite })
	selected, _ := resolv.ParseIPDNS("tls://192.0.2.53")
	for _, active := range []bool{false, true} {
		setting := configure.NewSetting()
		setting.TransparentType = configure.TransparentRedirect
		if active {
			setting.Transparent = configure.TransparentProxy
		}
		tmpl := &Template{Setting: setting, NodeDNS: selected}
		if err := tmpl.generateDnsModuleConfig([]serverInfo{{Info: &serverObj.V2Ray{Add: "node.example"}}}); err != nil {
			t.Fatal(err)
		}
		var cfg struct {
			Upstreams []struct{ ID, Addr, Protocol string }
		}
		if err := json.Unmarshal(tmpl.DnsModuleConfig, &cfg); err != nil {
			t.Fatal(err)
		}
		node := cfg.Upstreams[0]
		if active && (node.Addr != selected.Address() || node.Protocol != "tcp-tls") {
			t.Fatalf("selected endpoint lost: %+v", node)
		}
		if !active && node.Addr == selected.Address() {
			t.Fatal("node setting applied outside takeover")
		}
	}
	setting := configure.NewSetting()
	setting.Transparent = configure.TransparentProxy
	setting.TransparentType = configure.TransparentTun
	setting.TunAutoRoute = false
	if PlannedDNSHijack(setting) {
		t.Fatal("manual TUN assumes DNS takeover")
	}
	env.Lite = true
	setting.TransparentType = configure.TransparentSystemProxy
	if PlannedDNSHijack(setting) {
		t.Fatal("lite system proxy assumes DNS takeover")
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
