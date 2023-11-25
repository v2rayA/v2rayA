package tun

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/netip"
	"os"
	"strings"

	D "github.com/miekg/dns"
	tun "github.com/sagernet/sing-tun"
	"github.com/sagernet/sing/common/buf"
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
	prefix4   = netip.MustParsePrefix("172.19.0.1/30")
	prefix6   = netip.MustParsePrefix("fdfe:dcba:9876::1/126")
	dnsServer = netip.MustParseAddrPort(dnsAddr + ":53")

	defaultLogger = logger.NOP()
)

type singTun struct {
	dialer    N.Dialer
	client    *socks.Client
	cancel    context.CancelFunc
	closer    io.Closer
	whitelist Matcher
	systemDNS []netip.AddrPort
	backupDNS map[string][]string
}

func NewSingTun() Tun {
	dialer := N.SystemDialer
	return &singTun{
		dialer: dialer,
		client: socks.NewClient(dialer, M.ParseSocksaddrHostPort("127.0.0.1", 52345), socks.Version5, "", ""),
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
	ctx, cancel := context.WithCancel(context.TODO())
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
	t.whitelist, _ = GetWhitelistCN()
	t.systemDNS = dns.GetSystemDNS()
	backupDNS := make(map[string][]string)
	interfaces, _ := dns.GetValidNetworkInterfaces()
	for _, ifi := range interfaces {
		backupDNS[ifi], _ = dns.ReplaceDNSServer(ifi, dnsAddr)
	}
	t.backupDNS = backupDNS
	return nil
}

func (t *singTun) Close() error {
	if t.cancel != nil {
		t.cancel()
		t.closer.Close()
		t.cancel = nil
		t.closer = nil
		for ifi, server := range t.backupDNS {
			dns.SetDNSServer(ifi, server...)
		}
	}
	return nil
}

func (t *singTun) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	serverConn, err := t.client.DialContext(ctx, N.NetworkTCP, metadata.Destination)
	if err != nil {
		conn.Close()
		return err
	}
	return bufio.CopyConn(ctx, conn, serverConn)
}

func (t *singTun) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata M.Metadata) error {
	if len(t.whitelist) != 0 && len(t.systemDNS) != 0 && metadata.Destination.AddrPort() == dnsServer {
		return t.fakeDNS(ctx, conn)
	}
	serverConn, err := t.client.ListenPacket(ctx, metadata.Destination)
	if err != nil {
		conn.Close()
		return err
	}
	return bufio.CopyPacketConn(ctx, conn, bufio.NewPacketConn(serverConn))
}

func (t *singTun) NewError(ctx context.Context, err error) {
}

func (t *singTun) fakeDNS(ctx context.Context, conn N.PacketConn) error {
	defer conn.Close()
	buffer := buf.NewPacket()
	defer buffer.Release()
	buffer.FullReset()
	_, err := conn.ReadPacket(buffer)
	if err != nil {
		return err
	}
	var msg D.Msg
	err = msg.Unpack(buffer.Bytes())
	if err != nil {
		return err
	}
	bypass := false
	for _, q := range msg.Question {
		domain := strings.TrimSuffix(q.Name, ".")
		if !t.whitelist.Match(domain) {
			bypass = true
			break
		}
	}
	addr := t.getSystemDNS()
	var dialer N.Dialer
	if bypass || addr == dnsServer { //prevent recursion
		dialer = t.client
	} else {
		dialer = t.dialer
	}
	destination := M.SocksaddrFromNetIP(addr)
	serverConn, err := dialer.ListenPacket(ctx, destination)
	if err != nil {
		return err
	}
	conn = bufio.NewCachedPacketConn(conn, buffer, destination)
	return bufio.CopyPacketConn(ctx, conn, bufio.NewPacketConn(serverConn))
}

func (t *singTun) getSystemDNS() netip.AddrPort {
	if len(t.systemDNS) != 1 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(t.systemDNS))))
		if err == nil {
			return t.systemDNS[n.Uint64()]
		}
	}
	return t.systemDNS[0]
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
