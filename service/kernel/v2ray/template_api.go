package v2ray

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/go-leo/slicex"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func (t *Template) SetAPI(serverData *ServerData) (port int, err error) {
	// find a valid port
	config := configure.GetPortsNotNil()
	if config.Api.Port != 0 {
		port = config.Api.Port
	} else {
		for {
			if l, err := net.Listen("tcp4", "127.0.0.1:0"); err == nil {
				port = l.Addr().(*net.TCPAddr).Port
				_ = l.Close()
				break
			}
			time.Sleep(30 * time.Millisecond)
		}
	}
	services := []string{
		"LoggerService",
	}
	services = slicex.Uniq(append(services, config.Api.Services...))
	// observatory
	if serverData != nil {
		outbounds := t.outNames()
		for outbound, isGroup := range outbounds {
			if !isGroup {
				continue
			}

			//TODO: random, leastload
			strategy := serverData.OutboundName2Setting[outbound].Type
			interval, err := time.ParseDuration(serverData.OutboundName2Setting[outbound].ProbeInterval)
			if err != nil {
				log.Warn("observatory: %v", err)
				interval = 10 * time.Second
			}
			var selector []string

			for _, vi := range serverData.OutboundName2ServerObjs[outbound] {
				selector = append(selector, GroupWrapper(vi.GetName()))
			}

			t.Routing.Balancers = append(t.Routing.Balancers, coreObj.Balancer{
				Tag:      outbound,
				Selector: selector,
				Strategy: coreObj.BalancerStrategy{
					Type: strategy.String(),
					Settings: &coreObj.StrategySettings{
						ObserverTag: outbound,
					},
				},
			})

			if strings.ToLower(strategy.String()) == "leastping" {
				probeUrl := serverData.OutboundName2Setting[outbound].ProbeURL
				if _, err := url.Parse(probeUrl); err != nil {
					log.Warn("observatory: %v", err)
					probeUrl = "https://gstatic.com/generate_204"
				}

				// v2raya_core always uses MultiObservatory: one observer per balancer group.
				if t.MultiObservatory == nil {
					t.MultiObservatory = &coreObj.MultiObservatory{}
				}
				t.MultiObservatory.Observers = append(t.MultiObservatory.Observers, coreObj.ObservatoryItem{
					Tag: outbound,
					Settings: coreObj.Observatory{
						SubjectSelector: selector,
						PingConfig: &coreObj.PingConfig{
							Destination: probeUrl,
							Interval:    interval.String(),
						},
						// Keep legacy fields for backward compatibility with older custom cores.
						ProbeURL:      probeUrl,
						ProbeInterval: interval.String(),
					},
				})
			}
		}
		if t.MultiObservatory != nil || t.Observatory != nil {
			// v2raya_core supports ObservatoryService via the v2ray-compat gRPC path.
			if t.Variant == where.V2rayaCore {

				var observatoryTags []string
				for name, isGroup := range t.outNames() {
					if isGroup {
						observatoryTags = append(observatoryTags, name)
					}
				}
				t.ApiCloses = append(t.ApiCloses, ObservatoryProducer(port, observatoryTags))
			}
		}
	}
	t.API = &coreObj.APIObject{
		Tag:      "api-out",
		Services: services,
	}

	t.Inbounds = append(t.Inbounds, coreObj.Inbound{
		Port:     port,
		Protocol: "dokodemo-door",
		Listen:   "127.0.0.1",
		Settings: &coreObj.InboundSettings{
			Address: "127.0.0.1",
		},
		Tag: "api-in",
	})
	t.Routing.Rules = append(t.Routing.Rules, coreObj.RoutingRule{
		Type:        "field",
		InboundTag:  []string{"api-in"},
		OutboundTag: "api-out",
	})
	t.ApiPort = port
	return port, nil
}
