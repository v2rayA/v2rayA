package v2ray

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/iptables"
)

// The template tests read settings through the database; point it at a
// scratch directory before anything opens it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "v2raya-template-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("V2RAYA_CONFIG", dir)
	conf.GetEnvironmentConfig().Config = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func tunSetting() *configure.Setting {
	s := configure.NewSetting()
	s.Transparent = configure.TransparentFollowRule
	s.TransparentType = configure.TransparentTun
	s.TunAutoRoute = true
	s.TunExcludeProcesses = "curl\nwget"
	s.IpForward = false
	return s
}

// baseTemplate is what NewTemplate starts from, without resolving servers
// (which would run the core binary for its version).
func baseTemplate(t *testing.T, setting *configure.Setting) *Template {
	t.Helper()
	tmpl := new(Template)
	if err := json.Unmarshal([]byte(TemplateJson), tmpl); err != nil {
		t.Fatal(err)
	}
	tmpl.Setting = setting
	return tmpl
}

func findInbound(t *testing.T, tmpl *Template, protocol string) *coreObj.Inbound {
	t.Helper()
	var found *coreObj.Inbound
	for i := range tmpl.Inbounds {
		if tmpl.Inbounds[i].Protocol == protocol {
			if found != nil {
				t.Fatalf("more than one %s inbound", protocol)
			}
			found = &tmpl.Inbounds[i]
		}
	}
	return found
}

func TestTunModeGeneratesTheDeviceInboundAndDNSModule(t *testing.T) {
	setting := tunSetting()
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setInbound(setting); err != nil {
		t.Fatal(err)
	}
	if err := tmpl.setDNS(nil); err != nil {
		t.Fatal(err)
	}
	tmpl.setDualStack()

	ib := findInbound(t, tmpl, "tun-mips")
	if ib == nil {
		t.Fatal("no tun-mips inbound")
	}
	if ib.Tag != "transparent" || ib.Port != 0 || ib.Listen != "" {
		t.Fatalf("inbound %+v", ib)
	}
	for _, other := range tmpl.Inbounds {
		if other.Port == 52345 {
			t.Fatalf("the TinyTun socks loop-back inbound is still generated: %+v", other)
		}
	}
	s := ib.Settings
	if s.Name == "" || s.MTU != tunMTU || s.Address4 != tunAddress4 || s.DnsTarget != "127.0.0.1:52353" || s.DirectTag != "direct" {
		t.Fatalf("settings %+v", s)
	}
	if len(s.SelfPids) != 1 || s.SelfPids[0] != uint32(os.Getpid()) {
		t.Fatalf("selfPids %v", s.SelfPids)
	}
	if !contains(s.ExcludeProcesses, "curl") || !contains(s.ExcludeProcesses, "wget") || !contains(s.ExcludeProcesses, filepath.Base(os.Args[0])) {
		t.Fatalf("excludeProcesses %v lacks the user's list or v2rayA itself", s.ExcludeProcesses)
	}
	if !ib.Sniffing.Enabled {
		t.Fatal("sniffing not enabled on the tun inbound")
	}
	if len(tmpl.DnsModuleConfig) == 0 {
		t.Fatal("tun mode must generate the DNS module configuration")
	}
	var dnsCfg struct {
		Listener struct {
			ListenAddr string   `json:"listen_addr"`
			Extra      []string `json:"extra_listen_addrs"`
		} `json:"listener"`
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &dnsCfg); err != nil {
		t.Fatal(err)
	}
	if dnsCfg.Listener.ListenAddr != "127.0.0.1:52353" {
		t.Fatalf("DNS module listens on %q without port sharing", dnsCfg.Listener.ListenAddr)
	}
}

