package dns

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/miekg/dns"
)

type testRouteDispatcher func(context.Context, string, string, string) (net.Conn, error)

func (f testRouteDispatcher) Dispatch(ctx context.Context, network, addr, tag string) (net.Conn, error) {
	return f(ctx, network, addr, tag)
}

func TestDispatcherTransport(t *testing.T) {
	for _, tc := range []struct {
		name, protocol                          string
		truncated, wrongID, wrongQuestion, fail bool
		want                                    []string
	}{
		{name: "default UDP", want: []string{"udp"}},
		{name: "explicit UDP", protocol: "udp", want: []string{"udp"}},
		{name: "TCP", protocol: "tcp", want: []string{"tcp"}},
		{name: "truncated", truncated: true, want: []string{"udp", "tcp"}},
		{name: "wrong ID", truncated: true, wrongID: true, want: []string{"udp"}},
		{name: "wrong question", truncated: true, wrongQuestion: true, want: []string{"udp"}},
		{name: "dispatch error", fail: true, want: []string{"udp", "udp"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var networks []string
			mgr := NewUpstreamManager(nil)
			mgr.SetDispatcher(testRouteDispatcher(func(ctx context.Context, network, addr, tag string) (net.Conn, error) {
				networks = append(networks, network)
				if tag != "chosen-group" || addr != "192.0.2.1:53" {
					t.Errorf("route = %s %s", tag, addr)
				}
				if tc.fail {
					return nil, errors.New("unavailable")
				}
				client, server := net.Pipe()
				go func() {
					defer server.Close()
					var q *dns.Msg
					var err error
					if network == "udp" {
						b := make([]byte, 65535)
						var n int
						n, err = server.Read(b)
						q = new(dns.Msg)
						if err == nil {
							err = q.Unpack(b[:n])
						}
					} else {
						q, err = (&dns.Conn{Conn: server}).ReadMsg()
					}
					if err != nil {
						return
					}
					r := new(dns.Msg)
					r.SetReply(q)
					r.Truncated = tc.truncated && network == "udp"
					if tc.wrongID {
						r.Id++
					}
					if tc.wrongQuestion {
						r.Question[0].Name = "other.example."
					}
					if network == "udp" {
						b, _ := r.Pack()
						server.Write(b)
					} else {
						(&dns.Conn{Conn: server}).WriteMsg(r)
					}
				}()
				return client, nil
			}))
			_, err := mgr.Exchange(&UpstreamInstance{Addr: "192.0.2.1:53", ProxyTag: "chosen-group", Protocol: tc.protocol}, &DnsQuery{Name: "example.com", QType: TypeA})
			if (err != nil) != (tc.fail || tc.wrongID || tc.wrongQuestion) {
				t.Fatalf("exchange error = %v", err)
			}
			if !reflect.DeepEqual(networks, tc.want) {
				t.Fatalf("networks = %v, want %v", networks, tc.want)
			}
		})
	}
}

// Ignore deadlines like Xray's connection: only Close can interrupt this read.
type noDeadlineConn struct{ net.Conn }

func (noDeadlineConn) SetDeadline(time.Time) error      { return nil }
func (noDeadlineConn) SetReadDeadline(time.Time) error  { return nil }
func (noDeadlineConn) SetWriteDeadline(time.Time) error { return nil }

func TestDispatchedCancellation(t *testing.T) {
	for _, network := range []string{"udp", "tcp"} {
		t.Run(network, func(t *testing.T) {
			client, server := net.Pipe()
			defer server.Close()
			go io.Copy(io.Discard, server)
			mgr := NewUpstreamManager(nil)
			mgr.SetDispatcher(testRouteDispatcher(func(context.Context, string, string, string) (net.Conn, error) { return noDeadlineConn{client}, nil }))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			q := new(dns.Msg)
			q.SetQuestion("example.com.", dns.TypeA)
			done := make(chan error, 1)
			go func() { _, err := mgr.exchangeDispatched(ctx, &UpstreamInstance{}, q, network); done <- err }()
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("expected cancellation error")
				}
			case <-time.After(time.Second):
				client.Close()
				t.Fatal("exchange did not stop")
			}
		})
	}
}
