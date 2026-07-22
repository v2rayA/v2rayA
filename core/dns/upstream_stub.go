package dns

import (
	"context"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// ProxyAddrResolver is a function type that resolves a proxy tag to a
// SOCKS5/HTTP proxy address. This allows external code (e.g., v2rayA's
// process manager) to provide the proxy address lookup logic.
// Returns the proxy address (e.g., "127.0.0.1:1080") or empty string
// if the tag is unknown.
type ProxyAddrResolver func(proxyTag string) string

// SetProxyAddrResolver sets the proxy address resolver function.
// This is used by exchangeViaProxy to determine how to route DNS
// queries through a specific proxy channel.
func (m *UpstreamManager) SetProxyAddrResolver(resolver ProxyAddrResolver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.proxyAddrResolver = resolver
}

// SetDispatcher sets the xray-core routing dispatcher for internal routing.
// When set, proxy-tagged upstreams use xray-core's internal routing
// instead of external SOCKS5 proxy (like xray-core's traditional DNS module).
func (m *UpstreamManager) SetDispatcher(d RouteDispatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dispatcher = d
}

// Exchange sends a DNS query to the specified upstream server via UDP or TCP
// and returns the parsed response. This is a concrete implementation of the
// upstream query mechanism.
//
// The method:
//  1. Constructs a dns.Msg from the DnsQuery.
//  2. Adds EDNS0 Client Subnet (ECS) option if client IP is available.
//  3. Sets the DNSSEC OK (DO) bit for DNSSEC-aware upstreams.
//  4. Selects the transport protocol (udp/tcp) from the upstream config.
//  5. If xray-core dispatcher is available, ALWAYS uses it to prevent DNS loop:
//     - Without dispatcher: direct UDP/TCP query hits iptables --dport 53 redirect
//     → gets sent back to DNS module → infinite loop
//     - With dispatcher: xray outbound has sockopt.mark=0x80, iptables sees the
//     0x80 mark and skips redirect (TP_OUT/DNS_MARK chains RETURN on 0x80/0x80)
//  6. If the upstream has a non-empty, non-"direct" ProxyTag without dispatcher,
//     routes through external SOCKS5 proxy.
//  7. Otherwise, sends directly via UDP/TCP.
//  8. Sends the query and measures round-trip time.
//  9. If UDP response is truncated, automatically falls back to TCP.
//  10. On failure, retries once after a short delay.
//  11. Wraps the response into DnsResponse.
func (m *UpstreamManager) Exchange(upstream *UpstreamInstance, query *DnsQuery) (*DnsResponse, error) {
	if upstream == nil {
		return nil, nil
	}

	// Route through proxy if a proxy channel is specified — use dispatcher or SOCKS5.
	if upstream.ProxyTag != "" && upstream.ProxyTag != "direct" {
		if m.dispatcher != nil {
			return m.exchangeViaDispatcher(upstream, query)
		}
		return m.exchangeViaProxy(upstream, query)
	}

	// Direct upstream: use direct UDP/TCP with SO_MARK=0x80 on the socket.
	// This prevents the iptables DNS redirect loop:
	//   - DNS_MARK chain checks 0x80/0x80 → RETURN (skip)
	//   - No 0x40 mark set, so ip rule fwmark 0x40/0xc0 doesn't match
	//   - Packet goes out normally, no redirect to :52353
	return m.exchangeDirect(upstream, query)
}

// exchangeDirect sends a DNS query directly via UDP/TCP with SO_MARK=0x80 set
// on the socket. The 0x80 mark tells iptables to skip redirect (loop protection).
func (m *UpstreamManager) exchangeDirect(upstream *UpstreamInstance, query *DnsQuery) (*DnsResponse, error) {
	if upstream == nil {
		return nil, nil
	}

	// Build the DNS query message.
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(query.Name), uint16(query.QType))
	msg.RecursionDesired = true

	// IMPORTANT: SetEdns0 MUST be called BEFORE AddECSSubnet.
	// SetEdns0 in miekg/dns v1.1.72 always appends a new OPT record without
	// checking if one already exists. Calling AddECSSubnet first then SetEdns0
	// would create TWO OPT records in the query, causing FORMERR.
	msg.SetEdns0(4096, true)
	if query.ClientIP != nil {
		builder := NewResponseBuilder()
		builder.AddECSSubnet(msg, query.ClientIP)
	}

	protocol := upstream.Protocol
	if protocol == "" {
		protocol = "udp"
	}

	log.Printf("[dns upstream] exchange direct: %s %s → %s (%s)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, protocol)

	// Create DNS client with SO_MARK=0x80 on all sockets.
	// The Control function sets the socket mark to 0x80, which iptables
	// DNS_MARK/TP_OUT chains check and RETURN (skip), preventing the loop.
	markFd := func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			_ = setSocketMark(fd) // SO_MARK=36
		})
	}

	client := &dns.Client{
		Net:          protocol,
		Timeout:      5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Dialer: &net.Dialer{
			Timeout: 5 * time.Second,
			Control: markFd,
		},
	}

	// Attempt the exchange with retry logic.
	var resp *dns.Msg
	var rtt time.Duration
	var err error

	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			log.Printf("[dns upstream] retry %d: %s %s → %s", attempt+1,
				dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr)
			time.Sleep(100 * time.Millisecond)
		}

		if protocol == "udp" {
			resp, rtt, err = m.exchangeUDPWithMark(client, msg, upstream.Addr)
		} else {
			resp, rtt, err = client.Exchange(msg, upstream.Addr)
		}

		if err != nil {
			log.Printf("[dns upstream] attempt %d error: %s %s → %s: %v", attempt+1,
				dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, err)
			continue
		}

		// Handle UDP truncation → TCP fallback.
		if resp != nil && resp.Truncated && protocol == "udp" {
			log.Printf("[dns upstream] truncated response, falling back to TCP: %s %s → %s",
				dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr)

			tcpClient := &dns.Client{
				Net:     "tcp",
				Timeout: 5 * time.Second,
				Dialer: &net.Dialer{
					Timeout: 5 * time.Second,
					Control: markFd,
				},
			}
			resp, rtt, err = tcpClient.Exchange(msg, upstream.Addr)
			if err != nil {
				log.Printf("[dns upstream] TCP fallback error (attempt %d): %s %s → %s: %v", attempt+1,
					dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, err)
				continue
			}
		}

		break
	}

	if err != nil {
		return nil, fmt.Errorf("dns upstream: exchange failed: %w", err)
	}

	if resp == nil {
		return nil, nil
	}

	if err := ValidateResponse(resp); err != nil {
		return nil, fmt.Errorf("dns upstream: invalid response: %w", err)
	}
	if err := ValidateQuestionMatch(resp, query.Name, query.QType); err != nil {
		return nil, fmt.Errorf("dns upstream: question mismatch: %w", err)
	}

	var ttl uint32
	if len(resp.Answer) > 0 {
		ttl = resp.Answer[0].Header().Ttl
		for _, rr := range resp.Answer[1:] {
			if rr.Header().Ttl < ttl {
				ttl = rr.Header().Ttl
			}
		}
	}

	dnsResp := &DnsResponse{
		Query:      *query,
		RawMsg:     resp,
		Rcode:      resp.Rcode,
		Answer:     resp.Answer,
		Authority:  resp.Ns,
		Additional: resp.Extra,
		TTL:        ttl,
		Upstream:   upstream.Addr,
		ProxyTag:   upstream.ProxyTag,
		RTT:        rtt,
		Cached:     false,
	}

	log.Printf("[dns upstream] direct response: %s %s → %s (rcode=%d, rtt=%v, answers=%d)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr,
		resp.Rcode, rtt, len(resp.Answer))

	return dnsResp, nil
}

