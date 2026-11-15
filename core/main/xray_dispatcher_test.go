package main

import (
	"context"
	"github.com/miekg/dns"
	moddns "github.com/v2rayA/v2raya-core/dns"
	xcore "github.com/xtls/xray-core/core"
	"net"
	"testing"

	"github.com/xtls/xray-core/common/buf"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/transport"
	"github.com/xtls/xray-core/transport/pipe"
)

type dnsTestDispatcher struct {
	routing.Dispatcher
	t *testing.T
}

func (d dnsTestDispatcher) Dispatch(ctx context.Context, dest xnet.Destination) (*transport.Link, error) {
	if dest.Network != xnet.Network_UDP || dest.Address.String() != "192.0.2.1" || dest.Port != 53 {
		d.t.Errorf("destination = %v", dest)
	}
	if inbound := session.InboundFromContext(ctx); inbound == nil || inbound.Tag != "chosen-group" {
		d.t.Errorf("inbound = %v", inbound)
	}
	if content := session.ContentFromContext(ctx); content == nil || !content.SkipDNSResolve {
		d.t.Error("DNS recursion protection missing")
	}
	reader, writer := pipe.New(pipe.WithSizeLimit(0))
	a, b := buf.New(), buf.New()
	a.Write([]byte("first"))
	b.Write([]byte("second"))
	if err := writer.WriteMultiBuffer(buf.MultiBuffer{a, b}); err != nil {
		d.t.Fatal(err)
	}
	_, input := pipe.New(pipe.WithSizeLimit(0))
	return &transport.Link{Reader: reader, Writer: input}, nil
}
func TestDNSDispatcherPreservesDatagrams(t *testing.T) {
	adapter := &xrayDispatcherAdapter{dispatcher: dnsTestDispatcher{t: t}}
	conn, err := adapter.Dispatch(context.Background(), "udp", "192.0.2.1:53", "chosen-group")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data := make([]byte, 65535)
	for _, want := range []string{"first", "second"} {
		n, err := conn.Read(data)
		if err != nil || string(data[:n]) != want {
			t.Fatalf("read = %q, %v; want %q", data[:n], err, want)
		}
	}
}

func TestDNSDispatcherUDPWithRunningCore(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &dns.Server{PacketConn: pc, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, q *dns.Msg) { r := new(dns.Msg); r.SetReply(q); w.WriteMsg(r) })}
	ready := make(chan struct{})
	server.NotifyStartedFunc = func() { close(ready) }
	go server.ActivateAndServe()
	<-ready
	defer server.Shutdown()
	core, err := xcore.StartInstance("json", []byte(`{"outbounds":[{"protocol":"blackhole","tag":"blocked"},{"protocol":"freedom","tag":"resolver"}],"routing":{"rules":[{"type":"field","inboundTag":["chosen-group"],"outboundTag":"resolver"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	defer core.Close()
	adapter, err := newXrayDispatcherAdapter(core)
	if err != nil {
		t.Fatal(err)
	}
	mgr := moddns.NewUpstreamManager(nil)
	mgr.SetDispatcher(adapter)
	resp, err := mgr.Exchange(&moddns.UpstreamInstance{Addr: pc.LocalAddr().String(), Protocol: "udp", ProxyTag: "chosen-group"}, &moddns.DnsQuery{Name: "example.com", QType: moddns.TypeA})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != dns.RcodeSuccess {
		t.Fatalf("rcode = %d", resp.Rcode)
	}
}
