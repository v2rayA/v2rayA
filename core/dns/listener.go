package dns

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// Handler is the interface for processing DNS queries.
// Implementations must be safe for concurrent use.
type Handler interface {
	// HandleQuery processes a DNS query and returns a response.
	// The context may carry a deadline for timeout control.
	HandleQuery(ctx context.Context, query *DnsQuery) (*DnsResponse, error)
}

// serverPair holds UDP and TCP server instances for a single listen address.
type serverPair struct {
	udp *dns.Server
	tcp *dns.Server
}

// DnsListener listens for DNS queries on one or more addresses (UDP + TCP per address).
type DnsListener struct {
	config  *DnsListenerConfig
	handler Handler
	servers []serverPair
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewDnsListener creates a new DnsListener with the given config and handler.
func NewDnsListener(config *DnsListenerConfig, handler Handler) *DnsListener {
	return &DnsListener{
		config:  config,
		handler: handler,
	}
}

// Start starts UDP and TCP listeners for the primary address and all extra addresses.
func (l *DnsListener) Start() error {
	if l.config == nil {
		return fmt.Errorf("dns listener: config is nil")
	}
	if l.handler == nil {
		return fmt.Errorf("dns listener: handler is nil")
	}
	if err := l.config.Validate(); err != nil {
		return fmt.Errorf("dns listener: invalid config: %w", err)
	}

	l.ctx, l.cancel = context.WithCancel(context.Background())

	// Collect all listen addresses (primary + extras).
	allAddrs := []string{l.config.ListenAddr}
	allAddrs = append(allAddrs, l.config.ExtraListenAddrs...)

	for _, addr := range allAddrs {
		pair, err := l.startPair(addr)
		if err != nil {
			// Clean up already started listeners.
			l.Stop()
			return fmt.Errorf("dns listener: %s: %w", addr, err)
		}
		l.servers = append(l.servers, *pair)
	}

	return nil
}

// startPair starts UDP and TCP listeners for a single address.
func (l *DnsListener) startPair(addr string) (*serverPair, error) {
	timeout := time.Duration(l.config.Timeout) * time.Second

	udpSrv := &dns.Server{
		Addr:         addr,
		Net:          "udp",
		Handler:      dns.HandlerFunc(l.handlePacket),
		UDPSize:      1452,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		log.Printf("[dns] UDP listener starting on %s", addr)
		if err := udpSrv.ListenAndServe(); err != nil {
			log.Printf("[dns] UDP listener stopped: %v", err)
		}
	}()

	tcpSrv := &dns.Server{
		Addr:         addr,
		Net:          "tcp",
		Handler:      dns.HandlerFunc(l.handlePacket),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		log.Printf("[dns] TCP listener starting on %s", addr)
		if err := tcpSrv.ListenAndServe(); err != nil {
			log.Printf("[dns] TCP listener stopped: %v", err)
		}
	}()

	return &serverPair{udp: udpSrv, tcp: tcpSrv}, nil
}

// Stop gracefully shuts down all listeners.
func (l *DnsListener) Stop() error {
	if l.cancel != nil {
		l.cancel()
	}

	var lastErr error
	for _, pair := range l.servers {
		if pair.udp != nil {
			if err := pair.udp.Shutdown(); err != nil {
				log.Printf("[dns] UDP shutdown error: %v", err)
				lastErr = err
			}
		}
		if pair.tcp != nil {
			if err := pair.tcp.Shutdown(); err != nil {
				log.Printf("[dns] TCP shutdown error: %v", err)
				lastErr = err
			}
		}
	}

	l.wg.Wait()
	l.servers = nil

	if lastErr != nil {
		return fmt.Errorf("dns listener: shutdown error: %w", lastErr)
	}
	return nil
}

// handlePacket is the callback invoked by miekg/dns for each incoming query.
func (l *DnsListener) handlePacket(w dns.ResponseWriter, msg *dns.Msg) {
	if msg == nil {
		return
	}
	if len(msg.Question) == 0 {
		log.Printf("[dns] received message with no questions from %s", w.RemoteAddr())
		return
	}

	q := msg.Question[0]
	clientAddr := w.RemoteAddr()

	query := &DnsQuery{
		Name:  q.Name,
		QType: QueryType(q.Qtype),
	}

	switch addr := clientAddr.(type) {
	case *net.UDPAddr:
		query.ClientIP = addr.IP
	case *net.TCPAddr:
		query.ClientIP = addr.IP
	}

	log.Printf("[dns] query: %s %s from %s", dns.Type(q.Qtype).String(), q.Name, clientAddr.String())

	ctx, cancel := context.WithTimeout(l.ctx, time.Duration(l.config.Timeout)*time.Second)
	defer cancel()

	resp, err := l.handler.HandleQuery(ctx, query)
	if err != nil {
		log.Printf("[dns] error handling query %s %s: %v", dns.Type(q.Qtype).String(), q.Name, err)
		m := new(dns.Msg)
		m.SetReply(msg)
		m.Rcode = dns.RcodeServerFailure
		_ = w.WriteMsg(m)
		return
	}

	m := new(dns.Msg)
	m.SetReply(msg)

	if resp != nil && resp.RawMsg != nil {
		// Deep-copy the upstream response so we don't modify the cached RawMsg in-place.
		// This is critical: resp.RawMsg may be shared with the cache (via cache.Set),
		// and m.SetReply(msg) below would corrupt the cached message.
		rawBytes, err := resp.RawMsg.Pack()
		if err != nil {
			log.Printf("[dns] error packing response for %s %s: %v", dns.Type(q.Qtype).String(), q.Name, err)
			m := new(dns.Msg)
			m.SetReply(msg)
			m.Rcode = dns.RcodeServerFailure
			_ = w.WriteMsg(m)
			return
		}
		m = new(dns.Msg)
		if err := m.Unpack(rawBytes); err != nil {
			log.Printf("[dns] error unpacking response for %s %s: %v", dns.Type(q.Qtype).String(), q.Name, err)
			m := new(dns.Msg)
			m.SetReply(msg)
			m.Rcode = dns.RcodeServerFailure
			_ = w.WriteMsg(m)
			return
		}
		// Set the response header to match the client's request (ID, flags, question).
		// Preserve the upstream Rcode — SetReply always sets Rcode=RcodeSuccess,
		// but we need to keep NXDOMAIN etc. from the upstream.
		origRcode := m.Rcode
		m.SetReply(msg)
		m.Rcode = origRcode
	} else if resp != nil {
		m.Rcode = resp.Rcode
		m.Answer = resp.Answer
		m.Ns = resp.Authority
		m.Extra = resp.Additional
	}

	if err := w.WriteMsg(m); err != nil {
		log.Printf("[dns] error writing response to %s: %v", clientAddr, err)
	}
}

// Healthy returns true if all listeners are active.
func (l *DnsListener) Healthy() bool {
	if len(l.servers) == 0 {
		return false
	}
	select {
	case <-l.ctx.Done():
		return false
	default:
		return true
	}
}