// exchangeUDPWithMark sends a DNS query over UDP with SO_MARK=0x80 set on the socket.
// miekg/dns.Client.Exchange for UDP doesn't use the Dialer.Control function,
// so we create the UDP socket manually to set the socket mark.
func (m *UpstreamManager) exchangeUDPWithMark(client *dns.Client, msg *dns.Msg, addr string) (*dns.Msg, time.Duration, error) {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve udp: %w", err)
	}

	// Create UDP connection with SO_MARK=0x80
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, 0, fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	// Set SO_MARK on the UDP socket
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return nil, 0, fmt.Errorf("get raw conn: %w", err)
	}
	rawConn.Control(func(fd uintptr) {
		_ = setSocketMark(fd)
	})

	// Pack and send the DNS query
	packed, err := msg.Pack()
	if err != nil {
		return nil, 0, fmt.Errorf("dns pack: %w", err)
	}

	start := time.Now()

	if _, err := conn.Write(packed); err != nil {
		return nil, 0, fmt.Errorf("dns write: %w", err)
	}

	// Read response
	respBuf := make([]byte, dns.DefaultMsgSize)
	conn.SetReadDeadline(time.Now().Add(client.ReadTimeout))
	n, err := conn.Read(respBuf)
	rtt := time.Since(start)
	if err != nil {
		return nil, rtt, fmt.Errorf("dns read: %w", err)
	}

	// Unpack response
	resp := new(dns.Msg)
	if err := resp.Unpack(respBuf[:n]); err != nil {
		return nil, rtt, fmt.Errorf("dns unpack: %w", err)
	}

	return resp, rtt, nil
}

