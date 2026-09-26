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

func TestPutOutboundFromCachedClientPreservesAutomaticMembership(t *testing.T) {
	previous := configure.DefaultOutboundSetting()
	previous.AutoAdd = true
	previous.ProbeInterval = "300s"
	if err := configure.SetOutboundSetting("proxy", previous); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.SetOutboundSetting("proxy", configure.DefaultOutboundSetting()) })
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", strings.NewReader(`{"outbound":"proxy","setting":{"probeURL":"https://next.example/ping","probeInterval":"45s","type":"leastping"}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	PutOutbound(ctx)

	if got := configure.GetOutboundSetting("proxy"); !got.AutoAdd || got.ProbeInterval != "45s" {
		t.Fatalf("cached client changed automatic membership: %+v", got)
	}
}

func TestPutOutboundDefaultsNewAutomaticGroupToFiveMinutes(t *testing.T) {
	previous := configure.DefaultOutboundSetting()
	if err := configure.SetOutboundSetting("proxy", previous); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.SetOutboundSetting("proxy", configure.DefaultOutboundSetting()) })
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", strings.NewReader(`{"outbound":"proxy","setting":{"autoAdd":true,"probeURL":"https://next.example/ping","probeInterval":"60s","type":"leastping"}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	PutOutbound(ctx)

	if got := configure.GetOutboundSetting("proxy"); !got.AutoAdd || got.ProbeInterval != "300s" {
		t.Fatalf("new automatic group setting = %+v; want autoAdd with 300s", got)
	}
}

func TestPutOutboundInitializesAndHidesKeepCurrentState(t *testing.T) {
	previousSetting := configure.GetOutboundSetting("proxy")
	previousMembers := configure.GetConnectedServersByOutbound("proxy")
	serverIndex := configure.GetLenServers()
	node, err := service.ResolveURL("http-proxy://127.0.0.1:19090#sticky-current")
	if err != nil {
		t.Fatal(err)
	}
	if err := configure.AppendServers([]*configure.ServerRaw{{ServerObj: node}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = configure.ClearConnects("proxy")
		_ = configure.RemoveServers([]int{serverIndex})
		if previousMembers != nil && previousMembers.Len() > 0 {
			_ = configure.OverwriteConnects(previousMembers)
		}
		_ = configure.SetOutboundSetting("proxy", previousSetting)
	})
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	if err := configure.AddConnect(configure.NodeRef{TYPE: configure.ServerType, ID: serverIndex + 1, Outbound: "proxy"}); err != nil {
		t.Fatal(err)
	}
	if err := configure.SetOutboundSetting("proxy", configure.DefaultOutboundSetting()); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", strings.NewReader(`{"outbound":"proxy","setting":{"probeURL":"https://www.gstatic.com/generate_204","probeInterval":"300s","type":"keepcurrent"}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutOutbound(ctx)
	if code, _ := codeOf(t, recorder); code != common.SUCCESS {
		t.Fatalf("keep-current update failed: %s", recorder.Body.String())
	}
	want := configure.NodeFingerprint(node.ExportToURL())
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != want {
		t.Fatalf("initial sticky current = %q; want %q", got, want)
	}

	response := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/outbound?outbound=proxy", nil)
	GetOutbound(ctx)
	var body struct {
		Data struct {
			Setting configure.OutboundSetting `json:"setting"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Setting.StickyCurrent != "" {
		t.Fatalf("internal sticky state leaked through API: %q", body.Data.Setting.StickyCurrent)
	}

	recorder = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", strings.NewReader(`{"outbound":"proxy","setting":{"probeURL":"https://www.gstatic.com/generate_204","probeInterval":"300s","type":"random"}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutOutbound(ctx)
	if code, _ := codeOf(t, recorder); code != common.SUCCESS {
		t.Fatalf("random strategy update failed: %s", recorder.Body.String())
	}
	if got := configure.GetOutboundSetting("proxy").StickyCurrent; got != "" {
		t.Fatalf("leaving keep-current retained internal state %q", got)
	}
}

func TestPutOutboundConnectionsRejectsAutomaticGroupEdits(t *testing.T) {
	setting := configure.DefaultOutboundSetting()
	setting.AutoAdd = true
	if err := configure.SetOutboundSetting("proxy", setting); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.SetOutboundSetting("proxy", configure.DefaultOutboundSetting()) })
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/outboundConnections", strings.NewReader(`{"outbound":"proxy","touches":[]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	PutOutboundConnections(ctx)

	var response struct {
		Code common.Code `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != common.FAIL {
		t.Fatalf("automatic group accepted manual membership edit: %s", recorder.Body.String())
	}
}
