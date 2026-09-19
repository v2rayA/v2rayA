package dns

import (
	"testing"

	"github.com/miekg/dns"
)

func TestSanitizeForClientMatchesClientEDNS(t *testing.T) {
	upstream := func() *dns.Msg {
		m := new(dns.Msg)
		m.SetQuestion("example.org.", dns.TypeA)
		m.Response = true
		m.Answer = []dns.RR{
			&dns.A{Hdr: dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}, A: []byte{1, 2, 3, 4}},
			&dns.RRSIG{Hdr: dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeRRSIG, Class: dns.ClassINET, Ttl: 60}, TypeCovered: dns.TypeA},
		}
		m.SetEdns0(1232, true)
		m.IsEdns0().Option = append(m.IsEdns0().Option, &dns.EDNS0_LOCAL{Code: loopOptionCode, Data: []byte("tok")})
		return m
	}

	plain := new(dns.Msg)
	plain.SetQuestion("example.org.", dns.TypeA)
	m := upstream()
	sanitizeForClient(m, plain)
	if m.IsEdns0() != nil {
		t.Fatal("client without OPT got an OPT back")
	}
	if len(m.Answer) != 1 || m.Answer[0].Header().Rrtype != dns.TypeA {
		t.Fatalf("answers %v, want the A record only", m.Answer)
	}

	edns := new(dns.Msg)
	edns.SetQuestion("example.org.", dns.TypeA)
	edns.SetEdns0(4096, false)
	m = upstream()
	sanitizeForClient(m, edns)
	opt := m.IsEdns0()
	if opt == nil || opt.Do() || opt.UDPSize() != 4096 || len(opt.Option) != 0 {
		t.Fatalf("OPT %+v, want DO off, size 4096, no options", opt)
	}
	if len(m.Answer) != 1 {
		t.Fatalf("answers %v, want RRSIG dropped", m.Answer)
	}

	do := new(dns.Msg)
	do.SetQuestion("example.org.", dns.TypeA)
	do.SetEdns0(1232, true)
	m = upstream()
	sanitizeForClient(m, do)
	if opt := m.IsEdns0(); opt == nil || !opt.Do() {
		t.Fatal("client with DO lost DO")
	}
	if len(m.Answer) != 2 {
		t.Fatalf("answers %v, want RRSIG kept for a DO client", m.Answer)
	}
}
