package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
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
