package tun

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"net"
	"net/netip"
	"strings"
	"time"

	D "github.com/miekg/dns"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	DNSTimeout = 10 * time.Second
	FakeTTL    = 5 * time.Minute

	TunDNSListenPort = 6053

	FixedPacketSize = 16384
)

const (
	dnsDirect = iota
	dnsForward
	dnsFake4
	dnsFake6
)

var (
	fakePrefix4 = netip.MustParsePrefix("198.18.0.0/15")
	fakePrefix6 = netip.MustParsePrefix("fc00::/18")

	defaultDNSServer = M.ParseSocksaddrHostPort("127.0.0.1", TunDNSListenPort)
)

type DNS struct {
	dialer       N.Dialer
	forward      N.Dialer
	addrs        []M.Socksaddr
	forceProxy   bool
	whitelist    Matcher
	servers      []M.Socksaddr
	cache        *cache[netip.Addr, string]
	currentIP4   netip.Addr
	currentIP6   netip.Addr
	fakeCache    *cache[netip.Addr, string]
	domainCache4 *cache[string, netip.Addr]
	domainCache6 *cache[string, netip.Addr]
	useFakeIP    bool
}

func NewDNS(dialer, forward N.Dialer, forceProxy bool, addrs ...M.Socksaddr) *DNS {
	return &DNS{
		dialer:       dialer,
		forward:      forward,
		addrs:        addrs,
		forceProxy:   forceProxy,
		cache:        newCache[netip.Addr, string](),
		currentIP4:   fakePrefix4.Addr().Next(),
		currentIP6:   fakePrefix6.Addr().Next(),
		fakeCache:    newCache[netip.Addr, string](),
		domainCache4: newCache[string, netip.Addr](),
		domainCache6: newCache[string, netip.Addr](),
		useFakeIP:    true, // 默认启用 FakeIP
	}
}

func (d *DNS) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	d.fakeCache.Check()
	if !d.matchAddr(metadata.Destination) {
		return continueHandler
	}
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		for {
			var queryLength uint16
			err := binary.Read(conn, binary.BigEndian, &queryLength)
			if err != nil {
				cancel(err)
				return
			}
			if queryLength == 0 {
				cancel(errors.New("format error"))
				return
			}
			buffer := buf.NewSize(int(queryLength))
			defer buffer.Release()
			_, err = buffer.ReadFullFrom(conn, int(queryLength))
			if err != nil {
				cancel(err)
				return
			}
			var message D.Msg
			err = message.Unpack(buffer.Bytes())
			if err != nil {
				cancel(err)
				return
			}
			go func() {
				response, err := d.Exchange(ctx, &message)
				if err != nil {
					cancel(err)
					return
				}
				responseBuffer := buf.NewPacket()
				defer responseBuffer.Release()
				responseBuffer.Resize(2, 0)
				n, err := response.PackBuffer(responseBuffer.FreeBytes())
				if err != nil {
					cancel(err)
					return
				}
				responseBuffer.Truncate(len(n))
				binary.BigEndian.PutUint16(responseBuffer.ExtendHeader(2), uint16(len(n)))
				_, err = conn.Write(responseBuffer.Bytes())
				if err != nil {
					cancel(err)
				}
			}()
		}
	}()
	<-ctx.Done()
	cancel(nil)
	conn.Close()
	return ctx.Err()
}

