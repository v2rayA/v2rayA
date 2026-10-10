package v2ray

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestNodeDNSOffUsesSystemResolverWithoutModule(t *testing.T) {
	setting := configure.NewSetting()
	setting.DnsMode = configure.DnsModeOff
	p := &Process{template: &Template{Setting: setting}}
	blocked := errors.New("system DNS socket blocked by test")
	dialer := &net.Dialer{Control: func(string, string, syscall.RawConn) error { return blocked }}
	_, err := p.LookupNode(context.Background(), "node.example.invalid", dialer)
	if err == nil || !strings.Contains(err.Error(), blocked.Error()) {
		t.Fatalf("off mode did not use the system resolver: %v", err)
	}
}

func TestNodeDNSWaitsForCurrentModule(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	pc, err := net.ListenPacket("udp4", address)
	if err != nil {
		listener.Close()
		t.Fatal(err)
	}
	var token atomic.Pointer[string]
	old, current := "old", "current"
	token.Store(&old)
	var queries atomic.Int32
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		reply := new(dns.Msg)
		reply.SetReply(request)
		if request.Question[0].Name == "_v2raya-ready.invalid." {
			reply.Answer = []dns.RR{&dns.TXT{Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET}, Txt: []string{*token.Load()}}}
		} else {
			queries.Add(1)
			reply.Rcode = dns.RcodeServerFailure
		}
		_ = w.WriteMsg(reply)
	})
	for _, server := range []*dns.Server{{PacketConn: pc, Handler: handler}, {Listener: listener, Handler: handler}} {
		ready := make(chan struct{})
		server.NotifyStartedFunc = func() { close(ready) }
		go server.ActivateAndServe()
		<-ready
		t.Cleanup(func() { _ = server.Shutdown() })
	}
	env := conf.GetEnvironmentConfig()
	previous := env.CoreStartupTimeout
	env.CoreStartupTimeout = 1
	t.Cleanup(func() { env.CoreStartupTimeout = previous })
	lifetime, stop := context.WithCancel(context.Background())
	defer stop()
	endpoint, _ := resolv.ParseIPDNS("udp://192.0.2.53")
	p := &Process{ctx: lifetime, dnsToken: current, dnsAddress: address, template: &Template{Setting: configure.NewSetting(), NodeDNS: endpoint}}
	result := make(chan error, 1)
	go func() { _, err := p.LookupNode(context.Background(), "untested-node.invalid", nil); result <- err }()
	select {
	case err := <-result:
		t.Fatalf("accepted old listener: %v", err)
	case <-time.After(120 * time.Millisecond):
	}
	if queries.Load() != 0 {
		t.Fatal("node query sent before current module was ready")
	}
	token.Store(&current)
	if err := <-result; err == nil || !strings.Contains(err.Error(), "SERVFAIL") {
		t.Fatalf("upstream failure treated as readiness: %v", err)
	}
	if queries.Load() != 2 {
		t.Fatalf("wanted A and AAAA once, got %d", queries.Load())
	}
	if err := p.WaitDNSReady(context.Background()); err != nil {
		t.Fatalf("upstream failure invalidated readiness: %v", err)
	}
	token.Store(&old)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if err := p.WaitDNSReady(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("stale module wait: %v", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	if err := p.WaitDNSReady(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled wait: %v", err)
	}
	stop()
	if err := p.WaitDNSReady(context.Background()); err == nil || !strings.Contains(err.Error(), "exited or stopped") {
		t.Fatalf("process exit not handled: %v", err)
	}
	if queries.Load() != 2 {
		t.Fatal("waiting/cancellation sent more node queries")
	}
}

func TestNodeDNSReadinessTimeout(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := env.CoreStartupTimeout
	env.CoreStartupTimeout = 1
	t.Cleanup(func() { env.CoreStartupTimeout = previous })
	listener, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	p := &Process{ctx: context.Background(), dnsToken: "never-ready", dnsAddress: listener.LocalAddr().String()}
	if err := p.WaitDNSReady(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("startup timeout: %v", err)
	}
}

func TestLocalDNSAddressUsesConfiguredPort(t *testing.T) {
	for listen, expected := range map[string]string{"0.0.0.0:15353": "127.0.0.1:15353", "[::]:25353": "[::1]:25353", "127.2.0.17:35353": "127.2.0.17:35353"} {
		if address, err := localDNSAddress(listen); err != nil || address != expected {
			t.Errorf("%s -> %s, %v", listen, address, err)
		}
	}
}

func TestStopCancelsDNSStartupWithoutHoldingManagerLock(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	t.Cleanup(func() { *env = previous })
	env.Config = t.TempDir()
	env.V2rayAssetsDirectory = env.Config
	env.CoreStartupTimeout = 5
	env.Lite = true
	env.CoreHook = ""
	env.V2rayBin = filepath.Join(env.Config, "core")
	if err := os.WriteFile(env.V2rayBin, []byte("#!/bin/sh\nexec sleep 20\n"), 0700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	setting := configure.NewSetting()
	setting.Transparent = configure.TransparentClose
	raw, err := json.Marshal(map[string]interface{}{"listener": map[string]string{"listen_addr": address}})
	if err != nil {
		t.Fatal(err)
	}
	template := &Template{Setting: setting, API: &coreObj.APIObject{}, DnsModuleConfig: raw}
	var manager CoreProcessManager
	result := make(chan error, 1)
	go func() { result <- manager.Start(template) }()
	waitCrashCondition(t, manager.Running)
	stopped := make(chan struct{})
	go func() { manager.Stop(true); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop blocked behind DNS startup wait")
	}
	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "exited or stopped") {
			t.Fatalf("cancelled startup result: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("startup ignored process cancellation")
	}
}
