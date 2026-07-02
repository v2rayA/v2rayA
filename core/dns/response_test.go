package dns

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// ============================================================================
// Helper functions
// ============================================================================

func newTestQuery(name string, qtype QueryType) *DnsQuery {
	return &DnsQuery{
		Name:   name,
		QType:  qtype,
		QClass: dns.ClassINET,
		ClientIP: net.ParseIP("192.168.1.100"),
	}
}

func newTestResponse(query *DnsQuery, rcode int, answers []dns.RR) *DnsResponse {
	m := new(dns.Msg)
	m.SetReply(&dns.Msg{
		MsgHdr: dns.MsgHdr{Id: dns.Id()},
		Question: []dns.Question{
			{Name: dns.Fqdn(query.Name), Qtype: uint16(query.QType), Qclass: query.QClass},
		},
	})
	m.Rcode = rcode
	m.RecursionAvailable = true
	m.Answer = answers

	return &DnsResponse{
		Query:    *query,
		RawMsg:   m,
		Rcode:    rcode,
		Answer:   answers,
		Upstream: "8.8.8.8:53",
		ProxyTag: "proxy",
		RTT:      10 * time.Millisecond,
	}
}

// ============================================================================
// Test: Record Type Build and Parse
// ============================================================================

func TestBuildResponse_A(t *testing.T) {
	query := newTestQuery("example.com", TypeA)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("example.com"),
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP("1.2.3.4").To4(),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("expected RcodeSuccess, got %d", msg.Rcode)
	}
	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	a, ok := msg.Answer[0].(*dns.A)
	if !ok {
		t.Fatalf("expected *dns.A, got %T", msg.Answer[0])
	}
	if a.A.String() != "1.2.3.4" {
		t.Fatalf("expected IP 1.2.3.4, got %s", a.A.String())
	}
	if !msg.Compress {
		t.Fatal("expected Compress to be true")
	}
}

func TestBuildResponse_AAAA(t *testing.T) {
	query := newTestQuery("ipv6.example.com", TypeAAAA)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("ipv6.example.com"),
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			AAAA: net.ParseIP("2001:db8::1"),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("expected RcodeSuccess, got %d", msg.Rcode)
	}
	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	aaaa, ok := msg.Answer[0].(*dns.AAAA)
	if !ok {
		t.Fatalf("expected *dns.AAAA, got %T", msg.Answer[0])
	}
	if aaaa.AAAA.String() != "2001:db8::1" {
		t.Fatalf("expected IP 2001:db8::1, got %s", aaaa.AAAA.String())
	}
}

func TestBuildResponse_CNAME(t *testing.T) {
	query := newTestQuery("www.example.com", TypeCNAME)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("www.example.com"),
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Target: dns.Fqdn("example.com"),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	cname, ok := msg.Answer[0].(*dns.CNAME)
	if !ok {
		t.Fatalf("expected *dns.CNAME, got %T", msg.Answer[0])
	}
	if cname.Target != dns.Fqdn("example.com") {
		t.Fatalf("expected target example.com., got %s", cname.Target)
	}
}

func TestBuildResponse_TXT(t *testing.T) {
	query := newTestQuery("example.com", TypeTXT)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("example.com"),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Txt: []string{"v=spf1 include:_spf.example.com ~all"},
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	txt, ok := msg.Answer[0].(*dns.TXT)
	if !ok {
		t.Fatalf("expected *dns.TXT, got %T", msg.Answer[0])
	}
	if len(txt.Txt) != 1 || txt.Txt[0] != "v=spf1 include:_spf.example.com ~all" {
		t.Fatalf("unexpected TXT record content: %v", txt.Txt)
	}
}

