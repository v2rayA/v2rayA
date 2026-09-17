package dns

import (
	"testing"

	"github.com/miekg/dns"
)

func TestMarkOutgoingIsRecognisedBack(t *testing.T) {
	msg := new(dns.Msg)
	msg.SetQuestion("example.com.", dns.TypeA)
	if carriesOwnToken(msg) {
		t.Fatal("a fresh query must not look like one of ours")
	}
	markOutgoing(msg)
	if !carriesOwnToken(msg) {
		t.Fatal("a query we marked must be recognised when it comes back")
	}
	if opts := msg.IsEdns0(); opts == nil {
		t.Fatal("marking must leave an OPT record")
	}
}

func TestMarkOutgoingKeepsOneOptRecord(t *testing.T) {
	msg := new(dns.Msg)
	msg.SetQuestion("example.com.", dns.TypeA)
	msg.SetEdns0(4096, true)
	markOutgoing(msg)
	markOutgoing(msg)
	count := 0
	for _, rr := range msg.Extra {
		if _, ok := rr.(*dns.OPT); ok {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d OPT records, want 1 — two make a server answer FORMERR", count)
	}
	opt := msg.IsEdns0()
	marks := 0
	for _, o := range opt.Option {
		if local, ok := o.(*dns.EDNS0_LOCAL); ok && local.Code == loopOptionCode {
			marks++
		}
	}
	if marks != 1 {
		t.Fatalf("got %d loop options, want 1", marks)
	}
}

func TestAnotherInstancesTokenIsNotOurs(t *testing.T) {
	msg := new(dns.Msg)
	msg.SetQuestion("example.com.", dns.TypeA)
	msg.SetEdns0(4096, true)
	opt := msg.IsEdns0()
	opt.Option = append(opt.Option, &dns.EDNS0_LOCAL{
		Code: loopOptionCode,
		Data: []byte("notmine!"),
	})
	if carriesOwnToken(msg) {
		t.Fatal("another instance's token must not count as a loop")
	}
}

func TestOtherEdnsOptionsAreIgnored(t *testing.T) {
	msg := new(dns.Msg)
	msg.SetQuestion("example.com.", dns.TypeA)
	msg.SetEdns0(4096, true)
	opt := msg.IsEdns0()
	opt.Option = append(opt.Option, &dns.EDNS0_SUBNET{Code: dns.EDNS0SUBNET})
	if carriesOwnToken(msg) {
		t.Fatal("an ECS option must not be read as a loop token")
	}
	markOutgoing(msg)
	if !carriesOwnToken(msg) {
		t.Fatal("marking next to other options must still work")
	}
}
