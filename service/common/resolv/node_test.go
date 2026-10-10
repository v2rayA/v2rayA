package resolv

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func nodeAnswer(msg *dns.Msg) *dns.Msg {
	answer := new(dns.Msg)
	answer.SetReply(msg)
	name := msg.Question[0].Name
	if name == "failed.invalid." {
		answer.Rcode = dns.RcodeServerFailure
		return answer
	}
	if name == "node.invalid." {
		rr, _ := dns.NewRR(name + " 60 IN CNAME resolved.invalid.")
		answer.Answer = append(answer.Answer, rr)
		name = "resolved.invalid."
	}
	kind, address := "A", "192.0.2.1"
	if msg.Question[0].Qtype == dns.TypeAAAA {
		kind, address = "AAAA", "2001:db8::1"
	}
	rr, _ := dns.NewRR(fmt.Sprintf("%s 60 IN %s %s", name, kind, address))
	answer.Answer = append(answer.Answer, rr)
	return answer
}

func TestLookupNodeLocalModule(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	pc, err := net.ListenPacket("udp4", address)
	if err != nil {
		listener.Close()
		t.Fatal(err)
	}
	var truncated atomic.Bool
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, msg *dns.Msg) {
		opt := msg.IsEdns0()
		if opt == nil || len(opt.Option) != 1 {
			t.Error("missing module identity")
		} else if local, ok := opt.Option[0].(*dns.EDNS0_LOCAL); !ok || local.Code != 65002 || string(local.Data) != "current" {
			t.Error("wrong module identity")
		}
		answer := nodeAnswer(msg)
		if _, udp := w.RemoteAddr().(*net.UDPAddr); udp && truncated.Load() {
			answer.Answer = nil
			answer.Truncated = true
		}
		_ = w.WriteMsg(answer)
	})
	for _, server := range []*dns.Server{{PacketConn: pc, Handler: handler}, {Listener: listener, Handler: handler}} {
		ready := make(chan struct{})
		server.NotifyStartedFunc = func() { close(ready) }
		go server.ActivateAndServe()
		<-ready
		defer server.Shutdown()
	}
	var sockets atomic.Int32
	dialer := &net.Dialer{Timeout: time.Second, Control: func(_, destination string, _ syscall.RawConn) error {
		if destination != address {
			t.Errorf("query left local module: %s", destination)
		}
		sockets.Add(1)
		return nil
	}}
	for _, fallback := range []bool{false, true} {
		truncated.Store(fallback)
		ips, err := LookupNode(context.Background(), "node.invalid", address, "current", dialer)
		if err != nil || !reflect.DeepEqual(ips, []string{"192.0.2.1", "2001:db8::1"}) {
			t.Fatalf("got %v %v", ips, err)
		}
	}
	before := sockets.Load()
	if ips, err := LookupNode(context.Background(), "203.0.113.1", address, "current", dialer); err != nil || !reflect.DeepEqual(ips, []string{"203.0.113.1"}) || sockets.Load() != before {
		t.Fatal("IP node performed DNS query")
	}
	truncated.Store(false)
	if _, err := LookupNode(context.Background(), "failed.invalid", address, "current", dialer); err == nil || !strings.Contains(err.Error(), "SERVFAIL") {
		t.Fatalf("query failure hidden: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := LookupNode(ctx, "node.invalid", address, "current", dialer); err == nil {
		t.Fatal("cancelled query succeeded")
	}
}