func TestBuildResponse_MX(t *testing.T) {
	query := newTestQuery("example.com", TypeMX)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.MX{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("example.com"),
				Rrtype: dns.TypeMX,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Preference: 10,
			Mx:         dns.Fqdn("mail.example.com"),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	mx, ok := msg.Answer[0].(*dns.MX)
	if !ok {
		t.Fatalf("expected *dns.MX, got %T", msg.Answer[0])
	}
	if mx.Mx != dns.Fqdn("mail.example.com") || mx.Preference != 10 {
		t.Fatalf("unexpected MX record: pref=%d, target=%s", mx.Preference, mx.Mx)
	}
}

func TestBuildResponse_SRV(t *testing.T) {
	query := newTestQuery("_sip._tcp.example.com", TypeSRV)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.SRV{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("_sip._tcp.example.com"),
				Rrtype: dns.TypeSRV,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Priority: 10,
			Weight:   5,
			Port:     5060,
			Target:   dns.Fqdn("sip.example.com"),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	srv, ok := msg.Answer[0].(*dns.SRV)
	if !ok {
		t.Fatalf("expected *dns.SRV, got %T", msg.Answer[0])
	}
	if srv.Priority != 10 || srv.Weight != 5 || srv.Port != 5060 || srv.Target != dns.Fqdn("sip.example.com") {
		t.Fatalf("unexpected SRV record: prio=%d weight=%d port=%d target=%s",
			srv.Priority, srv.Weight, srv.Port, srv.Target)
	}
}

func TestBuildResponse_NS(t *testing.T) {
	query := newTestQuery("example.com", TypeNS)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.NS{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("example.com"),
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Ns: dns.Fqdn("ns1.example.com"),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	ns, ok := msg.Answer[0].(*dns.NS)
	if !ok {
		t.Fatalf("expected *dns.NS, got %T", msg.Answer[0])
	}
	if ns.Ns != dns.Fqdn("ns1.example.com") {
		t.Fatalf("expected ns1.example.com., got %s", ns.Ns)
	}
}

func TestBuildResponse_PTR(t *testing.T) {
	query := newTestQuery("4.3.2.1.in-addr.arpa", TypePTR)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.PTR{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("4.3.2.1.in-addr.arpa"),
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Ptr: dns.Fqdn("host.example.com"),
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	ptr, ok := msg.Answer[0].(*dns.PTR)
	if !ok {
		t.Fatalf("expected *dns.PTR, got %T", msg.Answer[0])
	}
	if ptr.Ptr != dns.Fqdn("host.example.com") {
		t.Fatalf("expected host.example.com., got %s", ptr.Ptr)
	}
}

func TestBuildResponse_SOA(t *testing.T) {
	query := newTestQuery("example.com", TypeSOA)
	builder := NewResponseBuilder()

	answer := []dns.RR{
		&dns.SOA{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn("example.com"),
				Rrtype: dns.TypeSOA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Ns:      dns.Fqdn("ns1.example.com"),
			Mbox:    dns.Fqdn("admin.example.com"),
			Serial:  2024010100,
			Refresh: 3600,
			Retry:   1800,
			Expire:  86400,
			Minttl:  300,
		},
	}
	resp := newTestResponse(query, dns.RcodeSuccess, answer)
	msg := builder.BuildResponse(query, resp)

	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
	soa, ok := msg.Answer[0].(*dns.SOA)
	if !ok {
		t.Fatalf("expected *dns.SOA, got %T", msg.Answer[0])
	}
	if soa.Ns != dns.Fqdn("ns1.example.com") || soa.Serial != 2024010100 {
		t.Fatalf("unexpected SOA record: ns=%s serial=%d", soa.Ns, soa.Serial)
	}
}

// ============================================================================
// Test: Error Response Building
// ============================================================================

func TestBuildRefused(t *testing.T) {
	query := newTestQuery("example.com", TypeA)
	builder := NewResponseBuilder()

	msg := builder.BuildRefused(query)

	if msg.Rcode != dns.RcodeRefused {
		t.Fatalf("expected RcodeRefused (%d), got %d", dns.RcodeRefused, msg.Rcode)
	}
	if !msg.Response {
		t.Fatal("expected Response bit to be set")
	}
	if len(msg.Question) != 1 {
		t.Fatalf("expected 1 question, got %d", len(msg.Question))
	}
	if msg.Question[0].Name != dns.Fqdn("example.com") {
		t.Fatalf("expected question name %q, got %q", dns.Fqdn("example.com"), msg.Question[0].Name)
	}
}

func TestBuildNXDOMAIN(t *testing.T) {
	query := newTestQuery("nonexistent.example.com", TypeA)
	builder := NewResponseBuilder()

	msg := builder.BuildNXDOMAIN(query)

	if msg.Rcode != dns.RcodeNameError {
		t.Fatalf("expected RcodeNameError (%d), got %d", dns.RcodeNameError, msg.Rcode)
	}
	if !msg.Response {
		t.Fatal("expected Response bit to be set")
	}
}

func TestBuildServerFailure(t *testing.T) {
	query := newTestQuery("example.com", TypeA)
	builder := NewResponseBuilder()

	msg := builder.BuildServerFailure(query)

	if msg.Rcode != dns.RcodeServerFailure {
		t.Fatalf("expected RcodeServerFailure (%d), got %d", dns.RcodeServerFailure, msg.Rcode)
	}
	if !msg.Response {
		t.Fatal("expected Response bit to be set")
	}
}

// ============================================================================
// Test: Response Validation
// ============================================================================

func TestValidateResponse_Valid(t *testing.T) {
	m := new(dns.Msg)
	m.Response = true
	m.Rcode = dns.RcodeSuccess
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}

	err := ValidateResponse(m)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateResponse_NilMessage(t *testing.T) {
	err := ValidateResponse(nil)
	if err == nil {
		t.Fatal("expected error for nil message")
	}
}

func TestValidateResponse_NotResponse(t *testing.T) {
	m := new(dns.Msg)
	m.Response = false
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}

	err := ValidateResponse(m)
	if err == nil {
		t.Fatal("expected error for non-response message")
	}
}

func TestValidateResponse_NoQuestion(t *testing.T) {
	m := new(dns.Msg)
	m.Response = true
	m.Rcode = dns.RcodeSuccess

	err := ValidateResponse(m)
	if err == nil {
		t.Fatal("expected error for missing question section")
	}
}

func TestValidateResponse_NilAnswer(t *testing.T) {
	m := new(dns.Msg)
	m.Response = true
	m.Rcode = dns.RcodeSuccess
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	m.Answer = []dns.RR{nil}

	err := ValidateResponse(m)
	if err == nil {
		t.Fatal("expected error for nil answer record")
	}
}

func TestValidateQuestionMatch(t *testing.T) {
	m := new(dns.Msg)
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}

	// Matching query.
	err := ValidateQuestionMatch(m, "example.com", TypeA)
	if err != nil {
		t.Fatalf("expected no error for matching query, got: %v", err)
	}

	// Mismatched name.
	err = ValidateQuestionMatch(m, "other.com", TypeA)
	if err == nil {
		t.Fatal("expected error for mismatched name")
	}

	// Mismatched type.
	err = ValidateQuestionMatch(m, "example.com", TypeAAAA)
	if err == nil {
		t.Fatal("expected error for mismatched type")
	}
}

func TestValidateAnswerCount(t *testing.T) {
	m := new(dns.Msg)
	m.Rcode = dns.RcodeSuccess
	m.Answer = nil

	err := ValidateAnswerCount(m)
	if err == nil {
		t.Fatal("expected error for success with zero answers")
	}

	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}
	err = ValidateAnswerCount(m)
	if err != nil {
		t.Fatalf("expected no error for success with answers, got: %v", err)
	}
}

// ============================================================================
// Test: Truncation
// ============================================================================

func TestTruncation(t *testing.T) {
	builder := NewResponseBuilder()

	m := new(dns.Msg)
	m.Truncated = false
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}

	// Initially not truncated.
	if builder.IsTruncated(m) {
		t.Fatal("expected not truncated initially")
	}

	// Set truncated.
	builder.SetTruncated(m)
	if !builder.IsTruncated(m) {
		t.Fatal("expected truncated after SetTruncated")
	}

	// Test handleTruncation with small max size.
	builder.handleTruncation(m, 0)
	if !builder.IsTruncated(m) {
		t.Fatal("expected truncated after handleTruncation")
	}
	if m.Answer != nil {
		t.Fatal("expected answer to be cleared after truncation")
	}
}

