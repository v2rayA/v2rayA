package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/net/cnc"
	"github.com/xtls/xray-core/common/session"
	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/routing"
)

// xrayDispatcherAdapter implements dns.RouteDispatcher using xray-core's
// internal routing dispatcher. This allows the DNS module to route queries
// through xray-core's proxy outbounds (VMESS, VLESS, Trojan, etc.) just like
// xray-core's built-in DNS module does.
//
// It sets session.ContextWithInbound on the dispatch context so that xray's
// routing engine selects the appropriate outbound based on routing rules.
type xrayDispatcherAdapter struct {
	dispatcher routing.Dispatcher
}

// newXrayDispatcherAdapter creates a new adapter from an xray core instance.
func newXrayDispatcherAdapter(srv *xcore.Instance) (*xrayDispatcherAdapter, error) {
	var dispatcher routing.Dispatcher
	if err := srv.RequireFeatures(func(d routing.Dispatcher) {
		dispatcher = d
	}, false); err != nil {
		return nil, fmt.Errorf("xray dispatcher not available: %w", err)
	}
	return &xrayDispatcherAdapter{dispatcher: dispatcher}, nil
}

// Dispatch implements dns.RouteDispatcher.
// It creates a TCP connection to addr through xray-core's routing engine.
// The proxyTag is used to set the inbound tag on the context, which xray's
// routing rules use to determine the outbound (like xray-core's DNS module).
func (a *xrayDispatcherAdapter) Dispatch(ctx context.Context, network, addr, proxyTag string) (net.Conn, error) {
	// Parse the destination address.
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		portStr = "53"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 53
	}

	var dest xnet.Destination
	switch strings.ToLower(network) {
	case "tcp", "tcp4", "tcp6":
		dest = xnet.TCPDestination(xnet.ParseAddress(host), xnet.Port(port))
	case "udp", "udp4", "udp6":
		dest = xnet.UDPDestination(xnet.ParseAddress(host), xnet.Port(port))
	default:
		dest = xnet.TCPDestination(xnet.ParseAddress(host), xnet.Port(port))
	}

	// Set the inbound tag on the context so xray's routing engine
	// selects the appropriate outbound (like xray-core's DNS module).
	// For "direct", no special tag is needed (uses default routing).
	dnsCtx := ctx
	if proxyTag != "" && proxyTag != "direct" {
		dnsCtx = session.ContextWithInbound(dnsCtx, &session.Inbound{Tag: proxyTag})
	}
	// Mark as DNS traffic to prevent recursive DNS resolution.
	dnsCtx = session.ContextWithContent(dnsCtx, &session.Content{
		Protocol:       "dns",
		SkipDNSResolve: true,
	})

	// Dispatch through xray-core's routing engine.
	link, err := a.dispatcher.Dispatch(dnsCtx, dest)
	if err != nil {
		return nil, fmt.Errorf("xray dispatch: %w", err)
	}

	// Wrap the transport link into a standard net.Conn.
	return cnc.NewConnection(
		cnc.ConnectionInputMulti(link.Writer),
		cnc.ConnectionOutputMulti(link.Reader),
	), nil
}