// exchangeViaProxy sends a DNS query through a proxy channel.
// It uses the configured ProxyAddrResolver to find the SOCKS5/HTTP proxy
// address for the given proxy tag, then sends the DNS query over TCP
// through that proxy.
//
// This is a simplified implementation that:
//  1. Resolves the proxy tag to a SOCKS5 proxy address.
//  2. Establishes a TCP connection through the SOCKS5 proxy.
//  3. Sends the DNS query over TCP (DNS over TCP is required for proxy).
//  4. Reads and parses the response.
//  5. Falls back to direct connection if no proxy address is available.
func (m *UpstreamManager) exchangeViaProxy(upstream *UpstreamInstance, query *DnsQuery) (*DnsResponse, error) {
	if upstream == nil {
		return nil, nil
	}

	log.Printf("[dns upstream] exchange via proxy: %s %s → %s (proxyTag=%s)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, upstream.ProxyTag)

	// Get the proxy address for the configured tag.
	proxyAddr := m.resolveProxyAddr(upstream.ProxyTag)
	if proxyAddr == "" {
		log.Printf("[dns upstream] no proxy address for tag %q, falling back to direct", upstream.ProxyTag)
		// Fallback: create a temporary upstream with ProxyTag="direct" and retry
		directUpstream := &UpstreamInstance{
			Addr:     upstream.Addr,
			Protocol: "tcp",
			ProxyTag: "direct",
			Client:   m.getClientForProtocol(upstream, "tcp"),
		}
		return m.Exchange(directUpstream, query)
	}

	// Build DNS query message.
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(query.Name), uint16(query.QType))
	msg.RecursionDesired = true

	// IMPORTANT: SetEdns0 MUST be called BEFORE AddECSSubnet.
	// SetEdns0 in miekg/dns v1.1.72 always appends a new OPT record without
	// checking if one already exists. Calling AddECSSubnet first then SetEdns0
	// would create TWO OPT records in the query, causing FORMERR.
	msg.SetEdns0(4096, true)
	if query.ClientIP != nil {
		builder := NewResponseBuilder()
		builder.AddECSSubnet(msg, query.ClientIP)
	}

	// Exchange via SOCKS5 proxy (DNS over TCP).
	start := time.Now()
	resp, err := m.exchangeViaSocks5(proxyAddr, upstream.Addr, msg)
	rtt := time.Since(start)

	if err != nil {
		log.Printf("[dns upstream] proxy exchange error: %s %s → %s via %s: %v",
			dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, proxyAddr, err)

		// Retry once
		time.Sleep(100 * time.Millisecond)
		start = time.Now()
		resp, err = m.exchangeViaSocks5(proxyAddr, upstream.Addr, msg)
		rtt = time.Since(start)

		if err != nil {
			log.Printf("[dns upstream] proxy exchange retry failed: %s %s → %s via %s: %v",
				dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, proxyAddr, err)
			return nil, fmt.Errorf("dns upstream: proxy exchange failed: %w", err)
		}
	}

	if resp == nil {
		log.Printf("[dns upstream] empty proxy response: %s %s → %s",
			dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr)
		return nil, nil
	}

	// Validate response.
	if err := ValidateResponse(resp); err != nil {
		return nil, fmt.Errorf("dns upstream: invalid proxy response: %w", err)
	}
	if err := ValidateQuestionMatch(resp, query.Name, query.QType); err != nil {
		return nil, fmt.Errorf("dns upstream: proxy question mismatch: %w", err)
	}

	// Calculate TTL.
	var ttl uint32
	if len(resp.Answer) > 0 {
		ttl = resp.Answer[0].Header().Ttl
		for _, rr := range resp.Answer[1:] {
			if rr.Header().Ttl < ttl {
				ttl = rr.Header().Ttl
			}
		}
	}

	dnsResp := &DnsResponse{
		Query:      *query,
		RawMsg:     resp,
		Rcode:      resp.Rcode,
		Answer:     resp.Answer,
		Authority:  resp.Ns,
		Additional: resp.Extra,
		TTL:        ttl,
		Upstream:   upstream.Addr,
		ProxyTag:   upstream.ProxyTag,
		RTT:        rtt,
		Cached:     false,
	}

	log.Printf("[dns upstream] proxy response: %s %s → %s via %s (rcode=%d, rtt=%v, answers=%d)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, proxyAddr,
		resp.Rcode, rtt, len(resp.Answer))

	return dnsResp, nil
}