// ============================================================================
// Test: EDNS0
// ============================================================================

func TestEDNS0(t *testing.T) {
	builder := NewResponseBuilder()
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)

	// Add EDNS0.
	builder.AddEDNS0(m, 4096)
	opt := m.IsEdns0()
	if opt == nil {
		t.Fatal("expected EDNS0 OPT record after AddEDNS0")
	}
	if opt.UDPSize() != 4096 {
		t.Fatalf("expected UDP size 4096, got %d", opt.UDPSize())
	}

	// Update EDNS0 UDP size.
	builder.AddEDNS0(m, 1232)
	if opt.UDPSize() != 1232 {
		t.Fatalf("expected UDP size 1232 after update, got %d", opt.UDPSize())
	}
}

func TestECSSubnet(t *testing.T) {
	builder := NewResponseBuilder()
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)

	// Add ECS for IPv4 client.
	builder.AddECSSubnet(m, net.ParseIP("192.168.1.100"))
	opt := m.IsEdns0()
	if opt == nil {
		t.Fatal("expected EDNS0 OPT record")
	}

	foundECS := false
	for _, o := range opt.Option {
		if subnet, ok := o.(*dns.EDNS0_SUBNET); ok {
			foundECS = true
			if subnet.Family != 1 {
				t.Fatalf("expected IPv4 family (1), got %d", subnet.Family)
			}
			if subnet.SourceNetmask != 24 {
				t.Fatalf("expected /24 mask, got %d", subnet.SourceNetmask)
			}
			break
		}
	}
	if !foundECS {
		t.Fatal("expected EDNS0_SUBNET option")
	}
}

