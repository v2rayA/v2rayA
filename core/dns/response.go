package dns

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

// ResponseBuilder builds DNS response messages.
// It handles all supported record types, response compression,
// EDNS0 options, TC bit handling, and DNS64 synthesis.
type ResponseBuilder struct {
	// EnableDNS64, when true, enables DNS64 synthetic AAAA record
	// synthesis from A records (stub for future implementation).
	EnableDNS64 bool
	// DNS64Prefix is the NAT64 prefix used for DNS64 synthesis.
	DNS64Prefix net.IP
}

// NewResponseBuilder creates a new ResponseBuilder.
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{}
}

// BuildResponse constructs a final *dns.Msg from an upstream DNS response.
// It copies the question section, sets response flags, applies compression,
// and preserves EDNS0 (OPT) records from the upstream response.
func (b *ResponseBuilder) BuildResponse(query *DnsQuery, upstreamResp *DnsResponse) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(upstreamResp.RawMsg)

	// Ensure the question matches the original query.
	m.Question = []dns.Question{
		{
			Name:   dns.Fqdn(query.Name),
			Qtype:  uint16(query.QType),
			Qclass: query.QClass,
		},
	}

	// Copy response flags from upstream.
	m.Response = true
	m.Rcode = upstreamResp.Rcode
	m.RecursionAvailable = upstreamResp.RawMsg.RecursionAvailable
	m.Authoritative = upstreamResp.RawMsg.Authoritative
	m.AuthenticatedData = upstreamResp.RawMsg.AuthenticatedData
	m.CheckingDisabled = upstreamResp.RawMsg.CheckingDisabled

	// Copy resource records.
	if len(upstreamResp.Answer) > 0 {
		m.Answer = make([]dns.RR, len(upstreamResp.Answer))
		copy(m.Answer, upstreamResp.Answer)
	}
	if len(upstreamResp.Authority) > 0 {
		m.Ns = make([]dns.RR, len(upstreamResp.Authority))
		copy(m.Ns, upstreamResp.Authority)
	}
	if len(upstreamResp.Additional) > 0 {
		m.Extra = make([]dns.RR, len(upstreamResp.Additional))
		copy(m.Extra, upstreamResp.Additional)
	}

	// Enable compression.
	m.Compress = true

	// Handle DNS64 synthesis if enabled.
	if b.EnableDNS64 && m.Rcode == dns.RcodeSuccess {
		b.synthesizeDNS64(m)
	}

	return m
}

// BuildRefused creates a REFUSED DNS response message.
func (b *ResponseBuilder) BuildRefused(query *DnsQuery) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(query.Name), uint16(query.QType))
	m.Response = true
	m.Rcode = dns.RcodeRefused
	m.RecursionAvailable = true
	m.Compress = true
	return m
}

// BuildNXDOMAIN creates an NXDOMAIN DNS response message.
func (b *ResponseBuilder) BuildNXDOMAIN(query *DnsQuery) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(query.Name), uint16(query.QType))
	m.Response = true
	m.Rcode = dns.RcodeNameError
	m.RecursionAvailable = true
	m.Compress = true
	return m
}

// BuildServerFailure creates a SERVFAIL DNS response message.
func (b *ResponseBuilder) BuildServerFailure(query *DnsQuery) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(query.Name), uint16(query.QType))
	m.Response = true
	m.Rcode = dns.RcodeServerFailure
	m.RecursionAvailable = true
	m.Compress = true
	return m
}

// SetTruncated sets the TC (Truncation) bit on the message.
// This is used to indicate UDP truncation to the client.
func (b *ResponseBuilder) SetTruncated(m *dns.Msg) {
	m.Truncated = true
}

// IsTruncated returns true if the message has the TC bit set.
func (b *ResponseBuilder) IsTruncated(m *dns.Msg) bool {
	return m.Truncated
}

// AddEDNS0 adds or updates an EDNS0 (OPT) pseudo-record to the message.
// If the message already has an OPT record, it updates its UDP payload size;
// otherwise it appends a new OPT record.
func (b *ResponseBuilder) AddEDNS0(m *dns.Msg, udpPayloadSize uint16, options ...dns.EDNS0) {
	opt := m.IsEdns0()
	if opt != nil {
		opt.SetUDPSize(udpPayloadSize)
		// Preserve existing options and append new ones.
		for _, o := range options {
			opt.Option = append(opt.Option, o)
		}
		return
	}

	// Create a new OPT record.
	extra := []dns.RR{
		&dns.OPT{
			Hdr: dns.RR_Header{
				Name:   ".",
				Rrtype: dns.TypeOPT,
				Class:  udpPayloadSize,
			},
			Option: options,
		},
	}
	m.Extra = append(m.Extra, extra...)
}

// AddECSSubnet adds an EDNS0 Client Subnet (ECS) option to the message.
// ECS allows the upstream to return location-specific responses.
func (b *ResponseBuilder) AddECSSubnet(m *dns.Msg, clientIP net.IP) {
	if clientIP == nil {
		return
	}

	var family uint16
	var mask uint8
	var ip net.IP

	if ipv4 := clientIP.To4(); ipv4 != nil {
		family = 1 // IPv4
		mask = 24
		ip = ipv4
	} else {
		family = 2 // IPv6
		mask = 48
		ip = clientIP.To16()
	}

	subnet := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        family,
		SourceNetmask: mask,
		SourceScope:   0,
		Address:       ip,
	}

	b.AddEDNS0(m, 4096, subnet)
}

