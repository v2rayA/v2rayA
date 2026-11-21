package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestFetchRoutingA(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		disposition string
		body        string
		wantError   string
	}{
		{"valid", "text/plain; charset=utf-8", "", "default: direct\ndomain(domain: example.com) -> proxy\n", ""},
		{"html", "text/html", "", "<html>default: direct</html>", "raw text"},
		{"download", "text/plain", "attachment; filename=rules.txt", "default: direct", "not a download"},
		{"invalid", "text/plain", "", "not RoutingA syntax", "invalid RoutingA"},
		{"comments", "text/plain", "", "# no rules\n", "no rules found"},
		{"binary", "application/octet-stream", "", "default: direct\x00", "UTF-8 rules"},
		{"too large", "text/plain", "", strings.Repeat("# comment\n", 120000), "exceeds 1 MiB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				if tt.disposition != "" {
					w.Header().Set("Content-Disposition", tt.disposition)
				}
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			got, err := FetchRoutingA(configure.RoutingASource{URL: server.URL, DirectUpdate: true})
			if tt.wantError == "" {
				if err != nil || got != tt.body {
					t.Fatalf("FetchRoutingA() = %q, %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("FetchRoutingA() error = %v, want %q", err, tt.wantError)
			}
		})
	}
}

func TestFetchRoutingARejectsNonHTTP(t *testing.T) {
	for _, source := range []string{"file:///etc/passwd", "javascript:alert(1)", "https://user:pass@example.com/rules.txt"} {
		if _, err := FetchRoutingA(configure.RoutingASource{URL: source}); err == nil {
			t.Fatalf("FetchRoutingA(%q) accepted invalid URL", source)
		}
	}
}

func TestUpdateRoutingAFromSourceKeepsLastGoodRules(t *testing.T) {
	previousRules := configure.GetRoutingA()
	previousSource := configure.GetRoutingASource()
	t.Cleanup(func() {
		_ = configure.SetRoutingA(&previousRules)
		_ = configure.SetRoutingASource(previousSource)
	})
	body := "default: direct\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, body)
	}))
	defer server.Close()
	source := configure.RoutingASource{URL: server.URL, DirectUpdate: true, IntervalHours: 24}
	if err := configure.SetRoutingASource(source); err != nil {
		t.Fatal(err)
	}
	if err := UpdateRoutingAFromSource(); err != nil || configure.GetRoutingA() != body {
		t.Fatalf("first update: %v, %q", err, configure.GetRoutingA())
	}
	body = "<html>not raw rules</html>"
	if err := UpdateRoutingAFromSource(); err == nil || configure.GetRoutingA() != "default: direct\n" {
		t.Fatalf("invalid update replaced last good rules: %v, %q", err, configure.GetRoutingA())
	}
}

func TestRoutingARejectsHTTPSDowngrade(t *testing.T) {
	client, err := routingAClient(true)
	if err != nil {
		t.Fatal(err)
	}
	start, _ := http.NewRequest(http.MethodGet, "https://rules.example/start", nil)
	target, _ := http.NewRequest(http.MethodGet, "http://rules.example/plain", nil)
	if err := client.CheckRedirect(target, []*http.Request{start}); err == nil {
		t.Fatal("HTTPS rules redirected to plaintext")
	}
}

func TestSubscriptionRoutingOfferImport(t *testing.T) {
	for _, tc := range []struct {
		name     string
		urls     []string
		interval string
		hours    int
		invalid  bool
	}{
		{"absent", nil, "", 0, false},
		{"valid", []string{"https://rules.example/routing"}, "3", 3, false},
		{"bad interval", []string{"https://rules.example/routing"}, "0", 24, false},
		{"duplicate", []string{"https://a.example/r", "https://b.example/r"}, "", 0, true},
		{"http", []string{"http://rules.example/r"}, "", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				for _, u := range tc.urls {
					w.Header().Add("X-V2RayA-RoutingA-URL", u)
				}
				w.Header().Set("X-V2RayA-RoutingA-Interval-Hours", tc.interval)
				fmt.Fprint(w, base64.StdEncoding.EncodeToString([]byte("socks5://127.0.0.1:1080#test")))
			}))
			defer srv.Close()
			nodes, meta, err := resolveSubscriptionWithContext(context.Background(), srv.URL, srv.Client())
			if err != nil || len(nodes) != 1 || calls != 1 {
				t.Fatalf("fetch: %v, nodes=%d, calls=%d", err, len(nodes), calls)
			}
			if tc.urls == nil {
				if meta.offer != nil {
					t.Fatal(meta.offer)
				}
				return
			}
			if meta.offer == nil || meta.offer.IntervalHours != tc.hours || meta.offer.Invalid != tc.invalid {
				t.Fatal(meta.offer)
			}
		})
	}
}