// exchangeViaSocks5 sends a DNS query through a SOCKS5 proxy using TCP.
// It implements the SOCKS5 protocol to establish a connection through
// the proxy, then sends the DNS query over TCP.
func (m *UpstreamManager) exchangeViaSocks5(proxyAddr, targetAddr string, msg *dns.Msg) (*dns.Msg, error) {
	// Establish TCP connection to the SOCKS5 proxy.
	conn, err := net.DialTimeout("tcp", proxyAddr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("socks5 dial: %w", err)
	}
	defer conn.Close()

	// SOCKS5 handshake: send greeting (no auth)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return nil, fmt.Errorf("socks5 greeting: %w", err)
	}

	// Read server response
	buf := make([]byte, 2)
	if _, err := conn.Read(buf); err != nil {
		return nil, fmt.Errorf("socks5 greeting response: %w", err)
	}
	if buf[0] != 0x05 {
		return nil, fmt.Errorf("socks5: invalid version: %d", buf[0])
	}

	// Parse target host and port for SOCKS5 request.
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		host = targetAddr
		portStr = "53"
	}
	port := 53
	if p, err := net.LookupPort("tcp", portStr); err == nil {
		port = p
	}

	// Build SOCKS5 connect request (ATYP: 0x01 for IPv4, 0x03 for domain)
	var req []byte
	ip := net.ParseIP(host)
	if ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			// IPv4
			req = []byte{0x05, 0x01, 0x00, 0x01}
			req = append(req, ip4...)
		} else {
			// IPv6
			req = []byte{0x05, 0x01, 0x00, 0x04}
			req = append(req, ip.To16()...)
		}
	} else {
		// Domain name
		req = []byte{0x05, 0x01, 0x00, 0x03}
		req = append(req, byte(len(host)))
		req = append(req, []byte(host)...)
	}
	req = append(req, byte(port>>8), byte(port))

	if _, err := conn.Write(req); err != nil {
		return nil, fmt.Errorf("socks5 connect request: %w", err)
	}

	// Read SOCKS5 response
	respBuf := make([]byte, 256)
	n, err := conn.Read(respBuf)
	if err != nil {
		return nil, fmt.Errorf("socks5 connect response: %w", err)
	}
	if n < 2 || respBuf[0] != 0x05 || respBuf[1] != 0x00 {
		return nil, fmt.Errorf("socks5 connect failed: version=%d status=%d", respBuf[0], respBuf[1])
	}

	// DNS over TCP: prepend length prefix.
	packed, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("dns pack: %w", err)
	}

	tcpMsg := make([]byte, 2+len(packed))
	tcpMsg[0] = byte(len(packed) >> 8)
	tcpMsg[1] = byte(len(packed))
	copy(tcpMsg[2:], packed)

	// Send DNS query.
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(tcpMsg); err != nil {
		return nil, fmt.Errorf("dns write via socks5: %w", err)
	}

	// Read DNS response (TCP length-prefixed).
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read the 2-byte length prefix.
	if _, err := conn.Read(tcpMsg[:2]); err != nil {
		return nil, fmt.Errorf("dns read length via socks5: %w", err)
	}
	respLen := int(tcpMsg[0])<<8 | int(tcpMsg[1])
	if respLen < 0 || respLen > 65535 {
		return nil, fmt.Errorf("dns invalid response length: %d", respLen)
	}

	// Read the full DNS response.
	respData := make([]byte, respLen)
	totalRead := 0
	for totalRead < respLen {
		n, err := conn.Read(respData[totalRead:])
		if err != nil {
			return nil, fmt.Errorf("dns read data via socks5: %w", err)
		}
		totalRead += n
	}

	// Unpack the DNS response.
	dnsResp := new(dns.Msg)
	if err := dnsResp.Unpack(respData); err != nil {
		return nil, fmt.Errorf("dns unpack: %w", err)
	}

	return dnsResp, nil
}

