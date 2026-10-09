package dns

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestConfiguredDefaultUpstreamRoutesUnmatchedQuery(t *testing.T) {
	config := &DnsModuleConfig{
		DefaultUpstream: "first",
		Upstreams: []UpstreamConfig{
			{ID: "first", Addr: "1.1.1.1:53"},
			{ID: "last", Addr: "8.8.8.8:53"},
		},
		Rules: []RuleConfig{{ID: "specific", Domain: []string{"example.com"}, Upstream: "last"}},
	}
	rules := []*DnsRule{{ID: "specific", Domain: []string{"example.com"}, Upstream: "last"}}
	router, err := NewRouter(rules, config.Upstreams, selectDefaultUpstream(config))
	if err != nil {
		t.Fatal(err)
	}
	got := router.Route(&DnsQuery{Name: "unmatched.test", QType: TypeA})
	if got.UpstreamID != "first" || got.UpstreamAddr != "1.1.1.1:53" {
		t.Fatalf("route = %+v, want first upstream", got)
	}
}

func TestNodeLookupUsesModuleRoutingAndCache(t *testing.T) {
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var upstreamQueries atomic.Int32
	upstream := &dns.Server{PacketConn: pc, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		upstreamQueries.Add(1)
		reply := new(dns.Msg)
		reply.SetReply(request)
		name := request.Question[0].Name
		if name == "failed.invalid." {
			reply.Rcode = dns.RcodeServerFailure
		} else if request.Question[0].Qtype == dns.TypeA {
			reply.Answer = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300}, A: net.ParseIP("192.0.2.42")}}
		}
		_ = w.WriteMsg(reply)
	})}
	ready := make(chan struct{})
	upstream.NotifyStartedFunc = func() { close(ready) }
	go upstream.ActivateAndServe()
	<-ready
	defer upstream.Shutdown()
	// Reserve a port shared by the module's UDP and TCP transports.
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	cfg := DefaultDnsModuleConfig()
	cfg.Listener = DnsListenerConfig{ListenAddr: address, Timeout: 1, ReadinessToken: "current"}
	cfg.Upstreams = []UpstreamConfig{{ID: "upstream-node", Addr: pc.LocalAddr().String(), Protocol: "udp", ProxyTag: "direct"}}
	cfg.Rules = []RuleConfig{{ID: "rule-node", Upstream: "upstream-node", Action: "route", DomainSuffix: []string{"connected.invalid"}}}
	cfg.Cache.Prefetch = false
	module := NewDnsModule(cfg)
	if err := module.Start(); err != nil {
		t.Fatal(err)
	}
	defer module.Stop()
	lookup := func(host, token string) (*dns.Msg, error) {
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
		msg.SetEdns0(4096, false)
		msg.IsEdns0().Option = append(msg.IsEdns0().Option, &dns.EDNS0_LOCAL{Code: nodeQueryOption, Data: []byte(token)})
		response, _, err := (&dns.Client{Net: "tcp", Timeout: time.Second}).ExchangeContext(context.Background(), msg, address)
		return response, err
	}
	// A TCP-test node absent from the connected set must still hit rule-node.
	for range 2 {
		response, err := lookup("untested.invalid", "current")
		if err != nil || response.Rcode != dns.RcodeSuccess || len(response.Answer) != 1 || response.Answer[0].(*dns.A).A.String() != "192.0.2.42" {
			t.Fatalf("module lookup: %v %v", response, err)
		}
	}
	if upstreamQueries.Load() != 1 {
		t.Fatalf("module cache bypassed: %d upstream queries", upstreamQueries.Load())
	}
	before := upstreamQueries.Load()
	if response, err := lookup("untested.invalid", "previous"); err != nil || response.Rcode != dns.RcodeRefused {
		t.Fatalf("accepted stale module token: %v", err)
	}
	if upstreamQueries.Load() != before {
		t.Fatal("stale query reached upstream")
	}
	if response, err := lookup("failed.invalid", "current"); err != nil || response.Rcode != dns.RcodeServerFailure {
		t.Fatalf("upstream error hidden: %v", err)
	}
	// No connected domains must not turn rule-node into a catch-all.
	router, err := NewRouter([]*DnsRule{{ID: "rule-node", Upstream: "upstream-node", Action: "route"}}, cfg.Upstreams, "ordinary")
	if err != nil {
		t.Fatal(err)
	}
	if result := router.Route(&DnsQuery{Name: "website.invalid.", QType: QueryType(dns.TypeA)}); result.UpstreamID != "ordinary" {
		t.Fatalf("ordinary lookup routed to node upstream: %+v", result)
	}
}
