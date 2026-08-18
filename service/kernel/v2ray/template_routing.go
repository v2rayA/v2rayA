package v2ray

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mohae/deepcopy"
	"github.com/v2rayA/RoutingA"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func (t *Template) outNames() map[string]bool {
	tags := make(map[string]bool)
	for _, o := range t.Outbounds {
		if len(o.Balancers) > 0 {
			for _, groupName := range o.Balancers {
				tags[groupName] = true
			}
		} else {
			tags[o.Tag] = false
		}
	}
	return tags
}

func (t *Template) FirstProxyOutboundName(filter func(outboundName string, isGroup bool) bool) (outboundName string, isGroup bool) {
	if filter == nil {
		filter = func(outboundName string, isGroup bool) bool {
			return true
		}
	}
	// deduplicate
	m := make(map[string]struct{})

	for _, o := range t.Outbounds {
		switch o.Tag {
		case "direct", "block", "dns-out":
			continue
		}
		if len(o.Balancers) > 0 {
			for _, v := range o.Balancers {
				if _, ok := m[v]; !ok {
					if filter(v, true) {
						return v, true
					}
					m[v] = struct{}{}
				}
			}
		} else {
			if filter(o.Tag, false) {
				return o.Tag, false
			}
		}
	}
	return
}

// dnsRuleToLines converts a configure.DnsRule to the "server -> outbound" line format
// and the domain list used for the DNS server object.
// Returns (serverLine, []domains). serverLine is empty if server is empty.
func (t *Template) AppendRoutingRuleByMode(mode configure.RulePortMode, inbounds []string) (err error) {
	firstOutboundTag, _ := t.FirstProxyOutboundName(nil)
	// apple pushing. #495 #479
	t.Routing.Rules = append(t.Routing.Rules, coreObj.RoutingRule{
		Type:        "field",
		OutboundTag: "direct",
		InboundTag:  deepcopy.Copy(inbounds).([]string),
		Domain:      []string{"domain:push-apple.com.akadns.net", "domain:push.apple.com"},
	})
	switch mode {
	case configure.WhitelistMode:
		// foreign domains with intranet IP should be proxied first rather than directly connected
		if asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
			t.Routing.Rules = append(t.Routing.Rules,
				coreObj.RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  deepcopy.Copy(inbounds).([]string),
					Domain:      []string{"ext:LoyalsoldierSite.dat:geolocation-!cn"},
				})
		} else {
			t.Routing.Rules = append(t.Routing.Rules,
				coreObj.RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  deepcopy.Copy(inbounds).([]string),
					Domain:      []string{"geosite:geolocation-!cn"},
				})
		}
		t.Routing.Rules = append(t.Routing.Rules,
			coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  deepcopy.Copy(inbounds).([]string),
				// https://github.com/v2rayA/v2rayA/issues/285
				Domain: []string{"geosite:google"},
			},
			coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  deepcopy.Copy(inbounds).([]string),
				Domain:      []string{"geosite:cn"},
			},
			coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  deepcopy.Copy(inbounds).([]string),
				IP:          []string{"geoip:hk", "geoip:mo"},
			},
			coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  deepcopy.Copy(inbounds).([]string),
				IP:          []string{"geoip:private", "geoip:cn"},
			},
		)
	case configure.GfwlistMode:
		if asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
			t.Routing.Rules = append(t.Routing.Rules,
				coreObj.RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  deepcopy.Copy(inbounds).([]string),
					Domain:      []string{"ext:LoyalsoldierSite.dat:gfw"},
				},
				coreObj.RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  deepcopy.Copy(inbounds).([]string),
					Domain:      []string{"ext:LoyalsoldierSite.dat:greatfire"},
				})
		} else {
			t.Routing.Rules = append(t.Routing.Rules,
				coreObj.RoutingRule{
					Type:        "field",
					OutboundTag: firstOutboundTag,
					InboundTag:  deepcopy.Copy(inbounds).([]string),
					Domain:      []string{"geosite:geolocation-!cn"},
				})
		}

		t.Routing.Rules = append(t.Routing.Rules,
			coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: firstOutboundTag,
				InboundTag:  deepcopy.Copy(inbounds).([]string),
				// From: https://github.com/Loyalsoldier/geoip/blob/release/text/telegram.txt
				IP: []string{"91.105.192.0/23", "91.108.4.0/22", "91.108.8.0/21", "91.108.16.0/21", "91.108.56.0/22",
					"95.161.64.0/20", "149.154.160.0/20", "185.76.151.0/24", "2001:67c:4e8::/48", "2001:b28:f23c::/47",
					"2001:b28:f23f::/48", "2a0a:f280:203::/48"},
			},
			coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				InboundTag:  deepcopy.Copy(inbounds).([]string),
			},
		)
	case configure.RoutingAMode:
		if err := parseRoutingA(t, deepcopy.Copy(inbounds).([]string)); err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) setRulePortRouting() error {
	// append rule-http and rule-socks no mather if they are enabled
	// because "The Same as the Rule Port" may need them
	return t.AppendRoutingRuleByMode(t.Setting.RulePortMode, []string{"rule-http", "rule-socks"})
}
func parseRoutingA(t *Template, routingInboundTags []string) error {
	lines := strings.Split(configure.GetRoutingA(), "\n")
	hardcodeReplacement := regexp.MustCompile(`\$\$.+?\$\$`)
	for i := range lines {
		hardcodes := hardcodeReplacement.FindAllString(lines[i], -1)
		for _, hardcode := range hardcodes {
			env := strings.TrimSuffix(strings.TrimPrefix(hardcode, "$$"), "$$")
			val, ok := os.LookupEnv(env)
			if !ok {
				log.Error("RoutingA: Environment variable \"%v\" is not found", env)
			} else {
				log.Info("RoutingA: Environment variable %v=%v", env, strconv.Quote(val))
			}
			lines[i] = strings.Replace(lines[i], hardcode, val, 1)
		}
	}
	rules, err := RoutingA.Parse(strings.Join(lines, "\n"))
	if err != nil {
		log.Warn("parseRoutingA: %v", err)
		return err
	}
	defaultOutbound, _ := t.FirstProxyOutboundName(nil)
	for _, rule := range rules {
		switch rule := rule.(type) {
		case RoutingA.Define:
			switch rule.Name {
			case "inbound", "outbound":
				switch o := rule.Value.(type) {
				case RoutingA.Bound:
					proto := o.Value
					switch proto.Name {
					case "http", "socks":
						if len(proto.NamedParams["address"]) < 1 ||
							len(proto.NamedParams["port"]) < 1 {
							continue
						}
						port, err := strconv.Atoi(proto.NamedParams["port"][0])
						if err != nil {
							continue
						}
						server := coreObj.Server{
							Port:    port,
							Address: proto.NamedParams["address"][0],
						}
						if unames := proto.NamedParams["user"]; len(unames) > 0 {
							passwords := proto.NamedParams["pass"]
							levels := proto.NamedParams["level"]
							for i, uname := range unames {
								u := coreObj.OutboundUser{
									User: uname,
								}
								if i < len(passwords) {
									u.Pass = passwords[i]
								}
								if i < len(levels) {
									level, err := strconv.Atoi(levels[i])
									if err == nil {
										u.Level = level
									}
								}
								server.Users = append(server.Users, u)
							}
						}
						switch rule.Name {
						case "outbound":
							t.Outbounds = append(t.Outbounds, coreObj.OutboundObject{
								Tag:      o.Name,
								Protocol: o.Value.Name,
								Settings: coreObj.Settings{
									Servers: []coreObj.Server{
										server,
									},
								},
							})
						case "inbound":
							// DEPRECATED: inbound definition in RoutingA is no longer supported.
							// Log a warning and skip creating the inbound.
							log.Warn("parseRoutingA: inbound definition '%s' is deprecated and will be ignored. "+
								"Use custom inbound settings with RoutingA rules instead.", o.Name)
						}
					case "freedom":
						settings := coreObj.Settings{}
						if len(proto.NamedParams["domainStrategy"]) > 0 {
							settings.DomainStrategy = proto.NamedParams["domainStrategy"][0]
						}
						if len(proto.NamedParams["redirect"]) > 0 {
							settings.Redirect = proto.NamedParams["redirect"][0]
						}
						if len(proto.NamedParams["userLevel"]) > 0 {
							level, err := strconv.Atoi(proto.NamedParams["userLevel"][0])
							if err == nil {
								settings.UserLevel = &level
							}
						}
						t.Outbounds = append(t.Outbounds, coreObj.OutboundObject{
							Tag:      o.Name,
							Protocol: o.Value.Name,
							Settings: settings,
						})
					}
				}
			}
		}
	}
	outboundTags := t.outNames()
	for _, rule := range rules {
		switch rule := rule.(type) {
		case RoutingA.Define:
			switch rule.Name {
			case "default":
				switch v := rule.Value.(type) {
				case string:
					defaultOutbound = v
					if _, ok := outboundTags[v]; !ok {
						return fmt.Errorf(`your RoutingA rules depend on the outbound "%v", thus you should select at least one server in this outbound`, v)
					}
				}
			}
		case RoutingA.Routing:
			rr := deepcopy.Copy(coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: rule.Out,
				InboundTag:  routingInboundTags,
			}).(coreObj.RoutingRule)
			for _, f := range rule.And {
				switch f.Name {
				case "domain", "domains":
					for k, vv := range f.NamedParams {
						for _, v := range vv {
							if k == "contains" {
								rr.Domain = append(rr.Domain, v)
								continue
							}
							if k == "ext" {
								datFilenameAndTag := strings.SplitN(v, ":", 2)
								if len(datFilenameAndTag) < 2 {
									return fmt.Errorf("%v: tag is not given", v)
								}
								if !asset.DoesV2rayAssetExist(datFilenameAndTag[0]) {
									return fmt.Errorf("%v: file is not found", datFilenameAndTag[0])
								}
							}
							rr.Domain = append(rr.Domain, fmt.Sprintf("%v:%v", k, v))
						}
					}
					// unnamed param is not recommended
					rr.Domain = append(rr.Domain, f.Params...)
				case "ip":
					for k, vv := range f.NamedParams {
						for _, v := range vv {
							if k == "ext" {
								datFilenameAndTag := strings.SplitN(v, ":", 2)
								if len(datFilenameAndTag) < 2 {
									return fmt.Errorf("%v: tag is not given", v)
								}
								if !asset.DoesV2rayAssetExist(datFilenameAndTag[0]) {
									return fmt.Errorf("%v: file is not found", datFilenameAndTag[0])
								}
							}
							rr.IP = append(rr.IP, fmt.Sprintf("%v:%v", k, v))
						}
					}
					rr.IP = append(rr.IP, f.Params...)
				case "network":
					rr.Network = strings.Join(f.Params, ",")
				case "port":
					rr.Port = strings.Join(f.Params, ",")
				case "sourcePort":
					rr.SourcePort = strings.Join(f.Params, ",")
				case "protocol":
					rr.Protocol = f.Params
				case "source":
					rr.Source = f.Params
				case "user":
					rr.User = f.Params
				case "inboundTag":
					rr.InboundTag = f.Params
				}
			}
			if rr.OutboundTag != "" {
				if _, ok := outboundTags[rr.OutboundTag]; !ok {
					return fmt.Errorf(`your RoutingA rules depend on the outbound "%v", thus you should select at least one server in this outbound`, rr.OutboundTag)
				}
			}
			t.Routing.Rules = append(t.Routing.Rules, rr)
		}
	}
	t.Routing.Rules = append(t.Routing.Rules, coreObj.RoutingRule{
		Type:        "field",
		OutboundTag: defaultOutbound,
		InboundTag:  []string{"rule-http", "rule-socks"},
	})
	return nil
}