// resolveProxyAddr resolves a proxy tag to a SOCKS5 proxy address.
// It first checks the configured ProxyAddrResolver, then falls back
// to known defaults for well-known tags.
func (m *UpstreamManager) resolveProxyAddr(proxyTag string) string {
	m.mu.RLock()
	resolver := m.proxyAddrResolver
	m.mu.RUnlock()

	if resolver != nil {
		if addr := resolver(proxyTag); addr != "" {
			return addr
		}
	}

	// Fallback: check well-known default proxy addresses.
	// These can be overridden by the resolver.
	switch proxyTag {
	case "proxy":
		return "127.0.0.1:1080" // Common SOCKS5 proxy port
	}

	return ""
}

// exchangeWithProtocol sends a DNS query using the specified protocol and
// returns the response, round-trip time, and any error.
func (m *UpstreamManager) exchangeWithProtocol(upstream *UpstreamInstance, msg *dns.Msg, protocol string) (*dns.Msg, time.Duration, error) {
	// Create or reuse the DNS client for the specified protocol.
	client := m.getClientForProtocol(upstream, protocol)
	if client == nil {
		return nil, 0, fmt.Errorf("failed to create DNS client for %s", protocol)
	}

	// Use a context with timeout for the exchange.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, rtt, err := client.ExchangeContext(ctx, msg, upstream.Addr)
	if err != nil {
		return resp, rtt, err
	}

	return resp, rtt, nil
}

