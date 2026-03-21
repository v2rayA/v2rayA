package tun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"slices"
	"sync"
	"time"

	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/control"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/protocol/socks"
	"github.com/v2fly/v2ray-core/v5/common/strmatcher"
	"github.com/v2rayA/v2ray-lib/router/routercommon"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"google.golang.org/protobuf/proto"
)

const (
	// DNS addresses should point to the TUN interface itself, not gateway
	// sing-tun intercepts DNS traffic to port 53 and handles it internally
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
	mu               sync.Mutex
	dialer           N.Dialer
	forward          N.Dialer
	cancel           context.CancelFunc
	closer           io.Closer
	waiter           *gvisorWaiter
	dns              *DNS
	whitelist        []netip.Addr
	excludeAddrs     []netip.Prefix // Addresses to exclude from TUN routing
	useIPv6          bool
	strictRoute      bool
	autoRoute        bool
	runningAutoRoute bool   // effective autoRoute after platform override
	tunName          string // TUN interface name for cleanup
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

// resolveDnsHost resolves a hostname to both A and AAAA records.
// Returns normalised IP addresses (IPv4-in-IPv6 addresses are unmapped to
// pure IPv4 so they match the canonical form sing-tun uses).
func resolveDnsHost(host string) []netip.Addr {
	var ips []netip.Addr

	// Already an IP address – just normalise and return.
	if addr, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{addr.Unmap()}
	}

	// Resolve A + AAAA records.
	if addrs, err := net.LookupIP(host); err == nil {
		for _, addr := range addrs {
			if ipAddr, ok := netip.AddrFromSlice(addr); ok {
				// net.LookupIP may return IPv4 as 16-byte IPv4-in-IPv6.
				// Unmap to canonical Is4() so whitelist lookups match.
				ips = append(ips, ipAddr.Unmap())
			}
		}
	} else {
		log.Warn("[TUN] Failed to resolve DNS host %s: %v", host, err)
	}

	return ips
}

// ResolveDnsServersToExcludes resolves DNS server hostnames to IP prefixes for TUN exclusion
// This prevents DNS server traffic from being intercepted by TUN, avoiding routing loops
func ResolveDnsServersToExcludes(dnsHosts []string) []netip.Prefix {
	var excludes []netip.Prefix
	seen := make(map[netip.Addr]bool)

	log.Info("[TUN] Resolving DNS servers for exclusion: %v", dnsHosts)

	for _, host := range dnsHosts {
		ips := resolveDnsHost(host)
		for _, ip := range ips {
			if seen[ip] {
				continue
			}
			seen[ip] = true

			// Convert IP to /32 (IPv4) or /128 (IPv6) prefix
			var prefix netip.Prefix
			if ip.Is4() {
				prefix = netip.PrefixFrom(ip, 32)
			} else {
				prefix = netip.PrefixFrom(ip, 128)
			}
			excludes = append(excludes, prefix)
			log.Info("[TUN] Added DNS server %s (%s) to exclusion list", host, ip)
		}
	}

	return excludes
}

func NewSingTun() Tun {
	dialer := N.SystemDialer
	client := socks.NewClient(dialer, M.ParseSocksaddrHostPort("127.0.0.1", 52345), socks.Version5, "", "")
	log.Info("[TUN] Initialized SOCKS5 client to 127.0.0.1:52345")
	return &singTun{
		dialer:      dialer,
		forward:     client,
		strictRoute: false,
		autoRoute:   true, // Default to enabled
		// DNS: dialer is for forwarding to dokodemo-door (127.0.0.1:6053)
		// forward is nil, so all DNS queries will use dnsForward mode
		// No DNS server addresses - TUN only forwards to port 6053
		dns: NewDNS(dialer, nil, false),
	}
}

