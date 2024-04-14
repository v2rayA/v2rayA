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
	"google.golang.org/protobuf/proto"
)

const (
	dnsAddr = "172.19.0.2"
)

var (
	prefix4 = netip.MustParsePrefix("172.19.0.1/30")
	prefix6 = netip.MustParsePrefix("fdfe:dcba:9876::1/126")

	defaultLogger = logger.NOP()

	continueHandler = errors.New("continue handler")
)

type singTun struct {
	mu        sync.Mutex
	dialer    N.Dialer
	forward   N.Dialer
	cancel    context.CancelFunc
	closer    io.Closer
	waiter    *gvisorWaiter
	dns       *DNS
	backupDNS map[string][]string
	whitelist []netip.Addr
}

func NewSingTun() Tun {
	dialer := N.SystemDialer
	client := socks.NewClient(dialer, M.ParseSocksaddrHostPort("127.0.0.1", 52345), socks.Version5, "", "")
	return &singTun{
		dialer:  dialer,
		forward: client,
		dns:     NewDNS(dialer, client, M.ParseSocksaddrHostPort(dnsAddr, 53)),
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
	tunOptions := tun.Options{
		Name:         tun.CalculateInterfaceName(""),
		MTU:          9000,
		Inet4Address: []netip.Prefix{prefix4},
		// Inet6Address:     []netip.Prefix{prefix6},
		AutoRoute:        true,
		StrictRoute:      false,
		InterfaceMonitor: interfaceMonitor,
		TableIndex:       2022,
	}
	tunInterface, err := tun.New(tunOptions)
	if err != nil {
		return err
	}
	failedCloser = append(failedCloser, tunInterface)
	ctx, cancel := context.WithCancel(context.Background())
	tunStack, err := tun.NewStack(string(stack), tun.StackOptions{
		Context:                ctx,
		Tun:                    tunInterface,
		MTU:                    tunOptions.MTU,
		Name:                   tunOptions.Name,
		Inet4Address:           tunOptions.Inet4Address,
		Inet6Address:           tunOptions.Inet6Address,
		EndpointIndependentNat: false,
		UDPTimeout:             30,
		Handler:                t,
		Logger:                 defaultLogger,
		InterfaceFinder:        control.DefaultInterfaceFinder(),
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
	t.dns.whitelist, _ = GetWhitelistCN()
	servers := dns.GetSystemDNS()
	t.dns.servers = make([]M.Socksaddr, len(servers))
	for i, addr := range servers {
		t.dns.servers[i] = M.SocksaddrFromNetIP(addr)
	}
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
	}
	t.mu.Unlock()
	return nil
}

func (t *singTun) AddDomainWhitelist(domain string) {
	t.dns.whitelist = append(t.dns.whitelist, strmatcher.FullMatcher(domain))
}

func (t *singTun) AddIPWhitelist(addr netip.Addr) {
	t.whitelist = append(t.whitelist, addr)
}

func (t *singTun) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	err := t.dns.NewConnection(ctx, conn, metadata)
	if err == continueHandler {
		var dialer N.Dialer
		if slices.Contains(t.whitelist, metadata.Destination.Addr) {
			dialer = t.dialer
		} else {
			dialer = t.forward
		}
		if domain, ok := t.dns.fakeCache.Load(metadata.Destination.Addr); ok {
			metadata.Destination.Addr = netip.Addr{}
			metadata.Destination.Fqdn = domain
		}
		serverConn, err := dialer.DialContext(ctx, N.NetworkTCP, metadata.Destination)
		if err != nil {
			conn.Close()
			return err
		}
		return bufio.CopyConn(ctx, conn, serverConn)
	}
	return err
}

func (t *singTun) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata M.Metadata) error {
	err := t.dns.NewPacketConnection(ctx, conn, metadata)
	if err == continueHandler {
		var dialer N.Dialer
		if slices.Contains(t.whitelist, metadata.Destination.Addr) {
			dialer = t.dialer
		} else {
			dialer = t.forward
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
		return bufio.CopyPacketConn(ctx, conn, bufio.NewPacketConn(serverConn))
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