func TestECSSubnet_IPv6(t *testing.T) {
	builder := NewResponseBuilder()
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeAAAA)

	builder.AddECSSubnet(m, net.ParseIP("2001:db8::1"))
	opt := m.IsEdns0()
	if opt == nil {
		t.Fatal("expected EDNS0 OPT record")
	}

	foundECS := false
	for _, o := range opt.Option {
		if subnet, ok := o.(*dns.EDNS0_SUBNET); ok {
			foundECS = true
			if subnet.Family != 2 {
				t.Fatalf("expected IPv6 family (2), got %d", subnet.Family)
			}
			if subnet.SourceNetmask != 48 {
				t.Fatalf("expected /48 mask, got %d", subnet.SourceNetmask)
			}
			break
		}
	}
	if !foundECS {
		t.Fatal("expected EDNS0_SUBNET option")
	}
}

func TestDNSSEC(t *testing.T) {
	builder := NewResponseBuilder()
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)

	builder.SetDNSSEC(m)
	opt := m.IsEdns0()
	if opt == nil {
		t.Fatal("expected EDNS0 OPT record")
	}
	if !opt.Do() {
		t.Fatal("expected DNSSEC OK (DO) bit to be set")
	}
}

// ============================================================================
// Test: ExtractRecords
// ============================================================================

func TestExtractRecords(t *testing.T) {
	m := new(dns.Msg)
	m.Response = true
	m.Rcode = dns.RcodeSuccess
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 600},
			A:   net.ParseIP("5.6.7.8").To4(),
		},
	}

	resp := ExtractRecords(m)
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Rcode != dns.RcodeSuccess {
		t.Fatalf("expected RcodeSuccess, got %d", resp.Rcode)
	}
	if len(resp.Answer) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(resp.Answer))
	}
	if resp.TTL != 300 {
		t.Fatalf("expected min TTL 300, got %d", resp.TTL)
	}
}

func TestExtractRecords_Nil(t *testing.T) {
	resp := ExtractRecords(nil)
	if resp != nil {
		t.Fatal("expected nil for nil input")
	}
}

// ============================================================================
// Test: ExtractQuestion
// ============================================================================

func TestExtractQuestion(t *testing.T) {
	m := new(dns.Msg)
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}

	q, err := ExtractQuestion(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q == nil {
		t.Fatal("expected non-nil query")
	}
	if q.Name != "example.com." {
		t.Fatalf("expected name example.com., got %s", q.Name)
	}
	if q.QType != TypeA {
		t.Fatalf("expected TypeA, got %d", q.QType)
	}
}

func TestExtractQuestion_NilMessage(t *testing.T) {
	_, err := ExtractQuestion(nil)
	if err == nil {
		t.Fatal("expected error for nil message")
	}
}

func TestExtractQuestion_NoQuestion(t *testing.T) {
	m := new(dns.Msg)
	_, err := ExtractQuestion(m)
	if err == nil {
		t.Fatal("expected error for no question section")
	}
}

// ============================================================================
// Test: BuildResponseMsg
// ============================================================================