func TestRoutingOfferOwnership(t *testing.T) {
	sub := resetSubscription(t)
	rules, src, setting := configure.GetRoutingA(), configure.GetRoutingASource(), configure.GetSettingNotNil()
	t.Cleanup(func() { _ = configure.SetRoutingAConfig(rules, src, setting) })
	offer := &configure.RoutingOffer{URL: "https://rules.example/a", IntervalHours: 24}
	if err := storeSubscriptionUpdate(0, sub, []serverObj.ServerObj{sub.Servers[0].ServerObj}, subscriptionMetadata{offer: offer}, false); err != nil {
		t.Fatal(err)
	}
	stored := configure.GetSubscription(0)
	if stored.RoutingOffer == nil || *stored.RoutingOffer != *offer {
		t.Fatal("offer missing after database round trip")
	}
	next := configure.RoutingASource{URL: offer.URL, IntervalHours: 3}
	preview := RoutingPreview{Revision: RoutingARevision(rules, src, setting), SubscriptionID: stored.DatabaseID}
	if err := SaveRoutingA("default: direct\n", next, &preview); err != nil {
		t.Fatal(err)
	}
	active := configure.GetRoutingASource()
	if active.SubscriptionID != sub.DatabaseID || active.IntervalHours != 3 || configure.GetSettingNotNil().RulePortMode != configure.RoutingAMode {
		t.Fatal(active)
	}
	if err := SaveRoutingA("default: proxy\n", next, &preview); err == nil {
		t.Fatal("stale preview applied")
	}
	b := &configure.SubscriptionRaw{Address: "https://b.example/list", RoutingOffer: &configure.RoutingOffer{URL: offer.URL, IntervalHours: 2}}
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{b}); err != nil {
		t.Fatal(err)
	}
	b = configure.GetSubscription(1)
	if configure.GetRoutingASource() != active {
		t.Fatal("a second subscription changed the accepted source")
	}
	other := &configure.RoutingOffer{URL: "https://rules.example/b", IntervalHours: 2}
	if err := storeSubscriptionUpdate(1, b, nil, subscriptionMetadata{offer: other}, false); err != nil {
		t.Fatal(err)
	}
	if configure.GetRoutingASource() != active {
		t.Fatal("background metadata update replaced source")
	}
	fresh := RoutingPreview{Revision: RoutingARevision(configure.GetRoutingA(), active, configure.GetSettingNotNil()), SubscriptionID: b.DatabaseID}
	if err := SaveRoutingA("default: proxy", next, &fresh); err == nil {
		t.Fatal("accepted an obsolete offer URL")
	}
	if err := DeleteWhich([]*configure.Which{{NodeRef: configure.NodeRef{ID: 1, TYPE: configure.SubscriptionType}}}); err != nil {
		t.Fatal(err)
	}
	if configure.GetRoutingASource().URL != "" || configure.GetRoutingA() != "default: direct\n" {
		t.Fatal("deleted owner left scheduled URL or lost rules")
	}
	manualSource := configure.RoutingASource{URL: other.URL, IntervalHours: 2, SubscriptionID: b.DatabaseID}
	if err := configure.SetRoutingASource(manualSource); err != nil {
		t.Fatal(err)
	}
	if err := DeleteWhichWithRoutingSource([]*configure.Which{{NodeRef: configure.NodeRef{ID: 1, TYPE: configure.SubscriptionType}}}, b.DatabaseID); err != nil {
		t.Fatal(err)
	}
	if kept := configure.GetRoutingASource(); kept.SubscriptionID != 0 || kept.URL != other.URL {
		t.Fatal("manual source lost on removal")
	}

}