// parseRoutingARules parses RoutingA rule text and returns routing rules
// filtered by the given inboundTag. This is used for custom inbound RoutingA rules.
// It does NOT create any inbounds or outbounds - only routing rules.
func parseRoutingARules(rulesText string, inboundTag string) (rules []coreObj.RoutingRule, err error) {
	lines := strings.Split(rulesText, "\n")
	hardcodeReplacement := regexp.MustCompile(`\$\$.+?\$\$`)
	for i := range lines {
		hardcodes := hardcodeReplacement.FindAllString(lines[i], -1)
		for _, hardcode := range hardcodes {
			env := strings.TrimSuffix(strings.TrimPrefix(hardcode, "$$"), "$$")
			val, ok := os.LookupEnv(env)
			if !ok {
				log.Error("parseRoutingARules: Environment variable \"%v\" is not found", env)
			} else {
				log.Info("parseRoutingARules: Environment variable %v=%v", env, strconv.Quote(val))
			}
			lines[i] = strings.Replace(lines[i], hardcode, val, 1)
		}
	}
	parsedRules, err := RoutingA.Parse(strings.Join(lines, "\n"))
	if err != nil {
		return nil, err
	}
	for _, rule := range parsedRules {
		switch rule := rule.(type) {
		case RoutingA.Routing:
			rr := deepcopy.Copy(coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: rule.Out,
				InboundTag:  []string{inboundTag},
			}).(coreObj.RoutingRule)
			for _, f := range rule.And {
				switch f.Name {
				case "domain", "domains":
					for k, vv := range f.NamedParams {
						for _, v := range vv {
							if k == "contains" {
								rr.Domain = append(rr.Domain, v)
								continue
							}
							if k == "ext" {
								datFilenameAndTag := strings.SplitN(v, ":", 2)
								if len(datFilenameAndTag) < 2 {
									return nil, fmt.Errorf("%v: tag is not given", v)
								}
								if !asset.DoesV2rayAssetExist(datFilenameAndTag[0]) {
									return nil, fmt.Errorf("%v: file is not found", datFilenameAndTag[0])
								}
							}
							rr.Domain = append(rr.Domain, fmt.Sprintf("%v:%v", k, v))
						}
					}
					rr.Domain = append(rr.Domain, f.Params...)
				case "ip":
					for k, vv := range f.NamedParams {
						for _, v := range vv {
							if k == "ext" {
								datFilenameAndTag := strings.SplitN(v, ":", 2)
								if len(datFilenameAndTag) < 2 {
									return nil, fmt.Errorf("%v: tag is not given", v)
								}
								if !asset.DoesV2rayAssetExist(datFilenameAndTag[0]) {
									return nil, fmt.Errorf("%v: file is not found", datFilenameAndTag[0])
								}
							}
							rr.IP = append(rr.IP, fmt.Sprintf("%v:%v", k, v))
						}
					}
					rr.IP = append(rr.IP, f.Params...)
				case "network":
					rr.Network = strings.Join(f.Params, ",")
				case "port":
					rr.Port = strings.Join(f.Params, ",")
				case "sourcePort":
					rr.SourcePort = strings.Join(f.Params, ",")
				case "protocol":
					rr.Protocol = f.Params
				case "source":
					rr.Source = f.Params
				case "user":
					rr.User = f.Params
				case "inboundTag":
					rr.InboundTag = f.Params
				}
			}
			if rr.OutboundTag != "" {
				rules = append(rules, rr)
			}
		}
	}
	return rules, nil
}

