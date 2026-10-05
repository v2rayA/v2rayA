package resolv

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// LookupNode queries only this endpoint, preserving the caller's socket policy.
// A failed query never falls through to the service's general-purpose resolver.
func LookupNode(host string, endpoint *IPDNSEndpoint, dialer *net.Dialer) ([]string, error) {
	return lookupNode(host, endpoint, dialer, nil)
}

func lookupNode(host string, endpoint *IPDNSEndpoint, dialer *net.Dialer, roots *x509.CertPool) ([]string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []string{ip.String()}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	d := *dialer
	d.Resolver = nil
	transport := &http.Transport{DialContext: d.DialContext, TLSClientConfig: &tls.Config{ServerName: endpoint.IP.String(), RootCAs: roots}, ForceAttemptHTTP2: true}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	exchange := func(msg *dns.Msg) (*dns.Msg, error) {
		if endpoint.Scheme != "https" {
			c := &dns.Client{Net: endpoint.Protocol(), Dialer: &d, Timeout: 3 * time.Second, UDPSize: 4096, TLSConfig: &tls.Config{ServerName: endpoint.IP.String(), RootCAs: roots}}
			resp, _, err := c.ExchangeContext(ctx, msg, endpoint.Address())
			if err == nil && resp.Truncated && endpoint.Scheme == "udp" {
				c.Net = "tcp"
				resp, _, err = c.ExchangeContext(ctx, msg, endpoint.Address())
			}
			return resp, err
		}
		packet, err := msg.Pack()
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(packet))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/dns-message")
		req.Header.Set("Accept", "application/dns-message")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("DoH: HTTP %d", resp.StatusCode)
		}
		if strings.Split(resp.Header.Get("Content-Type"), ";")[0] != "application/dns-message" {
			return nil, fmt.Errorf("DoH: invalid content type")
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, dns.MaxMsgSize+1))
		if err != nil {
			return nil, err
		}
		if len(body) > dns.MaxMsgSize {
			return nil, fmt.Errorf("DoH response too large")
		}
		answer := new(dns.Msg)
		return answer, answer.Unpack(body)
	}
	var results [2][]string
	var failures [2]error
	var wg sync.WaitGroup
	for i, qt := range []uint16{dns.TypeA, dns.TypeAAAA} {
		wg.Add(1)
		go func(i int, qt uint16) {
			defer wg.Done()
			name := dns.Fqdn(host)
			for depth := 0; depth < 8; depth++ {
				msg := new(dns.Msg)
				msg.SetQuestion(name, qt)
				resp, err := exchange(msg)
				if err != nil {
					failures[i] = err
					return
				}
				if resp == nil || !resp.Response || resp.Id != msg.Id || len(resp.Question) != 1 || resp.Question[0] != msg.Question[0] {
					failures[i] = fmt.Errorf("invalid DNS response")
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					failures[i] = fmt.Errorf("DNS response: %s", dns.RcodeToString[resp.Rcode])
					return
				}
				// Follow only the answer's CNAME chain, ignoring unrelated records.
				target := name
				for j := 0; j < len(resp.Answer); j++ {
					changed := false
					for _, rr := range resp.Answer {
						if cname, ok := rr.(*dns.CNAME); ok && strings.EqualFold(cname.Hdr.Name, target) {
							target = cname.Target
							changed = true
							break
						}
					}
					if !changed {
						break
					}
				}
				for _, rr := range resp.Answer {
					if !strings.EqualFold(rr.Header().Name, target) {
						continue
					}
					switch rr := rr.(type) {
					case *dns.A:
						if qt == dns.TypeA {
							results[i] = append(results[i], rr.A.String())
						}
					case *dns.AAAA:
						if qt == dns.TypeAAAA {
							results[i] = append(results[i], rr.AAAA.String())
						}
					}
				}
				if len(results[i]) > 0 || strings.EqualFold(name, target) {
					return
				}
				name = target
			}
			failures[i] = fmt.Errorf("DNS CNAME chain too long")
		}(i, qt)
	}
	wg.Wait()
	out := append(results[0], results[1]...)
	if len(out) > 0 {
		return out, nil
	}
	for _, err := range failures {
		if err != nil {
			return nil, fmt.Errorf("node DNS %s: %w", host, err)
		}
	}
	return nil, fmt.Errorf("node DNS %s: no A or AAAA addresses", host)
}