// getClientForProtocol returns a DNS client configured for the given protocol.
// It reuses the upstream's client if the protocol matches; otherwise, it creates
// a new client for the requested protocol.
func (m *UpstreamManager) getClientForProtocol(upstream *UpstreamInstance, protocol string) *dns.Client {
	// If the upstream already has a client with the matching protocol, reuse it.
	if upstream.Client != nil && upstream.Client.Net == protocol {
		return upstream.Client
	}

	// Create a new client for the requested protocol.
	return &dns.Client{
		Net:          protocol,
		Timeout:      5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

// ExchangeRaw sends a raw *dns.Msg to the upstream and returns the raw response.
// This is used for cases where the caller needs low-level access to the DNS
// message without DnsQuery/DnsResponse wrapping.
func (m *UpstreamManager) ExchangeRaw(upstream *UpstreamInstance, msg *dns.Msg) (*dns.Msg, time.Duration, error) {
	if upstream == nil {
		return nil, 0, fmt.Errorf("dns upstream: nil upstream")
	}
	if msg == nil {
		return nil, 0, fmt.Errorf("dns upstream: nil message")
	}

	protocol := upstream.Protocol
	if protocol == "" {
		protocol = "udp"
	}

	client := m.getClientForProtocol(upstream, protocol)

	resp, rtt, err := client.Exchange(msg, upstream.Addr)
	if err != nil {
		return nil, rtt, fmt.Errorf("dns upstream: exchange raw: %w", err)
	}

	// Handle UDP truncation → TCP fallback.
	if resp != nil && resp.Truncated && protocol == "udp" {
		log.Printf("[dns upstream] truncated raw response, falling back to TCP: %s",
			upstream.Addr)

		tcpClient := m.getClientForProtocol(upstream, "tcp")
		resp, rtt, err = tcpClient.Exchange(msg, upstream.Addr)
		if err != nil {
			return nil, rtt, fmt.Errorf("dns upstream: exchange raw tcp fallback: %w", err)
		}
	}

	return resp, rtt, nil
}

// exchangeViaDispatcher sends a DNS query through xray-core's internal routing
// dispatcher. It uses the RouteDispatcher to create a connection through
// xray-core's routing engine (like xray-core's traditional DNS module does),
// then performs DNS-over-TCP on that connection.
//
// This replaces the external SOCKS5 proxy approach with v2raya-core's built-in
// routing logic, supporting any outbound type (VMESS, VLESS, Trojan, etc.).
func (m *UpstreamManager) exchangeViaDispatcher(upstream *UpstreamInstance, query *DnsQuery) (*DnsResponse, error) {
	if upstream == nil || m.dispatcher == nil {
		return nil, nil
	}

	log.Printf("[dns upstream] exchange via xray routing: %s %s → %s (proxyTag=%s)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, upstream.ProxyTag)

	// Build DNS query message.
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(query.Name), uint16(query.QType))
	msg.RecursionDesired = true

	// IMPORTANT: SetEdns0 MUST be called BEFORE AddECSSubnet.
	// SetEdns0 in miekg/dns v1.1.72 always appends a new OPT record without
	// checking if one already exists. Calling AddECSSubnet first then SetEdns0
	// would create TWO OPT records in the query, causing FORMERR.
	msg.SetEdns0(4096, true)
	if query.ClientIP != nil {
		builder := NewResponseBuilder()
		builder.AddECSSubnet(msg, query.ClientIP)
	}

	// Pack DNS query for TCP transport.
	packed, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("dns pack: %w", err)
	}

	// Dispatch through xray-core's routing via TCP.
	// The proxyTag is passed to the dispatcher, which sets session.ContextWithInbound
	// internally (like xray-core's DNS module), so xray's routing engine
	// determines the outbound based on routing rules.
	start := time.Now()
	conn, err := m.dispatcher.Dispatch(context.Background(), "tcp", upstream.Addr, upstream.ProxyTag)
	if err != nil {
		log.Printf("[dns upstream] dispatcher error: %s %s → %s: %v",
			dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr, err)

		// Retry once
		time.Sleep(100 * time.Millisecond)
		conn, err = m.dispatcher.Dispatch(context.Background(), "tcp", upstream.Addr, upstream.ProxyTag)
		if err != nil {
			return nil, fmt.Errorf("dns upstream: dispatcher retry failed: %w", err)
		}
	}
	defer conn.Close()

	// DNS over TCP: prepend 2-byte length prefix.
	tcpMsg := make([]byte, 2+len(packed))
	tcpMsg[0] = byte(len(packed) >> 8)
	tcpMsg[1] = byte(len(packed))
	copy(tcpMsg[2:], packed)

	// Send DNS query.
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(tcpMsg); err != nil {
		return nil, fmt.Errorf("dns write via dispatcher: %w", err)
	}

	// Read 2-byte length prefix.
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Read(tcpMsg[:2]); err != nil {
		return nil, fmt.Errorf("dns read length via dispatcher: %w", err)
	}
	respLen := int(tcpMsg[0])<<8 | int(tcpMsg[1])
	if respLen < 0 || respLen > 65535 {
		return nil, fmt.Errorf("dns invalid response length: %d", respLen)
	}

	// Read full DNS response.
	respData := make([]byte, respLen)
	totalRead := 0
	for totalRead < respLen {
		n, err := conn.Read(respData[totalRead:])
		if err != nil {
			return nil, fmt.Errorf("dns read data via dispatcher: %w", err)
		}
		totalRead += n
	}

	rtt := time.Since(start)

	// Unpack DNS response.
	dnsResp := new(dns.Msg)
	if err := dnsResp.Unpack(respData); err != nil {
		return nil, fmt.Errorf("dns unpack: %w", err)
	}

	// Validate response.
	if err := ValidateResponse(dnsResp); err != nil {
		return nil, fmt.Errorf("dns upstream: invalid dispatcher response: %w", err)
	}
	if err := ValidateQuestionMatch(dnsResp, query.Name, query.QType); err != nil {
		return nil, fmt.Errorf("dns upstream: dispatcher question mismatch: %w", err)
	}

	// Calculate TTL.
	var ttl uint32
	if len(dnsResp.Answer) > 0 {
		ttl = dnsResp.Answer[0].Header().Ttl
		for _, rr := range dnsResp.Answer[1:] {
			if rr.Header().Ttl < ttl {
				ttl = rr.Header().Ttl
			}
		}
	}

	resp := &DnsResponse{
		Query:      *query,
		RawMsg:     dnsResp,
		Rcode:      dnsResp.Rcode,
		Answer:     dnsResp.Answer,
		Authority:  dnsResp.Ns,
		Additional: dnsResp.Extra,
		TTL:        ttl,
		Upstream:   upstream.Addr,
		ProxyTag:   upstream.ProxyTag,
		RTT:        rtt,
		Cached:     false,
	}

	log.Printf("[dns upstream] dispatcher response: %s %s → %s (rcode=%d, rtt=%v, answers=%d)",
		dns.Type(uint16(query.QType)).String(), query.Name, upstream.Addr,
		dnsResp.Rcode, rtt, len(dnsResp.Answer))

	return resp, nil
}
