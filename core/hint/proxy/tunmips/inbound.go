// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"sync"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/policy"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/transport/internet/stat"
	"golang.zx2c4.com/wireguard/tun"
)

// Handler is the "tun-mips" inbound. Like xray's tun inbound it listens on
// no port: its input is the device it opens in Start.
type Handler struct {
	ctx     context.Context
	config  *Config
	tag     string
	sniff   session.SniffingRequest
	pm      policy.Manager
	dsp     routing.Dispatcher
	addr4   netip.Prefix
	addr6   netip.Prefix
	uplink  stats.Counter
	downlnk stats.Counter

	mu    sync.Mutex
	dev   tun.Device
	stack *mipsStack
	fwd   *forwarder
}

func newHandler(ctx context.Context, config *Config) (*Handler, error) {
	h := &Handler{ctx: ctx, config: config}
	if config.Name == "" {
		return nil, errors.New("tun-mips: name is required")
	}
	if config.Mtu == 0 {
		return nil, errors.New("tun-mips: mtu is required")
	}
	var err error
	if h.addr4, err = netip.ParsePrefix(config.Address4); err != nil || !h.addr4.Addr().Is4() {
		return nil, errors.New("tun-mips: address4 must be an IPv4 prefix, got ", fmt.Sprintf("%q", config.Address4))
	}
	if config.Address6 != "" {
		if h.addr6, err = netip.ParsePrefix(config.Address6); err != nil || !h.addr6.Addr().Is6() {
			return nil, errors.New("tun-mips: address6 must be an IPv6 prefix, got ", fmt.Sprintf("%q", config.Address6))
		}
	}
	if config.DnsTarget == "" {
		return nil, errors.New("tun-mips: dnsTarget is required")
	}
	if _, err := netip.ParseAddrPort(config.DnsTarget); err != nil {
		return nil, errors.New("tun-mips: dnsTarget must be host:port, got ", fmt.Sprintf("%q", config.DnsTarget))
	}
	return h, nil
}

// Init receives the tag and sniffing request the proxyman set in ctx.
func (h *Handler) Init(ctx context.Context, pm policy.Manager, dsp routing.Dispatcher) error {
	if inbound := session.InboundFromContext(ctx); inbound != nil {
		h.tag = inbound.Tag
	}
	if content := session.ContentFromContext(ctx); content != nil {
		h.sniff = content.SniffingRequest
	}
	h.ctx = core.ToBackgroundDetachedContext(ctx)
	h.pm, h.dsp = pm, dsp
	if h.tag != "" {
		if sm, ok := core.MustFromContext(ctx).GetFeature(stats.ManagerType()).(stats.Manager); ok {
			if pm.ForSystem().Stats.InboundUplink {
				h.uplink, _ = sm.GetOrRegisterCounter("inbound>>>" + h.tag + ">>>traffic>>>uplink")
			}
			if pm.ForSystem().Stats.InboundDownlink {
				h.downlnk, _ = sm.GetOrRegisterCounter("inbound>>>" + h.tag + ">>>traffic>>>downlink")
			}
		}
	}
	return nil
}

// Start opens the device, assigns its addresses and starts the stack.
func (h *Handler) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	dev, err := openDevice(h.config.Name, h.config.Mtu, h.addr4, h.addr6)
	if err != nil {
		return err
	}
	name, _ := dev.Name()

	fwd := newForwarder(h.ctx, h.dsp)
	fwd.tag = h.tag
	fwd.sniffing = h.sniff
	fwd.userLevel = h.config.UserLevel
	fwd.uplink, fwd.downlink = h.uplink, h.downlnk
	fwd.dns = newDNSHijack(h.config.DnsTarget)
	selfPids := append([]uint32{uint32(os.Getpid())}, h.config.SelfPids...)
	fwd.excl = newExclusion(h.config.ExcludeProcesses, selfPids)
	fwd.directTag = h.config.DirectTag
	if fwd.directTag == "" {
		fwd.directTag = "direct"
	}

	stack := newMipsStack(h.config.Mtu)
	if err := stack.Start(dev, fwd.handlers()); err != nil {
		dev.Close()
		return errors.New("tun-mips: start stack on ", name).Base(err)
	}
	h.dev, h.stack, h.fwd = dev, stack, fwd
	errors.LogInfo(h.ctx, "tun-mips: ", name, " up, ", h.addr4, " ", h.config.Address6)
	go h.watch(name, stack, dev)
	return nil
}

// watch reports a pump failure and closes the device. The core keeps
// running, as with xray's tun inbound; the service's connectivity monitor
// sees the device disappear and pauses the transparent proxy.
func (h *Handler) watch(name string, stack *mipsStack, dev tun.Device) {
	<-stack.done
	if err := stack.Err(); err != nil {
		errors.LogError(h.ctx, "tun-mips: ", name, " stopped: ", err)
	}
	dev.Close()
}

func (h *Handler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fwd != nil && h.fwd.dns != nil {
		h.fwd.dns.close()
	}
	if h.stack != nil {
		h.stack.Close()
	}
	if h.dev != nil {
		h.dev.Close()
	}
	h.dev, h.stack, h.fwd = nil, nil, nil
	return nil
}

// Network declares no listening network: the proxyman opens no port.
func (h *Handler) Network() []net.Network { return []net.Network{} }

// Process is never called because nothing listens.
func (h *Handler) Process(context.Context, net.Network, stat.Connection, routing.Dispatcher) error {
	return nil
}

var _ common.Runnable = (*Handler)(nil)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		h, err := newHandler(ctx, config.(*Config))
		if err != nil {
			return nil, err
		}
		err = core.RequireFeatures(ctx, func(pm policy.Manager, dsp routing.Dispatcher) error {
			return h.Init(ctx, pm, dsp)
		})
		return h, err
	}))
}
