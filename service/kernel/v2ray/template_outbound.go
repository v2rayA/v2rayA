package v2ray

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/mohae/deepcopy"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func (t *Template) SetOutboundSockopt() {
	mark := 0x80
	//tos := 184
	for i := range t.Outbounds {
		if t.Outbounds[i].Protocol == "blackhole" || t.Outbounds[i].Protocol == "dns" {
			continue
		}
		if t.Outbounds[i].StreamSettings == nil {
			t.Outbounds[i].StreamSettings = new(coreObj.StreamSettings)
		}
		if t.Outbounds[i].StreamSettings.Sockopt == nil {
			t.Outbounds[i].StreamSettings.Sockopt = new(coreObj.Sockopt)
		}
		if t.Outbounds[i].Protocol == "freedom" && t.Outbounds[i].Tag == "direct" {
			t.Outbounds[i].Settings.DomainStrategy = "UseIP"
		}
		if t.Setting.TcpFastOpen != configure.Default {
			tmp := t.Setting.TcpFastOpen == configure.Yes
			t.Outbounds[i].StreamSettings.Sockopt.TCPFastOpen = &tmp
		}
		t.checkAndSetMark(&t.Outbounds[i], mark)
	}
}
func (t *Template) setDualStack() {
	const (
		tag4Suffix = "_ipv4"
		tag6Suffix = "_ipv6"
	)
	tagMap := make(map[string]struct{})
	inbounds6 := deepcopy.Copy(t.Inbounds).([]coreObj.Inbound)

	// Add ::1 twins for every inbound that listens on a 127.x loopback address.
	// Inbounds on 0.0.0.0 (LAN-shared) are not duplicated.
	// Special exclusions:
	//   transparent+Redirect  – kernel REDIRECT requires 0.0.0.0
	//   dns-in                – receives the 127.2.0.17 treatment below
	for i := range t.Inbounds {
		tag := t.Inbounds[i].Tag
		if tag == "transparent" && t.Setting.TransparentType == configure.TransparentRedirect {
			// https://ipset.netfilter.org/iptables-extensions.man.html#lbDK
			// REDIRECT rewrites the destination to the primary address of the incoming interface,
			// so the inbound must stay at 0.0.0.0.
			inbounds6[i].Tag = "THIS_IS_A_DROPPED_TAG"
			continue
		}
		if tag == "dns-in" {
			// Handled separately — skip generic duplication.
			inbounds6[i].Tag = "THIS_IS_A_DROPPED_TAG"
			continue
		}
		if !strings.HasPrefix(t.Inbounds[i].Listen, "127.") {
			// 0.0.0.0 or other non-loopback address — no ::1 twin needed.
			inbounds6[i].Tag = "THIS_IS_A_DROPPED_TAG"
			continue
		}
		inbounds6[i].Listen = "::1"
		if tag != "" {
			tagMap[tag] = struct{}{}
			t.Inbounds[i].Tag += tag4Suffix
			inbounds6[i].Tag += tag6Suffix
		}
	}

	for i := len(inbounds6) - 1; i >= 0; i-- {
		if inbounds6[i].Tag == "THIS_IS_A_DROPPED_TAG" {
			inbounds6 = append(inbounds6[:i], inbounds6[i+1:]...)
		}
	}

	if iptables.IsIPv6Supported() {
		t.Inbounds = append(t.Inbounds, inbounds6...)
	}

	// Update routing rules with _ipv4/_ipv6 suffixes for duplicated inbounds.
	for i := range t.Routing.Rules {
		tag6 := make([]string, 0)
		for j, tag := range t.Routing.Rules[i].InboundTag {
			if _, ok := tagMap[tag]; ok {
				t.Routing.Rules[i].InboundTag[j] += tag4Suffix
				tag6 = append(tag6, tag+tag6Suffix)
			}
		}
		if v6supported := iptables.IsIPv6Supported(); len(tag6) > 0 && v6supported {
			t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, tag6...)
		}
	}

	// dns-in special handling: always bind to 127.2.0.17 for split-routing local queries.
	// When PortSharing is on the inbound also stays on 0.0.0.0 for LAN DNS, so we add a
	// second copy at 127.2.0.17 with tag dns-in-local.
	for i := range t.Inbounds {
		if t.Inbounds[i].Tag != "dns-in" {
			continue
		}
		if !t.Setting.PortSharing {
			// PortSharing off: move the single dns-in inbound to loopback-only.
			t.Inbounds[i].Listen = "127.2.0.17"
		} else {
			// PortSharing on: keep 0.0.0.0 for LAN; also add a local copy.
			if couldListenLocalhost, e := CouldLocalDnsListen(); couldListenLocalhost && e != nil {
				// Port 53 is already in use on localhost; only listen on the special address.
				t.Inbounds[i].Listen = "127.2.0.17"
			} else {
				localDnsInbound := t.Inbounds[i]
				localDnsInbound.Listen = "127.2.0.17"
				localDnsInbound.Tag = "dns-in-local"
				t.Inbounds = append(t.Inbounds, localDnsInbound)
				// Add dns-in-local to any routing rule that references dns-in.
				for ri := range t.Routing.Rules {
					for _, rtag := range t.Routing.Rules[ri].InboundTag {
						if rtag == "dns-in" {
							t.Routing.Rules[ri].InboundTag = append(t.Routing.Rules[ri].InboundTag, "dns-in-local")
							break
						}
					}
				}
			}
		}
		break
	}
}
func (t *Template) setSendThrough() {
	for i := 0; i < len(t.Outbounds); i++ {
		outbound := &t.Outbounds[i]

		// Get server info for this outbound
		sInfo, exists := t.serverInfoMap[outbound.Tag]
		if !exists {
			continue
		}

		// Determine the appropriate local address
		sendThrough := t.getSendThroughForServer(sInfo)
		if sendThrough != "" {
			outbound.SendThrough = sendThrough
			log.Trace("[v2ray] Set sendThrough for %s: %s", outbound.Tag, sendThrough)
		}
	}
}

