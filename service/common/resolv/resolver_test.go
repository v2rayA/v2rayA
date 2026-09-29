package resolv

import (
	"context"
	"net"
	"reflect"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func serveDNS(t *testing.T, answer string, rcode int) (string, *atomic.Int32) {
	t.Helper()
	packet, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	queries := new(atomic.Int32)
	server := &dns.Server{PacketConn: packet, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		queries.Add(1)
		reply := new(dns.Msg).SetReply(request)
		reply.Rcode = rcode
		if answer != "" && request.Question[0].Qtype == dns.TypeA {
			reply.Answer = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 1}, A: net.ParseIP(answer)}}
		}
		_ = w.WriteMsg(reply)
	})}
	ready := make(chan struct{})
	server.NotifyStartedFunc = func() { close(ready) }
	go func() { _ = server.ActivateAndServe() }()
	<-ready
	t.Cleanup(func() { _ = server.Shutdown() })
	return packet.LocalAddr().String(), queries
}

func resolverAt(address string) *net.Resolver {
	return &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: time.Second}).DialContext(ctx, network, address)
	}}
}

func TestLookupHostRechecksUnusableSystemAnswers(t *testing.T) {
	for _, tc := range []struct {
		name, answer string
		rcode        int
		fallback     bool
	}{
		{"reachable", "192.0.2.1", dns.RcodeSuccess, false},
		{"loopback", "127.0.0.1", dns.RcodeSuccess, true},
		{"unspecified", "0.0.0.0", dns.RcodeSuccess, true},
		{"empty", "", dns.RcodeSuccess, true},
		{"nxdomain", "", dns.RcodeNameError, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			system, systemQueries := serveDNS(t, tc.answer, tc.rcode)
			fallback, fallbackQueries := serveDNS(t, "198.51.100.2", dns.RcodeSuccess)
			got, err := lookupHost("node.test", resolverAt(system), resolverAt(fallback))
			want := tc.answer
			if tc.fallback {
				want = "198.51.100.2"
			}
			if err != nil || !reflect.DeepEqual(got, []string{want}) {
				t.Fatalf("addresses=%v, err=%v, want %s", got, err, want)
			}
			if systemQueries.Load() == 0 || (fallbackQueries.Load() > 0) != tc.fallback {
				t.Fatalf("system queries=%d, fallback queries=%d", systemQueries.Load(), fallbackQueries.Load())
			}
		})
	}
}

func TestResolversUseDialerForSystemPreferredAndPublicDNS(t *testing.T) {
	address, _ := serveDNS(t, "192.0.2.1", dns.RcodeSuccess)
	oldPreferred, oldServers := PreferredServers, dnsServers
	t.Cleanup(func() { PreferredServers, dnsServers = oldPreferred, oldServers })
	dnsServers = []struct{ addr, network string }{{address, "udp"}}
	for _, mode := range []string{"system", "preferred", "public"} {
		t.Run(mode, func(t *testing.T) {
			PreferredServers = func() []string { return nil }
			if mode == "preferred" {
				PreferredServers = func() []string { return []string{address} }
			}
			var controlled atomic.Int32
			dialer := &net.Dialer{Timeout: time.Second, Control: func(_, got string, _ syscall.RawConn) error {
				if got != address {
					t.Errorf("DNS dialled %s, want %s", got, address)
				}
				controlled.Add(1)
				return nil
			}}
			fallback, system := newResolvers(dialer)
			resolver := fallback
			if mode == "system" {
				dial := system.Dial
				system.Dial = func(ctx context.Context, network, _ string) (net.Conn, error) { return dial(ctx, network, address) }
				resolver = system
			}
			got, err := resolver.LookupHost(context.Background(), "node.test")
			if err != nil || !reflect.DeepEqual(got, []string{"192.0.2.1"}) || controlled.Load() == 0 {
				t.Fatalf("addresses=%v, err=%v, controlled=%d", got, err, controlled.Load())
			}
		})
	}
}