func TestBuildResponseMsg(t *testing.T) {
	query := newTestQuery("example.com", TypeA)

	answers := []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}

	msg := BuildResponseMsg(query, answers, dns.RcodeSuccess)
	if msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("expected RcodeSuccess, got %d", msg.Rcode)
	}
	if len(msg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answer))
	}
}

// ============================================================================
// Test: QType Conversion
// ============================================================================

func TestQTypeToString(t *testing.T) {
	tests := []struct {
		qt       QueryType
		expected string
	}{
		{TypeA, "A"},
		{TypeAAAA, "AAAA"},
		{TypeCNAME, "CNAME"},
		{TypeTXT, "TXT"},
		{TypeMX, "MX"},
		{TypeSRV, "SRV"},
		{TypeNS, "NS"},
		{TypePTR, "PTR"},
		{TypeSOA, "SOA"},
		{QueryType(10000), "TYPE10000"},
	}

	for _, tt := range tests {
		got := QTypeToString(tt.qt)
		if got != tt.expected {
			t.Errorf("QTypeToString(%d): expected %q, got %q", uint16(tt.qt), tt.expected, got)
		}
	}
}

func TestStringToQType(t *testing.T) {
	tests := []struct {
		input    string
		expected QueryType
		wantErr  bool
	}{
		{"A", TypeA, false},
		{"AAAA", TypeAAAA, false},
		{"CNAME", TypeCNAME, false},
		{"TXT", TypeTXT, false},
		{"MX", TypeMX, false},
		{"SRV", TypeSRV, false},
		{"NS", TypeNS, false},
		{"PTR", TypePTR, false},
		{"SOA", TypeSOA, false},
		{"a", TypeA, false},
		{"txt", TypeTXT, false},
		{"1", TypeA, false},       // numeric
		{"28", TypeAAAA, false},   // numeric
		{"TYPE1", TypeA, false},   // TYPE format
		{"TYPE28", TypeAAAA, false}, // TYPE format
		{"", 0, true},
		{"INVALID", 0, true},
		{"ZZ_TYPE", 0, true},
	}

	for _, tt := range tests {
		got, err := StringToQType(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("StringToQType(%q): expected error, got %d", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("StringToQType(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("StringToQType(%q): expected %d, got %d", tt.input, uint16(tt.expected), uint16(got))
		}
	}
}

func TestSupportsQType(t *testing.T) {
	if !SupportsQType(TypeA) {
		t.Error("expected TypeA to be supported")
	}
	if !SupportsQType(TypeAAAA) {
		t.Error("expected TypeAAAA to be supported")
	}
	if !SupportsQType(TypeSOA) {
		t.Error("expected TypeSOA to be supported")
	}
	if SupportsQType(QueryType(99)) {
		t.Error("expected Type99 to NOT be supported")
	}
}

func TestSupportedQTypes(t *testing.T) {
	types := SupportedQTypes()
	if len(types) != 9 {
		t.Fatalf("expected 9 supported types, got %d", len(types))
	}

	// Verify all expected types are present.
	expected := map[QueryType]bool{
		TypeA: true, TypeAAAA: true, TypeCNAME: true,
		TypeTXT: true, TypeMX: true, TypeSRV: true,
		TypeNS: true, TypePTR: true, TypeSOA: true,
	}
	for _, qt := range types {
		if !expected[qt] {
			t.Errorf("unexpected type in SupportedQTypes: %d", uint16(qt))
		}
		delete(expected, qt)
	}
	if len(expected) > 0 {
		t.Errorf("missing types in SupportedQTypes: %v", expected)
	}
}

func TestIsAddressQuery(t *testing.T) {
	if !IsAddressQuery(TypeA) {
		t.Error("expected TypeA to be an address query")
	}
	if !IsAddressQuery(TypeAAAA) {
		t.Error("expected TypeAAAA to be an address query")
	}
	if IsAddressQuery(TypeCNAME) {
		t.Error("expected TypeCNAME to NOT be an address query")
	}
}

// ============================================================================
// Test: Pack / Unpack
// ============================================================================

func TestPackUnpack(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)
	m.Response = true
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}

	wire, err := PackMsg(m)
	if err != nil {
		t.Fatalf("PackMsg error: %v", err)
	}

	unpacked, err := UnpackMsg(wire)
	if err != nil {
		t.Fatalf("UnpackMsg error: %v", err)
	}

	if unpacked.Rcode != m.Rcode {
		t.Fatalf("Rcode mismatch after unpack: got %d, expected %d", unpacked.Rcode, m.Rcode)
	}
	if len(unpacked.Answer) != 1 {
		t.Fatalf("expected 1 answer after unpack, got %d", len(unpacked.Answer))
	}
}

