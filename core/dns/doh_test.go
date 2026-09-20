package dns

import (
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/miekg/dns"
)

// a DoH server answering every A query with 192.0.2.1
func dohServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/dns-message" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		body, _ := io.ReadAll(r.Body)
		q := new(dns.Msg)
		if err := q.Unpack(body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a := new(dns.Msg)
		a.SetReply(q)
		a.Answer = append(a.Answer, &dns.A{Hdr: dns.RR_Header{Name: q.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}, A: net.ParseIP("192.0.2.1")})
		out, _ := a.Pack()
		w.Header().Set("Content-Type", "application/dns-message")
		w.Write(out)
	}))
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	old := dohRootCAs
	dohRootCAs = pool
	t.Cleanup(func() { dohRootCAs = old; srv.Close() })
	return srv
}

func TestExchangeDoHDirect(t *testing.T) {
	srv := dohServer(t)
	mgr := NewUpstreamManager([]UpstreamConfig{{ID: "doh", Addr: srv.URL + "/dns-query", Protocol: "https"}})
	up, ok := mgr.GetUpstream("doh")
	if !ok {
		t.Fatal("upstream not registered")
	}
	resp, err := mgr.Exchange(up, &DnsQuery{Name: "example.com", QType: TypeA})
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if len(resp.Answer) != 1 || resp.Answer[0].(*dns.A).A.String() != "192.0.2.1" {
		t.Fatalf("answer = %v", resp.Answer)
	}
	// the transport is kept: a second query reuses the client
	if _, err := mgr.Exchange(up, &DnsQuery{Name: "example.org", QType: TypeA}); err != nil {
		t.Fatalf("second exchange: %v", err)
	}
}

func TestDialTargetUsesBootstrap(t *testing.T) {
	mgr := NewUpstreamManager(nil)
	mgr.SetBootstrapIPs(map[string]string{"dns.google": "8.8.8.8"})
	if got := mgr.dialTarget("dns.google:443"); got != "8.8.8.8:443" {
		t.Fatalf("dialTarget = %q", got)
	}
	if got := mgr.dialTarget("1.1.1.1:443"); got != "1.1.1.1:443" {
		t.Fatalf("dialTarget = %q", got)
	}
}
