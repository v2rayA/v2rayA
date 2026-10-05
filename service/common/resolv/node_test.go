package resolv

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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

func TestLookupNodeTransports(t *testing.T) {
	// The same local certificate verifies IP identity for DoT and DoH.
	certServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns-query" || r.URL.RawQuery != "test=1" || r.Method != "POST" {
			t.Errorf("wrong DoH endpoint: %s", r.URL)
		}
		packet, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		msg := new(dns.Msg)
		if err := msg.Unpack(packet); err != nil {
			t.Error(err)
			return
		}
		packet, _ = nodeAnswer(msg).Pack()
		w.Header().Set("Content-Type", "application/dns-message")
		w.Write(packet)
	}))
	defer certServer.Close()
	roots := x509.NewCertPool()
	roots.AddCert(certServer.Certificate())
	for _, scheme := range []string{"udp", "tcp", "tls", "https"} {
		t.Run(scheme, func(t *testing.T) {
			endpointURL := certServer.URL + "/dns-query?test=1"
			var count atomic.Int32
			if scheme != "https" {
				server := &dns.Server{Handler: dns.HandlerFunc(func(w dns.ResponseWriter, msg *dns.Msg) { count.Add(1); w.WriteMsg(nodeAnswer(msg)) })}
				if scheme == "udp" {
					pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
					if err != nil {
						t.Fatal(err)
					}
					server.PacketConn = pc
					endpointURL = "udp://" + pc.LocalAddr().String()
				} else {
					listener, err := net.Listen("tcp4", "127.0.0.1:0")
					if err != nil {
						t.Fatal(err)
					}
					if scheme == "tls" {
						listener = tls.NewListener(listener, certServer.TLS)
					}
					server.Listener = listener
					endpointURL = scheme + "://" + listener.Addr().String()
				}
				server.NotifyStartedFunc = func() {}
				started := make(chan struct{})
				server.NotifyStartedFunc = func() { close(started) }
				go server.ActivateAndServe()
				<-started
				defer server.Shutdown()
			}
			endpoint, err := ParseIPDNS(endpointURL)
			if err != nil {
				t.Fatal(err)
			}
			var sockets atomic.Int32
			dialer := &net.Dialer{Timeout: time.Second, Control: func(_, address string, _ syscall.RawConn) error {
				if address != endpoint.Address() {
					t.Errorf("query left selected endpoint: %s", address)
				}
				sockets.Add(1)
				return nil
			}}
			ips, err := lookupNode("node.invalid", endpoint, dialer, roots)
			if err != nil || !reflect.DeepEqual(ips, []string{"192.0.2.1", "2001:db8::1"}) || sockets.Load() == 0 {
				t.Fatalf("got %v %v, sockets %d", ips, err, sockets.Load())
			}
			before := sockets.Load()
			ips, err = lookupNode("203.0.113.1", endpoint, dialer, roots)
			if err != nil || !reflect.DeepEqual(ips, []string{"203.0.113.1"}) || sockets.Load() != before {
				t.Fatal("IP node performed DNS query")
			}
			if _, err := lookupNode("failed.invalid", endpoint, dialer, roots); err == nil || !strings.Contains(err.Error(), "SERVFAIL") {
				t.Fatalf("query failure hidden: %v", err)
			}
			if scheme == "tls" || scheme == "https" {
				if _, err := LookupNode("node.invalid", endpoint, dialer); err == nil {
					t.Fatal("untrusted TLS certificate accepted")
				}
			}
		})
	}
}