func (t *Template) setTransparentRouting() (err error) {
	defaultOutbound, _ := t.FirstProxyOutboundName(nil)
	switch t.Setting.Transparent {
	case configure.TransparentProxy:
		// Global transparent: route all transparent inbound to default outbound
		t.Routing.Rules = append(t.Routing.Rules, coreObj.RoutingRule{
			Type:        "field",
			InboundTag:  []string{"transparent"},
			OutboundTag: defaultOutbound,
		})
	case configure.TransparentWhitelist:
		return t.AppendRoutingRuleByMode(configure.WhitelistMode, []string{"transparent"})
	case configure.TransparentGfwlist:
		return t.AppendRoutingRuleByMode(configure.GfwlistMode, []string{"transparent"})
	case configure.TransparentFollowRule:
		// transparent mode is the same as rule
		for i := range t.Routing.Rules {
			ok := false
			for _, in := range t.Routing.Rules[i].InboundTag {
				if in == "rule-http" {
					ok = true
					break
				}
			}
			if ok {
				t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, "transparent")
			}
		}
	}
	return nil
}
func (t *Template) AppendDokodemoTProxy(tproxy string, port int, tag string) {
	dokodemo := coreObj.Inbound{
		Listen:   "127.0.0.1",
		Port:     port,
		Protocol: "dokodemo-door",
		Sniffing: coreObj.Sniffing{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
		Settings: &coreObj.InboundSettings{Network: "tcp,udp"},
		Tag:      tag,
	}
	dokodemo.StreamSettings = &coreObj.StreamSettings{Sockopt: &coreObj.Sockopt{Tproxy: &tproxy}}
	dokodemo.Settings.FollowRedirect = true
	t.Inbounds = append(t.Inbounds, dokodemo)
}

func (t *Template) updatePrivateRouting() {
	privateAddrs, _ := iptables.GetLocalCIDR()
	if len(privateAddrs) == 0 {
		return
	}
	for i := range t.Routing.Rules {
		for j := range t.Routing.Rules[i].IP {
			if t.Routing.Rules[i].IP[j] == "geoip:private" {
				t.Routing.Rules[i].IP = append(t.Routing.Rules[i].IP, privateAddrs...)
				break
			}
		}
	}
}

func (t *Template) optimizeGeoipMemoryOccupation() {
	if asset.DoesV2rayAssetExist("geoip-only-cn-private.dat") {
		for i := range t.Routing.Rules {
			for j := range t.Routing.Rules[i].IP {
				switch t.Routing.Rules[i].IP[j] {
				case "geoip:private", "geoip:cn":
					t.Routing.Rules[i].IP[j] = "ext:geoip-only-cn-private.dat:" + strings.TrimPrefix(t.Routing.Rules[i].IP[j], "geoip:")
				}
			}
		}
	}
}

func (t *Template) setWhitelistRouting(whitelist []Addr) {
	var rules []coreObj.RoutingRule
	for _, addr := range whitelist {
		if net.ParseIP(addr.host) != nil {
			rules = append(rules, coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				IP:          []string{addr.host},
				Port:        addr.port,
			})
		} else {
			rules = append(rules, coreObj.RoutingRule{
				Type:        "field",
				OutboundTag: "direct",
				Domain:      []string{addr.host},
				Port:        addr.port,
			})
		}
	}
	if len(rules) > 0 {
		t.Routing.Rules = append(rules, t.Routing.Rules...)
	}
}

func (t *Template) setGroupRouting() {
	outbounds := t.outNames()
	for i := range t.Routing.Rules {
		if t.Routing.Rules[i].OutboundTag != "" && outbounds[t.Routing.Rules[i].OutboundTag] == true {
			t.Routing.Rules[i].BalancerTag, t.Routing.Rules[i].OutboundTag = t.Routing.Rules[i].OutboundTag, ""
		}
	}
}

func (t *Template) setVmessInboundRouting() {
	if configure.GetPortsNotNil().Vmess <= 0 {
		return
	}
	for i := range t.Routing.Rules {
		var bHasRule bool
		for _, tag := range t.Routing.Rules[i].InboundTag {
			if tag == "rule-http" {
				bHasRule = true
			}
		}
		if bHasRule {
			t.Routing.Rules[i].InboundTag = append(t.Routing.Rules[i].InboundTag, "vmess")
		}
	}
}
