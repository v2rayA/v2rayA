package dns

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

type transportDispatcher struct {
	address string
	calls   []string
	reject  bool
}

func (d *transportDispatcher) Dispatch(ctx context.Context, network, _ string, tag string) (net.Conn, error) {
	d.calls = append(d.calls, network+":"+tag)
	if d.reject {
		return nil, fmt.Errorf("fixture rejects proxy")
	}
	return (&net.Dialer{Timeout: time.Second}).DialContext(ctx, network, d.address)
}

func TestUDPDispatcherDoesNotRequireTCPDNS(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &mdns.Server{PacketConn: pc, Handler: mdns.HandlerFunc(func(w mdns.ResponseWriter, q *mdns.Msg) {
		r := new(mdns.Msg)
		r.SetReply(q)
		r.Answer = []mdns.RR{&mdns.A{Hdr: mdns.RR_Header{Name: q.Question[0].Name, Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: 60}, A: net.ParseIP("198.51.100.123")}}
		_ = w.WriteMsg(r)
	})}
	go srv.ActivateAndServe()
	t.Cleanup(func() { _ = srv.Shutdown() })
	d := &transportDispatcher{address: pc.LocalAddr().String()}
	m := NewUpstreamManager([]UpstreamConfig{{ID: "test", Addr: d.address, Protocol: "udp", ProxyTag: "proxy"}})
	m.SetDispatcher(d)
	u, _ := m.GetUpstream("test")
	r, err := m.Exchange(u, &DnsQuery{Name: "stability.test", QType: TypeA})
	if err != nil || len(r.Answer) != 1 {
		t.Fatalf("UDP-only DNS failed: %v", err)
	}
	if len(d.calls) != 1 || d.calls[0] != "udp:proxy" {
		t.Fatalf("unexpected transport/path: %v", d.calls)
	}
}

func TestDispatcherFailureNeverFallsBackToDirect(t *testing.T) {
	d := &transportDispatcher{reject: true}
	m := NewUpstreamManager([]UpstreamConfig{{ID: "test", Addr: "192.0.2.1:53", Protocol: "udp", ProxyTag: "restricted"}})
	m.SetDispatcher(d)
	u, _ := m.GetUpstream("test")
	if _, err := m.Exchange(u, &DnsQuery{Name: "stability.test", QType: TypeA}); err == nil {
		t.Fatal("failed proxy was accepted")
	}
	for _, call := range d.calls {
		if call != "udp:restricted" && call != "tcp:restricted" {
			t.Fatalf("escaped configured proxy: %s", call)
		}
	}
}