// ============================================================================
// Test: DNS64 Synthesis
// ============================================================================

func TestDNS64Synthesis(t *testing.T) {
	prefix := net.ParseIP("64:ff9b::")
	builder := &ResponseBuilder{
		EnableDNS64:  true,
		DNS64Prefix:  prefix,
	}

	query := newTestQuery("example.com", TypeAAAA)

	// Build a response that has A records but no AAAA records.
	m := new(dns.Msg)
	m.SetReply(&dns.Msg{
		MsgHdr: dns.MsgHdr{Id: dns.Id()},
		Question: []dns.Question{
			{Name: "example.com.", Qtype: dns.TypeAAAA, Qclass: dns.ClassINET},
		},
	})
	m.Rcode = dns.RcodeSuccess
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}

	resp := &DnsResponse{
		Query:  *query,
		RawMsg: m,
		Rcode:  dns.RcodeSuccess,
		Answer: m.Answer,
	}

	msg := builder.BuildResponse(query, resp)

	// Should have the original A record plus a synthesized AAAA.
	foundAAAA := false
	foundA := false
	for _, rr := range msg.Answer {
		switch v := rr.(type) {
		case *dns.AAAA:
			foundAAAA = true
			expectedIP := net.ParseIP("64:ff9b::1.2.3.4")
			if !v.AAAA.Equal(expectedIP) {
				t.Fatalf("expected synthetic IP %s, got %s", expectedIP, v.AAAA.String())
			}
		case *dns.A:
			foundA = true
		}
	}
	if !foundA {
		t.Fatal("expected original A record to be preserved")
	}
	if !foundAAAA {
		t.Fatal("expected synthesized AAAA record")
	}
}

func TestDNS64Synthesis_Disabled(t *testing.T) {
	builder := NewResponseBuilder() // DNS64 disabled by default

	query := newTestQuery("example.com", TypeAAAA)
	m := new(dns.Msg)
	m.SetReply(&dns.Msg{
		MsgHdr: dns.MsgHdr{Id: dns.Id()},
		Question: []dns.Question{
			{Name: "example.com.", Qtype: dns.TypeAAAA, Qclass: dns.ClassINET},
		},
	})
	m.Rcode = dns.RcodeSuccess
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4").To4(),
		},
	}

	resp := &DnsResponse{
		Query:  *query,
		RawMsg: m,
		Rcode:  dns.RcodeSuccess,
		Answer: m.Answer,
	}

	msg := builder.BuildResponse(query, resp)

	// DNS64 disabled, should only have A record, no AAAA synthesized.
	for _, rr := range msg.Answer {
		if _, ok := rr.(*dns.AAAA); ok {
			t.Fatal("expected no AAAA synthesis when DNS64 is disabled")
		}
	}
}

// ============================================================================
// Test: ResponseTime
// ============================================================================

func TestResponseTime(t *testing.T) {
	// With RTT > 0, should return the RTT.
	rtt := ResponseTime(time.Now(), 50*time.Millisecond)
	if rtt != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", rtt)
	}

	// With zero RTT, should return time since queryTime.
	start := time.Now().Add(-100 * time.Millisecond)
	rtt = ResponseTime(start, 0)
	if rtt < 50*time.Millisecond || rtt > 200*time.Millisecond {
		t.Fatalf("expected ~100ms, got %v", rtt)
	}
}

// ============================================================================
// Test: NeedsDNS64 / IsAddressQuery
// ============================================================================

func TestNeedsDNS64(t *testing.T) {
	if !NeedsDNS64(TypeAAAA) {
		t.Error("expected TypeAAAA to need DNS64")
	}
	if NeedsDNS64(TypeA) {
		t.Error("expected TypeA to NOT need DNS64")
	}
	if NeedsDNS64(TypeCNAME) {
		t.Error("expected TypeCNAME to NOT need DNS64")
	}
}
