package resolv

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// nodeQueryOption scopes service/QUIC queries to rule-node and the current module.
const nodeQueryOption = 65002

// LookupNode queries the local module without consulting the system resolver.
func LookupNode(ctx context.Context, host, address, token string, dialer *net.Dialer) ([]string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []string{ip.String()}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	d := *dialer
	d.Resolver = nil
	exchange := func(msg *dns.Msg) (*dns.Msg, error) {
		c := &dns.Client{Net: "udp", Dialer: &d, Timeout: 3 * time.Second, UDPSize: 4096}
		resp, _, err := c.ExchangeContext(ctx, msg, address)
		if err == nil && resp.Truncated {
			c.Net = "tcp"
			resp, _, err = c.ExchangeContext(ctx, msg, address)
		}
		return resp, err
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
				msg.SetEdns0(4096, false)
				msg.IsEdns0().Option = append(msg.IsEdns0().Option, &dns.EDNS0_LOCAL{Code: nodeQueryOption, Data: []byte(token)})
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
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("node DNS %s: %w", host, err)
	}
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