func (t *singTun) Start(stack Stack) error {
	var failedCloser Closer
	defer func() {
		failedCloser.Close()
	}()

	// ── Snapshot the exclusion/whitelist config set by caller before Start() ──
	// Close() will clear t.excludeAddrs / t.whitelist / t.dns.whitelist,
	// so we must snapshot them first and restore after Close(),
	// in order for the new TUN instance to use these configurations.
	savedExclude := make([]netip.Prefix, len(t.excludeAddrs))
	copy(savedExclude, t.excludeAddrs)
	savedWhitelist := make([]netip.Addr, len(t.whitelist))
	copy(savedWhitelist, t.whitelist)
	savedDomainWhitelist := make(Matcher, len(t.dns.whitelist))
	copy(savedDomainWhitelist, t.dns.whitelist)

	// Pre-exclude addresses that should bypass TUN based on platform (e.g. public DNS on Windows)
	for _, prefix := range platformPreExcludeAddrs() {
		exists := false
		for _, ex := range savedExclude {
			if ex == prefix {
				exists = true
				break
			}
		}
		if !exists {
			savedExclude = append(savedExclude, prefix)
			log.Info("[TUN] Platform pre-excluded address: %s", prefix)
		}
	}

	// Add connected proxy server addresses to exclusion list to prevent routing loops
	proxyServerPrefixes := getConnectedProxyServerPrefixes()
	for _, prefix := range proxyServerPrefixes {
		exists := false
		for _, ex := range savedExclude {
			if ex == prefix {
				exists = true
				break
			}
		}
		if !exists {
			savedExclude = append(savedExclude, prefix)
			log.Info("[TUN] Added proxy server to exclusion list: %s", prefix)
		}
	}

	// Close the old instance (which clears fields), then immediately restore config.
	t.Close()
	t.excludeAddrs = savedExclude
	t.whitelist = savedWhitelist
	t.dns.whitelist = savedDomainWhitelist
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

	autoRoute := t.autoRoute
	strictRoute := t.strictRoute

	// Some platforms (e.g. Windows) need to disable sing-tun's AutoRoute and use manual routing instead
	if platformDisableAutoRoute() && autoRoute {
		autoRoute = false
		log.Info("[TUN] Platform requires disabling AutoRoute, switching to manual routing management")
	} else if !t.autoRoute {
		log.Info("[TUN] AutoRoute disabled by user configuration")
	}

	// Record the effective autoRoute for platform cleanup logic in Close()
	t.runningAutoRoute = autoRoute

	log.Info("[TUN] Starting with StrictRoute=%t, AutoRoute=%t", strictRoute, autoRoute)
	log.Info("[TUN] Total exclusions: %d IPv4, %d IPv6", len(inet4Exclude), len(inet6Exclude))
	if platformDisableAutoRoute() && strictRoute {
		log.Warn("[TUN] Current platform: StrictRoute might only allow traffic from the main process!")
	}

	// Must notify platform module of current AutoRoute mode before SetupExcludeRoutes/SetupTunRouteRules,
	// otherwise Windows tunAutoRoute might keep the last value (e.g. true),
	// causing SetupExcludeRoutes to be skipped when AutoRoute=false this time, leading to routing loops.
	setTunRouteAutoMode(autoRoute)

	// When AutoRoute is enabled, let sing-tun handle exclusion via InetRouteExcludeAddress.
	// Only install OS-level routes on platforms that disable AutoRoute (e.g. Windows)
	// or when the user explicitly turns AutoRoute off.
	if !autoRoute && len(t.excludeAddrs) > 0 {
		if err := SetupExcludeRoutes(t.excludeAddrs); err != nil {
			log.Warn("[TUN] Failed to pre-install exclude routes: %v", err)
		}
	} else if autoRoute {
		log.Info("[TUN] AutoRoute enabled; rely on sing-tun exclude list for %d entries", len(t.excludeAddrs))
	}

	// Interface name is determined by platform function (macOS returns empty string for auto-allocation utun*)
	tunName := platformTunName()
	if tunName == "" {
		log.Info("[TUN] Using system-allocated interface name")
	} else {
		log.Info("[TUN] Using interface name: %s", tunName)
	}

	tunOptions := tun.Options{
		Name:                     tun.CalculateInterfaceName(tunName),
		MTU:                      9000,
		Inet4Address:             []netip.Prefix{prefix4},
		Inet4RouteAddress:        []netip.Prefix{route4},
		Inet4RouteExcludeAddress: inet4Exclude, // Exclude loopback + server IPs
		AutoRoute:                autoRoute,
		StrictRoute:              strictRoute,
		InterfaceMonitor:         interfaceMonitor,
	}

	// Set DNS server for TUN interface
	var dnsServers []netip.Addr
	dnsAddrIP, _ := netip.ParseAddr(dnsAddr)
	if dnsAddrIP.IsValid() {
		dnsServers = append(dnsServers, dnsAddrIP)
		log.Info("[TUN] IPv4 DNS server: %s", dnsAddr)
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

		// Add IPv6 DNS server
		dnsAddrIPv6, _ := netip.ParseAddr(dnsAddr6)
		if dnsAddrIPv6.IsValid() {
			dnsServers = append(dnsServers, dnsAddrIPv6)
			log.Info("[TUN] IPv6 DNS server: %s", dnsAddr6)
		}
	}

	// Set DNS servers (if any)
	if len(dnsServers) > 0 {
		tunOptions.DNSServers = dnsServers
	}
	tunInterface, err := tun.New(tunOptions)
	if err != nil {
		return err
	}
	failedCloser = append(failedCloser, tunInterface)

	// Linux: always add fwmark policy routing rules; Windows/macOS: decided by AutoRoute state
	if err := SetupTunRouteRules(); err != nil {
		return err
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
	t.tunName = tunName // Save for cleanup

	// Perform platform-specific post-start operations (e.g. setting DNS when Windows AutoRoute is off, configuring macOS network service DNS)
	platformPostStart(dnsServers, t.tunName, autoRoute)

	// Append CN geosite list to existing domain whitelist (which includes server domains etc.).
	// Cannot assign directly—that would overwrite entries added via AddDomainWhitelist before Start().
	if cnWhitelist, err := GetWhitelistCN(); err == nil {
		t.dns.whitelist = append(t.dns.whitelist, cnWhitelist...)
	} else {
		log.Warn("[TUN] GetWhitelistCN failed: %v", err)
	}
	// Note: Do not clear t.whitelist here.
	// t.whitelist is the IP whitelist for connection layer (direct connection guarantee), must remain valid throughout TUN lifecycle.
	return nil
}

func (t *singTun) Close() error {
	t.mu.Lock()
	if t.cancel != nil {
		t.cancel()
		t.closer.Close()
		t.cancel = nil
		t.closer = nil
		if t.waiter != nil {
			t.waiter.Wait()
			t.waiter = nil
		}
		// Platform-specific cleanup (restore DNS when Windows is not in AutoRoute, restore macOS networksetup DNS, etc.)
		platformPreClose(t.tunName, t.runningAutoRoute)
		t.tunName = ""
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
	// Normalize address: net.LookupIP often returns IPv4 as 16-byte IPv4-in-IPv6 form
	// (Is4In6()), while sing-tun passes connection targets as 4-byte pure IPv4 (Is4()).
	// Both are not equal in slices.Contains, causing whitelist failure.
	// Store after unified Unmap().
	addr = addr.Unmap()
	if !addr.IsValid() {
		return
	}
	t.whitelist = append(t.whitelist, addr)
	log.Info("[TUN] Added IP to whitelist: %s", addr.String())
	// Also add to route exclusion list: on platforms without fwmark (e.g. Windows/macOS),
	// OS level routing rules are needed to completely block these IPs from entering TUN.
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

func (t *singTun) SetAutoRoute(enabled bool) {
	t.autoRoute = enabled
}

func (t *singTun) PrepareConnection(network string, source M.Socksaddr, destination M.Socksaddr, routeContext tun.DirectRouteContext, timeout time.Duration) (tun.DirectRouteDestination, error) {
	// v2rayA does not implement direct routing (all traffic is handled via newConnectionEx/newPacketConnectionEx)
	return nil, nil
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
			// Add physical interface direct route for this IP on Windows to prevent routing loops.
			// This function is a no-op on platforms like Linux (fwmark).
			DynAddExcludeRoute(metadata.Destination.Addr)
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
	log.Trace("[TUN-NEW] New UDP connection: %s -> %s", source, destination)
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
			DynAddExcludeRoute(metadata.Destination.Addr)
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

// getConnectedProxyServerPrefixes 获取已连接的代理服务器地址前缀列表
// 用于防止代理服务器流量被 TUN 捕获导致路由回环
func getConnectedProxyServerPrefixes() []netip.Prefix {
	var prefixes []netip.Prefix
	seen := make(map[netip.Addr]bool)

	// 获取已连接的服务器信息
	css := configure.GetConnectedServers()
	if css == nil {
		return prefixes
	}

	for _, which := range css.Get() {
		serverRaw, err := which.LocateServerRaw()
		if err != nil {
			log.Warn("[TUN] Failed to locate server for which %+v: %v", which, err)
			continue
		}

		hostname := serverRaw.ServerObj.GetHostname()
		port := serverRaw.ServerObj.GetPort()

		// 解析主机名为 IP 地址
		ips := resolveDnsHost(hostname)
		for _, ip := range ips {
			if seen[ip] {
				continue
			}
			seen[ip] = true

			// 转换为 /32 (IPv4) 或 /128 (IPv6) 前缀
			var prefix netip.Prefix
			if ip.Is4() {
				prefix = netip.PrefixFrom(ip, 32)
			} else {
				prefix = netip.PrefixFrom(ip, 128)
			}
			prefixes = append(prefixes, prefix)
			log.Info("[TUN] Resolved proxy server %s:%d -> %s", hostname, port, ip)
		}
	}

	return prefixes
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