func TestDnsModuleListenAddressFollowsPortSharingAndUserValue(t *testing.T) {
	s := configure.NewSetting()
	s.IpForward = false
	if got := dnsModuleListenAddr(s); got != "127.0.0.1:52353" {
		t.Fatalf("default without port sharing: %q", got)
	}
	s.IpForward = true
	if got := dnsModuleListenAddr(s); got != "0.0.0.0:52353" {
		t.Fatalf("default with IP forwarding (LAN clients behind this host): %q", got)
	}
	s.IpForward = false
	s.PortSharing = true
	if got := dnsModuleListenAddr(s); got != "0.0.0.0:52353" {
		t.Fatalf("default with port sharing: %q", got)
	}
	s.DnsListenAddr = "0.0.0.0:5353"
	if got := dnsModuleListenAddr(s); got != "0.0.0.0:5353" {
		t.Fatalf("explicit value: %q", got)
	}
	if got := tunDnsTarget(s); got != "127.0.0.1:5353" {
		t.Fatalf("dns target for a wildcard listener: %q", got)
	}
	if got := dnsModulePort(s); got != "5353" {
		t.Fatalf("dns port: %q", got)
	}
	s.DnsListenAddr = "192.168.1.2:5353"
	if got := tunDnsTarget(s); got != "192.168.1.2:5353" {
		t.Fatalf("dns target for a specific listener: %q", got)
	}
}

func TestExtraDnsListenersFollowThePrimary(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux listeners")
	}
	s := tunSetting()
	loop := dnsModuleExtraListenAddrs(s)
	s.PortSharing = true
	wild := dnsModuleExtraListenAddrs(s)
	for _, a := range wild {
		if strings.HasPrefix(a, "[::1]") {
			t.Fatalf("wildcard primary listener must not get a ::1 twin, got %v", wild)
		}
	}
	if iptables.IsIPv6Supported() && !contains(loop, "[::1]:52353") {
		t.Fatalf("loopback primary listener lacks the ::1 twin: %v", loop)
	}
}

func TestDnsRedirectCommandsUseTheModulePort(t *testing.T) {
	s := tunSetting()
	s.DnsListenAddr = "0.0.0.0:5353"
	if err := configure.SetSetting(s); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.SetSetting(configure.NewSetting()) })
	del := dnsRedirectDeleteCommands()
	if !strings.Contains(del, "--to-port 5353") || !strings.Contains(del, "--to-port 52353") {
		t.Fatalf("delete commands must cover the configured and the default port:\n%s", del)
	}
	// A port used earlier in this process stays on the list after the
	// setting moved on, otherwise its rules would survive the change.
	rememberDnsRedirectPort("5353")
	s.DnsListenAddr = "0.0.0.0:5354"
	if err := configure.SetSetting(s); err != nil {
		t.Fatal(err)
	}
	del = dnsRedirectDeleteCommands()
	for _, port := range []string{"5353", "5354", "52353"} {
		if !strings.Contains(del, "--to-port "+port) {
			t.Fatalf("delete commands lack port %s after a change:\n%s", port, del)
		}
	}
}

// TestTunTemplateLoadsInTheCore feeds the generated document to the core
// binary named by V2RAYA_CORE_BIN — the seam test proper. Skipped when
// the binary is not given.
func TestTunTemplateLoadsInTheCore(t *testing.T) {
	bin := os.Getenv("V2RAYA_CORE_BIN")
	if bin == "" {
		t.Skip("V2RAYA_CORE_BIN not set")
	}
	setting := tunSetting()
	tmpl := baseTemplate(t, setting)
	if err := tmpl.setInbound(setting); err != nil {
		t.Fatal(err)
	}
	if err := tmpl.setDNS(nil); err != nil {
		t.Fatal(err)
	}
	tmpl.setDualStack()
	tmpl.Outbounds = append(tmpl.Outbounds, coreObj.OutboundObject{Protocol: "freedom", Tag: "direct"})
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, tmpl.ToConfigBytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(bin, "run", "-test", "-c", path).CombinedOutput()
	if err != nil {
		t.Fatalf("core rejected the generated configuration: %v\n%s", err, out)
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