func (t *Template) getSendThroughForServer(sInfo *serverInfo) string {
	serverHost := sInfo.Info.GetHostname()

	// If connecting to plugin (localhost), use 127.0.0.1
	if sInfo.PluginPort > 0 {
		return "127.0.0.1"
	}

	// Check if server is IPv4 or IPv6
	if serverIP := net.ParseIP(serverHost); serverIP != nil {
		if serverIP.To4() != nil {
			// IPv4 server - use IPv4 local address
			if ip, err := GetBestLocalIP(false); err == nil {
				return ip.String()
			}
		} else {
			// IPv6 server - use IPv6 local address
			if ip, err := GetBestLocalIP(true); err == nil {
				return ip.String()
			}
			// Fallback to link-local IPv6 if no global address available
			if ip, err := GetLinkLocalIPv6(); err == nil {
				return ip.String()
			}
		}
	} else {
		// Domain name - prefer IPv4
		if ip, err := GetBestLocalIP(false); err == nil {
			return ip.String()
		}
	}

	return ""
}

// GetBestLocalIP returns the best local IP address, skipping TUN, loopback, and APIPA addresses
func resolveEffectiveBackend(obj serverObj.ServerObj, setting *configure.Setting) string {
	bg, ok := obj.(serverObj.BackendGetter)
	if !ok {
		return ""
	}
	nodeBackend := bg.GetBackend()
	if nodeBackend != "" {
		return nodeBackend
	}
	// Fall back to system setting
	switch obj.GetProtocol() {
	case "shadowsocks", "ss":
		return setting.SsBackend
	case "trojan", "trojan-go":
		return setting.TrojanBackend
	}
	return ""
}

