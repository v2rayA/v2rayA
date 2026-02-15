package tun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"runtime"
	"slices"
	"sync"

	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/control"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/protocol/socks"
	"github.com/v2fly/v2ray-core/v5/common/strmatcher"
	"github.com/v2rayA/v2ray-lib/router/routercommon"
	"github.com/v2rayA/v2rayA/core/dns"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"google.golang.org/protobuf/proto"
)

const (
	dnsAddr  = "172.19.0.2"
	dnsAddr6 = "fdfe:dcba:9876::2"
)

var (
	prefix4 = netip.MustParsePrefix("172.19.0.1/30")
	prefix6 = netip.MustParsePrefix("fdfe:dcba:9876::1/126")
	route4  = netip.MustParsePrefix("0.0.0.0/0")
	route6  = netip.MustParsePrefix("::/0")

	defaultLogger = logger.NOP()

	continueHandler = errors.New("continue handler")
)

type singTun struct {
	mu           sync.Mutex
	dialer       N.Dialer
	forward      N.Dialer
	cancel       context.CancelFunc
	closer       io.Closer
	waiter       *gvisorWaiter
	dns          *DNS
	backupDNS    map[string][]string
	whitelist    []netip.Addr
	excludeAddrs []netip.Prefix // Addresses to exclude from TUN routing
	useIPv6      bool
	strictRoute  bool
}

// isReservedAddress checks if an IP address belongs to reserved address ranges
// that should not be proxied (loopback, private, link-local, multicast, etc.)
func isReservedAddress(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}

	// IPv4 reserved ranges
	if addr.Is4() {
		// 127.0.0.0/8 - Loopback
		if addr.AsSlice()[0] == 127 {
			return true
		}
		// 10.0.0.0/8 - Private network Class A
		if addr.AsSlice()[0] == 10 {
			return true
		}
		// 172.16.0.0/12 - Private network Class B
		if addr.AsSlice()[0] == 172 && (addr.AsSlice()[1]&0xF0) == 16 {
			return true
		}
		// 192.168.0.0/16 - Private network Class C
		if addr.AsSlice()[0] == 192 && addr.AsSlice()[1] == 168 {
			return true
		}
		// 169.254.0.0/16 - Link-local
		if addr.AsSlice()[0] == 169 && addr.AsSlice()[1] == 254 {
			return true
		}
		// 224.0.0.0/4 - Multicast
		if (addr.AsSlice()[0] & 0xF0) == 224 {
			return true
		}
		// 240.0.0.0/4 - Reserved
		if (addr.AsSlice()[0] & 0xF0) == 240 {
			return true
		}
		// 0.0.0.0/8 - Current network
		if addr.AsSlice()[0] == 0 {
			return true
		}
	}

	// IPv6 reserved ranges
	if addr.Is6() {
		// ::1/128 - Loopback
		if addr.IsLoopback() {
			return true
		}
		// fe80::/10 - Link-local unicast
		if addr.IsLinkLocalUnicast() {
			return true
		}
		// fc00::/7 - Unique local address (ULA)
		if (addr.AsSlice()[0] & 0xFE) == 0xFC {
			return true
		}
		// ff00::/8 - Multicast
		if addr.IsMulticast() {
			return true
		}
		// ::/128 - Unspecified address
		if addr.IsUnspecified() {
			return true
		}
	}

	return false
}

