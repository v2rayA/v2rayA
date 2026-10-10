package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/server/service"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "v2raya-controller-test-*")
	if err != nil {
		panic(err)
	}
	conf.GetEnvironmentConfig().Config = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func runningCoreWithInvalidConnection(t *testing.T) {
	t.Helper()
	env := conf.GetEnvironmentConfig()
	previous := *env
	t.Cleanup(func() { *env = previous })
	env.Lite = true
	env.CoreHook = ""
	env.V2rayAssetsDirectory = t.TempDir()
	env.V2rayBin = filepath.Join(t.TempDir(), "core")
	if err := os.WriteFile(env.V2rayBin, []byte("#!/bin/sh\nexec sleep 60\n"), 0700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) {
			t.Skip("sandbox does not permit the fake core readiness listener")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	setting := configure.NewSetting()
	setting.Transparent = configure.TransparentClose
	// The fixture supplies the readiness port; config regeneration still uses the real database.
	tmpl := &v2ray.Template{Setting: setting, API: &coreObj.APIObject{}, ApiPort: listener.Addr().(*net.TCPAddr).Port}
	if err := v2ray.ProcessManager.Start(tmpl); err != nil {
		t.Fatal(err)
	}
	p := v2ray.ProcessManager.Process()
	t.Cleanup(func() {
		v2ray.ProcessManager.Stop(true)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := p.WaitUntilExit(ctx); err != nil {
			t.Error(err)
		}
	})
	if err := configure.AddConnect(configure.NodeRef{TYPE: configure.ServerType, ID: 999}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { configure.ClearConnects("proxy") })
}

func TestPutDnsRulesReportsReloadFailure(t *testing.T) {
	runningCoreWithInvalidConnection(t)
	previous := configure.GetDnsRulesNotNil()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/dnsRules", strings.NewReader(`[{"upstream":"1.1.1.1","domains":"example.com"}]`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutDnsRules(ctx)
	var response struct {
		Code common.Code `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != common.FAIL {
		t.Fatalf("reload failure answered %s", recorder.Body.String())
	}
	if got := configure.GetDnsRulesNotNil(); !reflect.DeepEqual(got, previous) {
		t.Fatalf("DNS rules were not restored: %+v", got)
	}
}

func TestPutRoutingARestoresAfterReloadFailure(t *testing.T) {
	original := configure.GetRoutingA()
	t.Cleanup(func() { _ = configure.SetRoutingA(&original) })
	previous := "default: proxy"
	if err := configure.SetRoutingA(&previous); err != nil {
		t.Fatal(err)
	}
	runningCoreWithInvalidConnection(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/routingA", strings.NewReader(`{"routingA":"default: direct"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutRoutingA(ctx)
	var response struct {
		Code common.Code `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != common.FAIL {
		t.Fatalf("reload failure answered %s", recorder.Body.String())
	}
	if got := configure.GetRoutingA(); got != previous {
		t.Fatalf("RoutingA = %q, want %q", got, previous)
	}
}

func TestSetPortsRestoresAfterReloadFailure(t *testing.T) {
	runningCoreWithInvalidConnection(t)
	previous := service.GetPorts()
	previous.Socks5 = 12345
	if err := configure.SetPorts(&previous); err != nil {
		t.Fatal(err)
	}
	next := previous
	next.Socks5 = 0
	if err := service.SetPorts(&next); err == nil {
		t.Fatal("invalid connected server did not reject reload")
	}
	if got := service.GetPorts(); !reflect.DeepEqual(got, previous) {
		t.Fatalf("ports were not restored: %+v, want %+v", got, previous)
	}
}

func TestPutOutboundRestoresAfterReloadFailure(t *testing.T) {
	runningCoreWithInvalidConnection(t)
	previous := configure.OutboundSetting{
		ProbeURL:      "https://previous.example/ping",
		ProbeInterval: "30s",
		Type:          configure.LeastPing,
	}
	if err := configure.SetOutboundSetting("proxy", previous); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", strings.NewReader(`{"outbound":"proxy","setting":{"probeURL":"https://next.example/ping","probeInterval":"45s","type":"leastping"}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	PutOutbound(ctx)

	var response struct {
		Code common.Code `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != common.FAIL {
		t.Fatalf("reload failure answered %s", recorder.Body.String())
	}
	if got := configure.GetOutboundSetting("proxy"); got != previous {
		t.Fatalf("outbound setting was not restored: %+v, want %+v", got, previous)
	}
}

func TestNodeDNSRevalidatesSourcesBeforeMutations(t *testing.T) {
	oldGFW, oldSubscription := conf.TickerUpdateGFWList, conf.TickerUpdateSubscription
	conf.TickerUpdateGFWList = time.NewTicker(time.Hour)
	conf.TickerUpdateSubscription = time.NewTicker(time.Hour)
	t.Cleanup(func() {
		conf.TickerUpdateGFWList.Stop()
		conf.TickerUpdateSubscription.Stop()
		conf.TickerUpdateGFWList, conf.TickerUpdateSubscription = oldGFW, oldSubscription
	})
	previousSetting, previousRules := configure.GetSettingNotNil(), configure.GetDnsRulesNotNil()
	t.Cleanup(func() { _ = configure.SetSetting(previousSetting); _ = configure.SetDnsRules(previousRules) })
	setting := configure.NewSetting()
	setting.NodeDns = "tls://192.0.2.53:853"
	if err := configure.SetSetting(setting); err != nil {
		t.Fatal(err)
	}
	if err := configure.SetDnsRules([]configure.DnsRule{{Server: "tls://192.0.2.53", Outbound: "custom"}}); err != nil {
		t.Fatal(err)
	}
	options := v2ray.CollectNodeDNSOptions(setting, configure.GetDnsRulesNotNil())
	if _, err := options.Select(setting.NodeDns); err != nil {
		t.Fatal(err)
	}
	request := func(body, path string, handler gin.HandlerFunc) string {
		t.Helper()
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler(ctx)
		return recorder.Body.String()
	}
	// A DNS rule edit may not remove the currently selected source.
	response := request(`[{"server":"8.8.8.8","outbound":"direct"}]`, "/dnsRules", PutDnsRules)
	if !strings.Contains(response, "NODE_DNS_INVALID") {
		t.Fatalf("stale rule change accepted: %s", response)
	}
	if configure.GetDnsRulesNotNil()[0].Server != "tls://192.0.2.53" {
		t.Fatal("rejected rules persisted")
	}
	// The candidate list can become obsolete before a settings submission.
	if err := configure.SetDnsRules([]configure.DnsRule{{Server: "8.8.8.8", Outbound: "direct"}}); err != nil {
		t.Fatal(err)
	}
	response = request(`{"nodeDns":"tls://192.0.2.53:853"}`, "/setting", PutSetting)
	if !strings.Contains(response, "NODE_DNS_INVALID") {
		t.Fatalf("stale settings accepted: %s", response)
	}
	response = request(`{"nodeDns":"auto"}`, "/setting", PutSetting)
	if strings.Contains(response, "NODE_DNS_INVALID") || configure.GetSettingNotNil().NodeDns != "auto" {
		t.Fatalf("auto not saved: %s", response)
	}
	// Old clients preserve an explicit choice, and input is canonicalized.
	response = request(`{"nodeDns":"tcp://8.8.8.8"}`, "/setting", PutSetting)
	if !strings.Contains(response, "NODE_DNS_INVALID") {
		t.Fatal("wrong protocol accepted")
	}
	response = request(`{"nodeDns":"8.8.8.8"}`, "/setting", PutSetting)
	if configure.GetSettingNotNil().NodeDns != "udp://8.8.8.8:53" {
		t.Fatalf("not normalized: %s", response)
	}
	response = request(`{"logLevel":"debug"}`, "/setting", PutSetting)
	if configure.GetSettingNotNil().NodeDns != "udp://8.8.8.8:53" {
		t.Fatalf("old client overwrote selection: %s", response)
	}
	// Proposed listener settings are checked before storing anything.
	response = request(`{"dnsListenAddr":"8.8.8.8:53"}`, "/setting", PutSetting)
	if !strings.Contains(response, "NODE_DNS_INVALID") || configure.GetSettingNotNil().DnsListenAddr == "8.8.8.8:53" {
		t.Fatalf("self-referencing selection accepted: %s", response)
	}
	// Another source for the same endpoint permits removing the first rule.
	response = request(`[{"server":"8.8.8.8","outbound":"proxy"}]`, "/dnsRules", PutDnsRules)
	if configure.GetDnsRulesNotNil()[0].Outbound != "proxy" {
		t.Fatalf("surviving source rejected: %s", response)
	}
}

func TestDnsDialogCombinedSave(t *testing.T) {
	previousSetting, previousRules := configure.GetSettingNotNil(), configure.GetDnsRulesNotNil()
	t.Cleanup(func() { _ = configure.SetDnsSettings(previousRules, previousSetting) })
	setting := configure.NewSetting()
	setting.DnsMode = configure.DnsModeService
	setting.NodeDns = "tls://192.0.2.53:853"
	setting.LogLevel = "debug"
	rules := []configure.DnsRule{{Server: "tls://192.0.2.53", Outbound: "direct"}}
	if err := configure.SetDnsSettings(rules, setting); err != nil {
		t.Fatal(err)
	}
	request := func(body string, handler gin.HandlerFunc) string {
		t.Helper()
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPut, "/dnsRules", strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler(ctx)
		return recorder.Body.String()
	}
	setting = configure.GetSettingNotNil()
	response := request(`{"rules":[{"server":"tcp://192.0.2.54","outbound":"direct"}]}`, PostNodeDNSOptions)
	if !strings.Contains(response, "tcp://192.0.2.54:53") || !reflect.DeepEqual(configure.GetDnsRulesNotNil(), rules) || !reflect.DeepEqual(configure.GetSettingNotNil(), setting) {
		t.Fatalf("preview did not preserve configuration: %s", response)
	}
	response = request(`{"rules":[{"server":"tcp://192.0.2.54","outbound":"direct"}],"dnsMode":"hijack","nodeDns":"tls://192.0.2.53:853"}`, PutDnsRules)
	if !strings.Contains(response, "NODE_DNS_INVALID") || !reflect.DeepEqual(configure.GetDnsRulesNotNil(), rules) || !reflect.DeepEqual(configure.GetSettingNotNil(), setting) {
		t.Fatalf("invalid replacement changed configuration: %s", response)
	}
	response = request(`{"rules":[{"server":"tcp://192.0.2.54","outbound":"direct"}],"dnsMode":"hijack","nodeDns":"tcp://192.0.2.54","logLevel":"error"}`, PutDnsRules)
	got := configure.GetSettingNotNil()
	if got.NodeDns != "tcp://192.0.2.54:53" || got.DnsMode != configure.DnsModeHijack || got.DnsHijack != configure.Yes || got.LogLevel != "debug" || configure.GetDnsRulesNotNil()[0].Server != "tcp://192.0.2.54" {
		t.Fatalf("combined replacement failed: %s, %+v", response, got)
	}
	// Turning the module off must remain possible even if the saved source
	// disappeared. Hidden edits must not replace the retained configuration.
	if err := configure.SetDnsRules(rules); err != nil {
		t.Fatal(err)
	}
	response = request(`{"dnsMode":"off","nodeDns":"auto","rules":[{"server":"bad://address"}]}`, PutDnsRules)
	got = configure.GetSettingNotNil()
	if got.DnsMode != configure.DnsModeOff || got.DnsHijack != configure.No || got.NodeDns != "tcp://192.0.2.54:53" || !reflect.DeepEqual(configure.GetDnsRulesNotNil(), rules) {
		t.Fatalf("off changed hidden configuration: %s, %+v", response, got)
	}
	response = request(`{"dnsMode":"servic"}`, PutDnsRules)
	if configure.GetSettingNotNil().DnsMode != configure.DnsModeOff || !strings.Contains(response, "unknown dnsMode") {
		t.Fatalf("unknown mode accepted: %s", response)
	}
}

func TestDnsDialogRestoresRulesAndSettingAfterReloadFailure(t *testing.T) {
	previousSetting, previousRules := configure.GetSettingNotNil(), configure.GetDnsRulesNotNil()
	t.Cleanup(func() { _ = configure.SetDnsSettings(previousRules, previousSetting) })
	runningCoreWithInvalidConnection(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/dnsRules", strings.NewReader(`{"rules":[{"server":"192.0.2.54","outbound":"direct"}],"dnsMode":"service","nodeDns":"192.0.2.54"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutDnsRules(ctx)
	if !reflect.DeepEqual(configure.GetDnsRulesNotNil(), previousRules) || !reflect.DeepEqual(configure.GetSettingNotNil(), previousSetting) || !strings.Contains(recorder.Body.String(), `"code":"FAIL"`) {
		t.Fatalf("combined reload failure did not restore configuration: %s", recorder.Body.String())
	}
}

func TestDnsDialogStorageFailureKeepsBothValues(t *testing.T) {
	previousSetting, previousRules := configure.GetSettingNotNil(), configure.GetDnsRulesNotNil()
	// Reject the second write, after the transaction has written new rules.
	_, err := db.GetDB().Exec(`CREATE TRIGGER reject_dns_setting BEFORE INSERT ON system_config
		WHEN NEW.key = 'system:setting' BEGIN SELECT RAISE(ABORT, 'settings write rejected'); END`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.GetDB().Exec("DROP TRIGGER reject_dns_setting") })
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/dnsRules", strings.NewReader(`{"rules":[{"server":"192.0.2.54","outbound":"direct"}],"dnsMode":"service","nodeDns":"192.0.2.54"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutDnsRules(ctx)
	if !reflect.DeepEqual(configure.GetDnsRulesNotNil(), previousRules) || !reflect.DeepEqual(configure.GetSettingNotNil(), previousSetting) || !strings.Contains(recorder.Body.String(), "settings write rejected") {
		t.Fatalf("failed transaction changed configuration: %s", recorder.Body.String())
	}
}
