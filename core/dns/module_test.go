package dns

import "testing"

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