func filterTunDNSServers(servers []netip.AddrPort) []netip.AddrPort {
	dnsAddrIP, _ := netip.ParseAddr(dnsAddr)
	dnsAddrIPv6, _ := netip.ParseAddr(dnsAddr6)
	filtered := make([]netip.AddrPort, 0, len(servers))
	for _, server := range servers {
		addr := server.Addr()
		if !addr.IsValid() {
			continue
		}
		if addr.IsLoopback() || addr.IsUnspecified() {
			continue
		}
		if dnsAddrIP.IsValid() && addr == dnsAddrIP {
			continue
		}
		if dnsAddrIPv6.IsValid() && addr == dnsAddrIPv6 {
			continue
		}
		filtered = append(filtered, server)
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func NewSingTun() Tun {
	dialer := N.SystemDialer
	client := socks.NewClient(dialer, M.ParseSocksaddrHostPort("127.0.0.1", 52345), socks.Version5, "", "")
	log.Info("[TUN] Initialized SOCKS5 client to 127.0.0.1:52345")
	return &singTun{
		dialer:      dialer,
		forward:     client,
		strictRoute: false,
		// DNS is sent to local dokodemo-door listener instead of SOCKS
		dns: NewDNS(dialer, nil, false, M.ParseSocksaddrHostPort(dnsAddr, 53), M.ParseSocksaddrHostPort(dnsAddr6, 53)),
	}
}

func (t *singTun) Start(stack Stack) error {
	var failedCloser Closer
	defer func() {
		failedCloser.Close()
	}()

	t.Close()
	networkUpdateMonitor, err := tun.NewNetworkUpdateMonitor(defaultLogger)
	if err != nil {
		return err
	}
	failedCloser = append(failedCloser, networkUpdateMonitor)
	interfaceMonitor, err := tun.NewDefaultInterfaceMonitor(networkUpdateMonitor, defaultLogger, tun.DefaultInterfaceMonitorOptions{})
	if err != nil {
		return err
	}
	failedCloser = append(failedCloser, interfaceMonitor)
	// Separate excluded addresses by IP version for TUN options
	var inet4Exclude, inet6Exclude []netip.Prefix

	// Always exclude 127.0.0.0/8 (loopback) to prevent capturing local proxy traffic
	loopback4 := netip.MustParsePrefix("127.0.0.0/8")
	inet4Exclude = append(inet4Exclude, loopback4)
	log.Info("[TUN] Excluding loopback: 127.0.0.0/8")

	for _, prefix := range t.excludeAddrs {
		if prefix.Addr().Is4() {
			inet4Exclude = append(inet4Exclude, prefix)
			log.Info("[TUN] Excluding IPv4: %s", prefix.String())
		} else {
			inet6Exclude = append(inet6Exclude, prefix)
			log.Info("[TUN] Excluding IPv6: %s", prefix.String())
		}
	}

	autoRoute := true
	strictRoute := t.strictRoute

	// On Windows, disable sing-tun's AutoRoute as it adds metric=0 route
	// We will add manual default route with proper metric instead
	if runtime.GOOS == "windows" {
		autoRoute = false
		log.Info("[TUN] Windows: Disabling AutoRoute, will use manual routing")
	}

	log.Info("[TUN] Starting with StrictRoute=%t, AutoRoute=%t", strictRoute, autoRoute)
	log.Info("[TUN] Total exclusions: %d IPv4, %d IPv6", len(inet4Exclude), len(inet6Exclude))
	if runtime.GOOS == "windows" && strictRoute {
		log.Warn("[TUN] Windows: StrictRoute may only allow main process traffic!")
	}

	// Pre-install exclusion routes on platforms without fwmark support (e.g. Windows/macOS)
	if len(t.excludeAddrs) > 0 {
		if err := SetupExcludeRoutes(t.excludeAddrs); err != nil {
			log.Warn("[TUN] Failed to pre-install exclude routes: %v", err)
		}
	}

	tunOptions := tun.Options{
		Name:                     tun.CalculateInterfaceName(""),
		MTU:                      9000,
		Inet4Address:             []netip.Prefix{prefix4},
		Inet4RouteAddress:        []netip.Prefix{route4},
		Inet4RouteExcludeAddress: inet4Exclude, // Exclude loopback + server IPs
		AutoRoute:                autoRoute,
		StrictRoute:              strictRoute,
		InterfaceMonitor:         interfaceMonitor,
	}
	// Enable IPv6 if requested
	if t.useIPv6 {
		tunOptions.Inet6Address = []netip.Prefix{prefix6}
		tunOptions.Inet6RouteAddress = []netip.Prefix{route6}
		// Exclude IPv6 loopback (::1/128)
		loopback6 := netip.MustParsePrefix("::1/128")
		inet6Exclude = append([]netip.Prefix{loopback6}, inet6Exclude...)
		tunOptions.Inet6RouteExcludeAddress = inet6Exclude
		log.Info("[TUN] Excluding IPv6 loopback: ::1/128")
	}
	tunInterface, err := tun.New(tunOptions)
	if err != nil {
		return err
	}
	failedCloser = append(failedCloser, tunInterface)

	// Setup policy routing rules to exclude fwmark 0x80 traffic (v2ray/xray/plugin)
	if err := SetupTunRouteRules(); err != nil {
		// Log warning but continue - the reserved address check still provides protection
		// This is mainly for Linux systems with root privileges
	}

	ctx, cancel := context.WithCancel(context.Background())
	tunStack, err := tun.NewStack(string(stack), tun.StackOptions{
		Context:         ctx,
		Tun:             tunInterface,
		TunOptions:      tunOptions,
		UDPTimeout:      30,
		Handler:         t,
		Logger:          defaultLogger,
		InterfaceFinder: control.NewDefaultInterfaceFinder(),
	})
	if err != nil {
		cancel()
		return err
	}
	failedCloser = append(failedCloser, tunStack)
	err = tunStack.Start()
	if err != nil {
		cancel()
		return err
	}
	t.cancel = cancel
	t.closer = failedCloser
	failedCloser = nil
	t.waiter = &gvisorWaiter{tunStack}

	// Note: Server addresses are now excluded via Inet4/6RouteExcludeAddress
	// No need for manual static routes - sing-tun handles it natively

	t.dns.whitelist, _ = GetWhitelistCN()
	// Route DNS to local dokodemo-door listener to avoid SOCKS loop
	t.dns.servers = []M.Socksaddr{M.ParseSocksaddrHostPort("127.0.0.1", TunDNSListenPort)}
	backupDNS := make(map[string][]string)
	interfaces, _ := dns.GetValidNetworkInterfaces()
	for _, ifi := range interfaces {
		backupDNS[ifi], _ = dns.ReplaceDNSServer(ifi, dnsAddr)
	}
	t.backupDNS = backupDNS
	t.whitelist = nil
	return nil
}

func (t *singTun) Close() error {
	t.mu.Lock()
	if t.cancel != nil {
		for ifi, server := range t.backupDNS {
			dns.SetDNSServer(ifi, server...)
		}
		t.backupDNS = nil
		t.cancel()
		t.closer.Close()
		t.cancel = nil
		t.closer = nil
		if t.waiter != nil {
			t.waiter.Wait()
			t.waiter = nil
		}
		// Cleanup routing rules
		CleanupTunRouteRules()
		// Cleanup exclude routes on non-Linux platforms
		if err := CleanupExcludeRoutes(); err != nil {
			log.Warn("[TUN] Failed to cleanup exclude routes: %v", err)
		}
		// Clear whitelist and exclusion list
		t.whitelist = nil
		t.excludeAddrs = nil
	}
	t.mu.Unlock()
	return nil
}

func (t *singTun) AddDomainWhitelist(domain string) {
	t.dns.whitelist = append(t.dns.whitelist, strmatcher.FullMatcher(domain))
	log.Trace("[TUN] Added domain to DNS whitelist: %s", domain)
}

func (t *singTun) AddIPWhitelist(addr netip.Addr) {
	t.whitelist = append(t.whitelist, addr)
	log.Info("[TUN] Added IP to whitelist: %s", addr.String())
	// Also add to route exclusion list to prevent routing through TUN
	// This is critical for Windows/macOS where fwmark routing is not available
	prefix := netip.PrefixFrom(addr, addr.BitLen())
	t.excludeAddrs = append(t.excludeAddrs, prefix)
	log.Info("[TUN] Added %s to route exclusion list", prefix.String())
}

func (t *singTun) SetFakeIP(enabled bool) {
	t.dns.useFakeIP = enabled
}

func (t *singTun) SetIPv6(enabled bool) {
	t.useIPv6 = enabled
}

func (t *singTun) SetStrictRoute(enabled bool) {
	t.strictRoute = enabled
}

func (t *singTun) PrepareConnection(network string, source M.Socksaddr, destination M.Socksaddr) error {
	return nil
}

func (t *singTun) NewConnectionEx(ctx context.Context, conn net.Conn, source M.Socksaddr, destination M.Socksaddr, onClose N.CloseHandlerFunc) {
	metadata := M.Metadata{
		Source:      source,
		Destination: destination,
	}
	log.Trace("[TUN-NEW] New TCP connection: %s -> %s", source, destination)
	err := t.newConnection(ctx, conn, metadata)
	if err != nil {
		N.CloseOnHandshakeFailure(conn, onClose, err)
	}
}

func (t *singTun) newConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	err := t.dns.NewConnection(ctx, conn, metadata)
	if err == continueHandler {
		var dialer N.Dialer
		var dialType string
		// Use direct connection for whitelisted IPs or reserved address ranges
		if slices.Contains(t.whitelist, metadata.Destination.Addr) || isReservedAddress(metadata.Destination.Addr) {
			dialer = t.dialer
			dialType = "direct"
			log.Trace("[TUN-TCP] %s -> %s: using DIRECT (whitelisted/reserved)", metadata.Source, metadata.Destination)
		} else {
			dialer = t.forward
			dialType = "socks5"
			log.Trace("[TUN-TCP] %s -> %s: using SOCKS5", metadata.Source, metadata.Destination)
		}
		if domain, ok := t.dns.fakeCache.Load(metadata.Destination.Addr); ok {
			metadata.Destination.Addr = netip.Addr{}
			metadata.Destination.Fqdn = domain
		}
		serverConn, err := dialer.DialContext(ctx, N.NetworkTCP, metadata.Destination)
		if err != nil {
			log.Warn("[TUN-TCP] Failed to dial %s via %s: %v", metadata.Destination, dialType, err)
			conn.Close()
			return err
		}
		log.Trace("[TUN-TCP] Connected to %s via %s", metadata.Destination, dialType)
		err = bufio.CopyConn(ctx, conn, serverConn)
		if err != nil {
			log.Warn("[TUN-TCP] Relay failed %s -> %s via %s: %v", metadata.Source, metadata.Destination, dialType, err)
			return err
		}
		log.Trace("[TUN-TCP] Relay closed %s -> %s via %s", metadata.Source, metadata.Destination, dialType)
		return nil
	}
	return err
}

func (t *singTun) NewPacketConnectionEx(ctx context.Context, conn N.PacketConn, source M.Socksaddr, destination M.Socksaddr, onClose N.CloseHandlerFunc) {
	metadata := M.Metadata{
		Source:      source,
		Destination: destination,
	}
	err := t.newPacketConnection(ctx, conn, metadata)
	if err != nil {
		N.CloseOnHandshakeFailure(conn, onClose, err)
	}
}

func (t *singTun) newPacketConnection(ctx context.Context, conn N.PacketConn, metadata M.Metadata) error {
	err := t.dns.NewPacketConnection(ctx, conn, metadata)
	if err == continueHandler {
		var dialer N.Dialer
		// Use direct connection for whitelisted IPs or reserved address ranges
		if slices.Contains(t.whitelist, metadata.Destination.Addr) || isReservedAddress(metadata.Destination.Addr) {
			dialer = t.dialer
			log.Trace("[TUN-UDP] %s -> %s: using DIRECT (whitelisted/reserved)", metadata.Source, metadata.Destination)
		} else {
			dialer = t.forward
			log.Trace("[TUN-UDP] %s -> %s: using SOCKS5", metadata.Source, metadata.Destination)
		}
		if domain, ok := t.dns.fakeCache.Load(metadata.Destination.Addr); ok {
			metadata.Destination.Addr = netip.Addr{}
			metadata.Destination.Fqdn = domain
		}
		serverConn, err := dialer.ListenPacket(ctx, metadata.Destination)
		if err != nil {
			conn.Close()
			return err
		}
		err = bufio.CopyPacketConn(ctx, conn, bufio.NewPacketConn(serverConn))
		if err != nil {
			log.Warn("[TUN-UDP] Relay failed %s -> %s: %v", metadata.Source, metadata.Destination, err)
			return err
		}
		return nil
	}
	return err
}

func (t *singTun) NewError(ctx context.Context, err error) {
}

func GetWhitelistCN() (Matcher, error) {
	datpath, err := asset.GetV2rayLocationAsset("geosite.dat")
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(datpath)
	if err != nil {
		return nil, fmt.Errorf("GetWhitelistCn: %w", err)
	}
	var siteList routercommon.GeoSiteList
	err = proto.Unmarshal(b, &siteList)
	if err != nil {
		return nil, fmt.Errorf("GetWhitelistCn: %w", err)
	}
	var matcher Matcher
	for _, e := range siteList.Entry {
		if e.CountryCode == "CN" {
			for _, dm := range e.Domain {
				switch dm.Type {
				case routercommon.Domain_Plain:
					matcher = append(matcher, strmatcher.SubstrMatcher(dm.Value))
				case routercommon.Domain_Regex:
					r, err := strmatcher.Regex.New(dm.Value)
					if err != nil {
						continue
					}
					matcher = append(matcher, r)
				case routercommon.Domain_RootDomain:
					matcher = append(matcher, strmatcher.DomainMatcher(dm.Value))
				case routercommon.Domain_Full:
					matcher = append(matcher, strmatcher.FullMatcher(dm.Value))
				}
			}
			break
		}
	}
	return matcher, nil
}