func (d *DNS) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata M.Metadata) error {
	d.fakeCache.Check()
	if !d.matchAddr(metadata.Destination) {
		return continueHandler
	}
	var reader N.PacketReader = conn
	var counters []N.CountFunc
	var cachedPackets []*N.PacketBuffer
	for {
		reader, counters = N.UnwrapCountPacketReader(reader, counters)
		if cachedReader, isCached := reader.(N.CachedPacketReader); isCached {
			packet := cachedReader.ReadCachedPacket()
			if packet != nil {
				cachedPackets = append(cachedPackets, packet)
				continue
			}
		}
		if readWaiter, created := bufio.CreatePacketReadWaiter(reader); created {
			return d.newPacketConnection(ctx, conn, readWaiter, counters, cachedPackets, metadata.Destination)
		}
		break
	}
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		for {
			var message D.Msg
			var destination M.Socksaddr
			var err error
			if len(cachedPackets) > 0 {
				packet := cachedPackets[0]
				cachedPackets = cachedPackets[1:]
				for _, counter := range counters {
					counter(int64(packet.Buffer.Len()))
				}
				err = message.Unpack(packet.Buffer.Bytes())
				packet.Buffer.Release()
				if err != nil {
					cancel(err)
					return
				}
				destination = packet.Destination
			} else {
				timeout := time.AfterFunc(DNSTimeout, func() {
					cancel(context.DeadlineExceeded)
				})
				buffer := buf.NewPacket()
				destination, err = conn.ReadPacket(buffer)
				if err != nil {
					buffer.Release()
					cancel(err)
					return
				}
				for _, counter := range counters {
					counter(int64(buffer.Len()))
				}
				err = message.Unpack(buffer.Bytes())
				buffer.Release()
				if err != nil {
					cancel(err)
					return
				}
				timeout.Stop()
			}
			go func() {
				response, err := d.Exchange(ctx, &message)
				if err != nil {
					cancel(err)
					return
				}
				responseBuffer := buf.NewPacket()
				n, err := response.PackBuffer(responseBuffer.FreeBytes())
				if err != nil {
					cancel(err)
					responseBuffer.Release()
					return
				}
				responseBuffer.Truncate(len(n))
				err = conn.WritePacket(responseBuffer, destination)
				if err != nil {
					cancel(err)
				}
			}()
		}
	}()
	<-ctx.Done()
	cancel(nil)
	conn.Close()
	return ctx.Err()
}

func (d *DNS) matchAddr(addr M.Socksaddr) bool {
	if addr.Port == 53 {
		return true
	}
	for _, item := range d.addrs {
		if addr == item {
			return true
		}
	}
	return false
}

func (d *DNS) Exchange(ctx context.Context, msg *D.Msg) (*D.Msg, error) {
	if len(msg.Question) != 1 {
		return d.newResponse(msg, D.RcodeFormatError), nil
	}
	mode := dnsDirect
	if len(d.servers) == 0 {
		mode = dnsForward
	}
	question := msg.Question[0]
	domain := strings.TrimSuffix(question.Name, ".")
	if d.useFakeIP && !d.whitelist.Match(domain) {
		switch question.Qtype {
		case D.TypeA:
			mode = dnsFake4
		case D.TypeAAAA:
			return d.newResponse(msg, D.RcodeSuccess), nil
		case D.TypeMX, D.TypeHTTPS:
			return d.newResponse(msg, D.RcodeSuccess), nil
		}
	}
	var dialer N.Dialer
	server := defaultDNSServer
	switch mode {
	case dnsDirect:
		if d.forceProxy {
			dialer = d.forward
		} else {
			dialer = d.dialer
		}
		server = d.getServer()
	case dnsForward:
		// In TUN mode, d.forward is nil. Fall back to using dokodemo-door for DNS forwarding.
		if d.forward != nil {
			dialer = d.forward
		} else {
			dialer = d.dialer
			server = defaultDNSServer // Use dokodemo-door at 127.0.0.1:6053
		}
	case dnsFake4:
		addr, ok := d.getAvailableIP4(domain)
		if ok {
			return d.newResponse(msg, D.RcodeSuccess, addr), nil
		}
		// FakeIP allocation failed (rare), fall back to querying DNS
		if d.forward != nil {
			dialer = d.forward
		} else {
			dialer = d.dialer
			server = defaultDNSServer
		}
	case dnsFake6:
		if addr, ok := d.getAvailableIP6(domain); ok {
			return d.newResponse(msg, D.RcodeSuccess, addr), nil
		}
		// FakeIP allocation failed (rare), fall back to querying DNS
		if d.forward != nil {
			dialer = d.forward
		} else {
			dialer = d.dialer
			server = defaultDNSServer
		}
	}
	if dialer != nil {
		useForward := dialer == d.forward
		preferTCP := false
		var usedTCP bool
		resp, err := func() (*D.Msg, error) {
			ctxDial, cancel := context.WithTimeout(ctx, DNSTimeout)
			defer cancel()

			buffer := make([]byte, 2048)
			data, err := msg.PackBuffer(buffer)
			if err != nil {
				return nil, err
			}

			// First try UDP. For dokodemo-door (port 6053), this is a direct UDP connection.
			// The dialer is SystemDialer in normal cases, connecting to the local dokodemo-door listener.
			serverConn, err := dialer.ListenPacket(ctxDial, server)
			if err == nil {
				defer serverConn.Close()
				serverConn.SetDeadline(time.Now().Add(DNSTimeout))
				if _, err = serverConn.WriteTo(data, server.UDPAddr()); err == nil {
					n, _, rErr := serverConn.ReadFrom(buffer)
					if rErr == nil {
						var resp D.Msg
						unpackErr := resp.Unpack(buffer[:n])
						if unpackErr == nil {
							return &resp, nil
						}
						err = unpackErr
					} else {
						err = rErr
					}
				}
			}

			// Fallback to TCP when UDP fails (e.g., packet too large or server requires TCP).
			if err != nil && preferTCP {
				usedTCP = true
				return exchangeTCP(ctxDial, dialer, server, msg)
			}

			// If UDP failed and we do not prefer TCP, return the UDP error.
			if err != nil {
				return nil, err
			}

			// Should not reach here; UDP success already returned.
			return nil, nil
		}()
		if err != nil {
			log.Warn("[TUN-DNS] query=%s qtype=%d via=%s server=%s err=%v", domain, question.Qtype, dialLabel(useForward, usedTCP), server.String(), err)
			return d.newResponse(msg, D.RcodeServerFailure), nil
		}
		log.Trace("[TUN-DNS] query=%s qtype=%d via=%s server=%s rcode=%d", domain, question.Qtype, dialLabel(useForward, usedTCP), server.String(), resp.Rcode)
		return resp, nil
	}
	return d.newResponse(msg, D.RcodeRefused), nil
}