func (t *Template) resolveOutbounds(
	serverData *ServerData,
) (supportUDP map[string]bool, outboundTags []string, err error) {

	supportUDP = make(map[string]bool)
	t.serverInfoMap = make(map[string]*serverInfo)
	type _outbound struct {
		weight   int
		outbound coreObj.OutboundObject
		balancer bool
	}
	serverInfo2Index := make(map[*serverInfo]int)
	for i := range serverData.ServerInfos {
		serverInfo2Index[&serverData.ServerInfos[i]] = i
	}
	// keep order with serverInfos
	outboundTags = make([]string, len(serverData.ServerInfos))
	var extraOutbounds []coreObj.OutboundObject
	var outbounds []_outbound
	setting := configure.GetSettingNotNil()
	for obj, sInfos := range serverData.ServerObj2ServerInfos() {
		var (
			usedByBalancer bool
		)
		// an vmessInfo(server template) may be used by multiple serverInfos(a connected server)

		// outbound name is not just v2ray outbound tag, it may be a balancer
		type balancer struct {
			name       string
			serverInfo *serverInfo
		}
		var balancers []balancer
		for _, sInfo := range sInfos {
			if len(serverData.OutboundName2ServerObjs[sInfo.OutboundName]) > 1 {
				// balancer
				if !usedByBalancer {
					usedByBalancer = true
				}
				balancers = append(balancers, balancer{
					name:       sInfo.OutboundName,
					serverInfo: sInfo,
				})
			} else {
				// pure outbound
				outboundTag := sInfo.OutboundName
				c, err := obj.Configuration(serverObj.PriorInfo{
					Variant:     t.Variant,
					CoreVersion: t.CoreVersion,
					Tag:         outboundTag,
					PluginPort:  sInfo.PluginPort,
					Backend:     resolveEffectiveBackend(obj, setting),
				})
				if err != nil {
					return nil, nil, err
				}
				// Store server info for this outbound
				t.serverInfoMap[outboundTag] = sInfo
				extraOutbounds = append(extraOutbounds, c.ExtraOutbounds...)
				outbounds = append(outbounds, _outbound{
					weight:   serverInfo2Index[sInfo],
					outbound: c.CoreOutbound,
					balancer: false,
				})
				outboundTags[serverInfo2Index[sInfo]] = outboundTag
				supportUDP[sInfo.OutboundName] = c.UDPSupport
			}
		}
		if usedByBalancer {
			// the v2ray outbound is shared by balancers
			outboundTag := GroupWrapper(obj.GetName())
			c, err := obj.Configuration(serverObj.PriorInfo{
				Variant:     t.Variant,
				CoreVersion: t.CoreVersion,
				Tag:         outboundTag,
				PluginPort:  0,
				Backend:     resolveEffectiveBackend(obj, setting),
			})
			if err != nil {
				// Store server info for balancer outbound (use first balancer's info)
				if len(balancers) > 0 {
					t.serverInfoMap[outboundTag] = balancers[0].serverInfo
				}
				return nil, nil, err
			}
			extraOutbounds = append(extraOutbounds, c.ExtraOutbounds...)
			for _, v := range balancers {
				c.CoreOutbound.Balancers = append(c.CoreOutbound.Balancers, v.name)
			}
			// we use the lowest serverInfo index as the order weight of the balancer outbound
			weight := -1
			for _, v := range balancers {
				index := serverInfo2Index[v.serverInfo]
				if weight == -1 || weight > index {
					weight = index
				}
				// tag
				outboundTags[index] = outboundTag
			}
			outbounds = append(outbounds, _outbound{
				weight:   weight,
				outbound: c.CoreOutbound,
				balancer: true,
			})

			// if any node does not support UDP, the outbound should be tagged as UDP unsupported
			for _, outboundName := range c.CoreOutbound.Balancers {
				_supportUDP := c.UDPSupport
				if _, ok := supportUDP[outboundName]; !ok {
					supportUDP[outboundName] = _supportUDP
				}
				if supportUDP[outboundName] && !_supportUDP {
					supportUDP[outboundName] = false
				}
			}
		}
	}
	sort.Slice(outbounds, func(i, j int) bool {
		return outbounds[i].weight < outbounds[j].weight
	})
	for _, v := range outbounds {
		t.Outbounds = append(t.Outbounds, v.outbound)
	}
	t.Outbounds = append(t.Outbounds, coreObj.OutboundObject{
		Tag:      "direct",
		Protocol: "freedom",
	}, coreObj.OutboundObject{
		Tag:      "block",
		Protocol: "blackhole",
	})
	t.Outbounds = append(t.Outbounds, extraOutbounds...)
	return supportUDP, outboundTags, nil
}

func (t *Template) checkAndSetMark(o *coreObj.OutboundObject, mark int) {
	if !IsTransparentOn(t.Setting) {
		return
	}
	if o.StreamSettings == nil {
		o.StreamSettings = new(coreObj.StreamSettings)
	}
	if o.StreamSettings.Sockopt == nil {
		o.StreamSettings.Sockopt = new(coreObj.Sockopt)
	}
	o.StreamSettings.Sockopt.Mark = &mark
}

func (t *Template) InsertMappingOutbound(o serverObj.ServerObj, inboundPort string, udpSupport bool, pluginPort int, protocol string) (err error) {
	if t.CoreVersion == "" {
		t.Variant, t.CoreVersion, _ = where.GetV2rayServiceVersion()
	}
	c, err := o.Configuration(serverObj.PriorInfo{
		Variant:     t.Variant,
		CoreVersion: t.CoreVersion,
		Tag:         "outbound" + inboundPort,
		PluginPort:  pluginPort,
		Backend:     resolveEffectiveBackend(o, configure.GetSettingNotNil()),
	})
	if err != nil {
		return err
	}
	var mark = 0x80
	t.checkAndSetMark(&c.CoreOutbound, mark)
	t.Outbounds = append(t.Outbounds, c.CoreOutbound)
	t.Outbounds = append(t.Outbounds, c.ExtraOutbounds...)
	iPort, err := strconv.Atoi(inboundPort)
	if err != nil || iPort <= 0 {
		return fmt.Errorf("port of inbound must be a positive number with string type")
	}
	if protocol == "" {
		protocol = "socks"
	}
	t.Inbounds = append(t.Inbounds, coreObj.Inbound{
		Port:     iPort,
		Protocol: protocol,
		Listen:   "0.0.0.0",
		Sniffing: coreObj.Sniffing{
			Enabled:      false,
			DestOverride: []string{"http", "tls"},
		},
		Settings: &coreObj.InboundSettings{
			Auth: "noauth",
			UDP:  udpSupport,
		},
		Tag: "inbound" + inboundPort,
	})
	if t.Routing.DomainStrategy == "" {
		t.Routing.DomainStrategy = "IPOnDemand"
	}
	// Insert at the beginning
	tmp := make([]coreObj.RoutingRule, 1, len(t.Routing.Rules)+1)
	tmp[0] = coreObj.RoutingRule{
		Type:        "field",
		OutboundTag: "outbound" + inboundPort,
		InboundTag:  []string{"inbound" + inboundPort},
	}
	t.Routing.Rules = append(tmp, t.Routing.Rules...)
	return
}