// SetDNSSEC sets the DNSSEC OK (DO) bit on the message.
func (b *ResponseBuilder) SetDNSSEC(m *dns.Msg) {
	opt := m.IsEdns0()
	if opt != nil {
		opt.SetDo()
		return
	}
	// Create new OPT with DO bit set.
	extra := []dns.RR{
		&dns.OPT{
			Hdr: dns.RR_Header{
				Name:   ".",
				Rrtype: dns.TypeOPT,
				Class:  dns.DefaultMsgSize,
			},
		},
	}
	m.Extra = append(m.Extra, extra...)
	m.Extra[len(m.Extra)-1].(*dns.OPT).SetDo()
}

// ExtractRecords extracts resource records from a *dns.Msg into a DnsResponse.
// It populates Answer, Authority, Additional, Rcode, and computes the
// minimum TTL across answer records.
func ExtractRecords(msg *dns.Msg) *DnsResponse {
	if msg == nil {
		return nil
	}

	resp := &DnsResponse{
		RawMsg:     msg,
		Rcode:      msg.Rcode,
		Answer:     msg.Answer,
		Authority:  msg.Ns,
		Additional: msg.Extra,
	}

	// Compute minimum TTL across answer records.
	if len(msg.Answer) > 0 {
		resp.TTL = msg.Answer[0].Header().Ttl
		for _, rr := range msg.Answer[1:] {
			if rr.Header().Ttl < resp.TTL {
				resp.TTL = rr.Header().Ttl
			}
		}
	}

	return resp
}

// ExtractQuestion extracts a DnsQuery from a *dns.Msg.
func ExtractQuestion(msg *dns.Msg) (*DnsQuery, error) {
	if msg == nil {
		return nil, fmt.Errorf("dns response: nil message")
	}
	if len(msg.Question) == 0 {
		return nil, fmt.Errorf("dns response: no question section")
	}

	q := msg.Question[0]
	return &DnsQuery{
		Name:   dns.Fqdn(q.Name),
		QType:  QueryType(q.Qtype),
		QClass: q.Qclass,
	}, nil
}

// BuildResponseMsg constructs a *dns.Msg from DnsQuery and answer resource records.
// This is a convenience helper that creates a response directly from records,
// without needing an upstream response.
func BuildResponseMsg(query *DnsQuery, answer []dns.RR, rcode int) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(&dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id: dns.Id(),
		},
		Question: []dns.Question{
			{
				Name:   dns.Fqdn(query.Name),
				Qtype:  uint16(query.QType),
				Qclass: query.QClass,
			},
		},
	})
	m.Response = true
	m.Rcode = rcode
	m.RecursionAvailable = true
	m.Compress = true

	if len(answer) > 0 {
		m.Answer = answer
	}

	return m
}

// PackMsg serializes a *dns.Msg into wire format bytes.
// It handles compression and returns the packed bytes.
func PackMsg(m *dns.Msg) ([]byte, error) {
	wire, err := m.Pack()
	if err != nil {
		return nil, fmt.Errorf("dns response: pack error: %w", err)
	}
	return wire, nil
}

// UnpackMsg deserializes wire format bytes into a *dns.Msg.
func UnpackMsg(wire []byte) (*dns.Msg, error) {
	m := new(dns.Msg)
	if err := m.Unpack(wire); err != nil {
		return nil, fmt.Errorf("dns response: unpack error: %w", err)
	}
	return m, nil
}

// handleTruncation checks if the response is truncated and, if so,
// sets the TC bit and clears all resource records except the question.
// This follows the standard DNS truncation behavior for UDP.
func (b *ResponseBuilder) handleTruncation(m *dns.Msg, maxSize int) {
	if m.Len() > maxSize {
		m.Truncated = true
		m.Answer = nil
		m.Ns = nil
		m.Extra = nil
	}
}

// synthesizeDNS64 performs DNS64 synthetic AAAA record synthesis (stub).
// When enabled, if the response has A records but no AAAA records,
// this synthesizes AAAA records using the configured NAT64 prefix.
func (b *ResponseBuilder) synthesizeDNS64(m *dns.Msg) {
	if !b.EnableDNS64 || b.DNS64Prefix == nil {
		return
	}

	// Check if response already has AAAA records.
	hasAAAA := false
	for _, rr := range m.Answer {
		if rr.Header().Rrtype == dns.TypeAAAA {
			hasAAAA = true
			break
		}
	}
	if hasAAAA {
		return
	}

	// Check if response has A records to synthesize from.
	var aRecords []*dns.A
	for _, rr := range m.Answer {
		if a, ok := rr.(*dns.A); ok {
			aRecords = append(aRecords, a)
		}
	}
	if len(aRecords) == 0 {
		return
	}

	// Synthesize AAAA records using NAT64 prefix.
	prefix := b.DNS64Prefix.To16()
	if prefix == nil {
		return
	}

	for _, a := range aRecords {
		ipv4 := a.A.To4()
		if ipv4 == nil {
			continue
		}

		// Construct IPv4-embedded IPv6 address.
		syntheticIP := make(net.IP, 16)
		copy(syntheticIP, prefix)
		syntheticIP[12] = ipv4[0]
		syntheticIP[13] = ipv4[1]
		syntheticIP[14] = ipv4[2]
		syntheticIP[15] = ipv4[3]

		aaaa := &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   a.Hdr.Name,
				Rrtype: dns.TypeAAAA,
				Class:  a.Hdr.Class,
				Ttl:    a.Hdr.Ttl,
			},
			AAAA: syntheticIP,
		}
		m.Answer = append(m.Answer, aaaa)
	}
}

// ResponseTime returns the appropriate duration for the DNS response time.
// It calculates the response time based on when the query was received
// and the upstream response RTT.
func ResponseTime(queryTime time.Time, rtt time.Duration) time.Duration {
	if rtt > 0 {
		return rtt
	}
	return time.Since(queryTime)
}