func exchangeTCP(ctx context.Context, dialer N.Dialer, server M.Socksaddr, msg *D.Msg) (*D.Msg, error) {
	buffer := make([]byte, 4096)
	data, err := msg.PackBuffer(buffer)
	if err != nil {
		return nil, err
	}
	conn, err := dialer.DialContext(ctx, N.NetworkTCP, server)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(DNSTimeout))
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(data)))
	if _, err = conn.Write(length); err != nil {
		return nil, err
	}
	if _, err = conn.Write(data); err != nil {
		return nil, err
	}
	if _, err = io.ReadFull(conn, length); err != nil {
		return nil, err
	}
	respLen := int(binary.BigEndian.Uint16(length))
	respBuf := make([]byte, respLen)
	if _, err = io.ReadFull(conn, respBuf); err != nil {
		return nil, err
	}
	var resp D.Msg
	if err = resp.Unpack(respBuf); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (d *DNS) newPacketConnection(ctx context.Context, conn N.PacketConn, readWaiter N.PacketReadWaiter, readCounters []N.CountFunc, cached []*N.PacketBuffer, metadata M.Socksaddr) error {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		readWaiter.InitializeReadWaiter(N.ReadWaitOptions{})
		defer readWaiter.InitializeReadWaiter(N.ReadWaitOptions{})
		for {
			var message D.Msg
			var destination M.Socksaddr
			var err error
			if len(cached) > 0 {
				packet := cached[0]
				cached = cached[1:]
				for _, counter := range readCounters {
					counter(int64(packet.Buffer.Len()))
				}
				err = message.Unpack(packet.Buffer.Bytes())
				packet.Buffer.Release()
				if err != nil {
					cancel(err)
					return
				}
				destination = packet.Destination
			} else {
				timeout := time.AfterFunc(DNSTimeout, func() {
					cancel(context.DeadlineExceeded)
				})
				buffer, dest, rErr := readWaiter.WaitReadPacket()
				if rErr != nil {
					if buffer != nil {
						buffer.Release()
					}
					cancel(rErr)
					return
				}
				destination = dest
				for _, counter := range readCounters {
					counter(int64(buffer.Len()))
				}
				err = message.Unpack(buffer.Bytes())
				buffer.Release()
				if err != nil {
					cancel(err)
					return
				}
				timeout.Stop()
			}
			go func() {
				response, err := d.Exchange(ctx, &message)
				if err != nil {
					cancel(err)
					return
				}
				responseBuffer := buf.NewPacket()
				n, err := response.PackBuffer(responseBuffer.FreeBytes())
				if err != nil {
					cancel(err)
					responseBuffer.Release()
					return
				}
				responseBuffer.Truncate(len(n))
				err = conn.WritePacket(responseBuffer, destination)
				if err != nil {
					cancel(err)
				}
			}()
		}
	}()
	<-ctx.Done()
	cancel(nil)
	conn.Close()
	return ctx.Err()
}

func (d *DNS) newResponse(msg *D.Msg, code int, answer ...netip.Addr) *D.Msg {
	resp := D.Msg{
		MsgHdr: D.MsgHdr{
			Id:       msg.Id,
			Response: true,
			Rcode:    code,
		},
		Question: msg.Question,
	}
	for _, addr := range answer {
		var rr D.RR
		if addr.Is4() {
			rr = &D.A{
				Hdr: D.RR_Header{
					Name:     msg.Question[0].Name,
					Rrtype:   D.TypeA,
					Class:    D.ClassINET,
					Ttl:      uint32(FakeTTL / time.Second),
					Rdlength: 4,
				},
				A: addr.AsSlice(),
			}
		} else if addr.Is6() {
			rr = &D.AAAA{
				Hdr: D.RR_Header{
					Name:     msg.Question[0].Name,
					Rrtype:   D.TypeAAAA,
					Class:    D.ClassINET,
					Ttl:      uint32(FakeTTL / time.Second),
					Rdlength: 16,
				},
				AAAA: addr.AsSlice(),
			}
		}
		resp.Answer = append(resp.Answer, rr)
	}
	return &resp
}

func (d *DNS) getAvailableIP4(domain string) (netip.Addr, bool) {
	d.domainCache4.Lock()
	defer d.domainCache4.Unlock()
	d.domainCache4.UnsafeCheck()
	addr, ok := d.domainCache4.UnsafeLoad(domain)
	if ok {
		d.fakeCache.Store(addr, domain, FakeTTL)
		d.domainCache4.UnsafeStore(domain, addr, FakeTTL)
		return addr, true
	}
	begin := d.currentIP4
	for {
		addr = d.currentIP4.Next()
		if !fakePrefix4.Contains(addr) {
			addr = fakePrefix4.Addr().Next().Next()
		}
		d.currentIP4 = addr
		if !d.fakeCache.Contains(addr) {
			d.fakeCache.Store(addr, domain, FakeTTL)
			d.domainCache4.UnsafeStore(domain, addr, FakeTTL)
			return addr, true
		} else if addr == begin {
			break
		}
	}
	return addr, false
}

func (d *DNS) getAvailableIP6(domain string) (netip.Addr, bool) {
	d.domainCache6.Lock()
	defer d.domainCache6.Unlock()
	d.domainCache6.UnsafeCheck()
	addr, ok := d.domainCache6.UnsafeLoad(domain)
	if ok {
		d.fakeCache.Store(addr, domain, FakeTTL)
		d.domainCache6.UnsafeStore(domain, addr, FakeTTL)
		return addr, true
	}
	begin := d.currentIP6
	for {
		addr = d.currentIP6.Next()
		if !fakePrefix6.Contains(addr) {
			addr = fakePrefix6.Addr().Next().Next()
		}
		d.currentIP6 = addr
		if !d.fakeCache.Contains(addr) {
			d.fakeCache.Store(addr, domain, FakeTTL)
			d.domainCache6.UnsafeStore(domain, addr, FakeTTL)
			return addr, true
		} else if addr == begin {
			break
		}
	}
	return addr, false
}

func (t *DNS) getServer() M.Socksaddr {
	if len(t.servers) != 1 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(t.servers))))
		if err == nil {
			return t.servers[n.Uint64()]
		}
	}
	return t.servers[0]
}

func dialLabel(useForward bool, useTCP bool) string {
	switch {
	case useForward && useTCP:
		return "socks5-tcp"
	case useForward:
		return "socks5-udp"
	case useTCP:
		return "dokodemo-tcp"
	default:
		return "dokodemo-udp"
	}
}
