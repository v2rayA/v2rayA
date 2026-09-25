package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// ConfigurationMu serializes scheduled updates with API operations that change the core or database.
var ConfigurationMu sync.Mutex

var ErrNoReachableSubscriptionServer = errors.New("no reachable server")

type subscriptionProbeResult struct {
	latency time.Duration
	err     error
}

type subscriptionProber func([]serverObj.ServerObj, string) []subscriptionProbeResult

func subscriptionOwnsProxy(index int) bool {
	connected := configure.GetConnectedServersByOutbound("proxy").Get()
	if len(connected) > 0 {
		return connected[0].TYPE == configure.SubscriptionServerType && connected[0].Sub == index
	}
	for i, sub := range configure.GetSubscriptions() {
		if sub.AutoSelect {
			return i == index
		}
	}
	return false
}

func updateSubscriptionWithProbe(index int, old *configure.SubscriptionRaw, servers []serverObj.ServerObj, info string, probe subscriptionProber) error {
	return applySubscriptionSelection(index, old, servers, info, policyProbe(old.PreferFirst, probe), false)
}

// First-entry mode probes only its allowed endpoint. Other entries are never
// fallback candidates, even when monitoring is enabled.
func policyProbe(first bool, probe subscriptionProber) subscriptionProber {
	return func(servers []serverObj.ServerObj, url string) []subscriptionProbeResult {
		if !first || len(servers) == 0 {
			return probe(servers, url)
		}
		checked := probe(servers[:1], url)
		if len(checked) != 1 {
			return nil
		}
		results := make([]subscriptionProbeResult, len(servers))
		for i := range results {
			results[i].err = fmt.Errorf("not checked in first-entry mode")
		}
		results[0] = checked[0]
		return results
	}
}

func applySubscriptionSelection(index int, old *configure.SubscriptionRaw, servers []serverObj.ServerObj, info string, probe subscriptionProber, restart bool) error {
	probeURL := configure.GetOutboundSetting("proxy").ProbeURL
	if probeURL == "" {
		probeURL = HttpTestURL
	}
	results := probe(servers, probeURL)
	if len(results) != len(servers) {
		return fmt.Errorf("incomplete subscription probe results; previous configuration kept")
	}
	best := -1
	next := *old
	next.Servers = make([]configure.ServerRaw, len(servers))
	for i, result := range results {
		next.Servers[i].ServerObj = servers[i]
		if old.PreferFirst && i > 0 {
			continue
		}
		if result.err != nil {
			next.Servers[i].Latency = "UNAVAILABLE"
			log.Debug("[AutoSelect] Subscription %d, server %d: %v", index+1, i+1, result.err)
			continue
		}
		next.Servers[i].Latency = fmt.Sprintf("%dms", result.latency.Milliseconds())
		if best < 0 || result.latency < results[best].latency {
			best = i
		}
	}
	if old.PreferFirst && len(servers) > 0 {
		best = 0
	}
	if best < 0 {
		return fmt.Errorf("%w in subscription %d; previous configuration kept", ErrNoReachableSubscriptionServer, index+1)
	}
	previous := configure.GetConnectedServers()
	updated := configure.NewWhiches(nil)
	unchanged := false
	for _, w := range previous.Get() {
		if w.Outbound == "proxy" {
			if w.TYPE == configure.SubscriptionServerType && w.Sub == index && w.ID > 0 && w.ID <= len(old.Servers) {
				unchanged = old.Servers[w.ID-1].ServerObj.ExportToURL() == servers[best].ExportToURL()
			}
			continue
		}
		copy := *w
		if copy.TYPE == configure.SubscriptionServerType && copy.Sub == index {
			if copy.ID <= 0 || copy.ID > len(old.Servers) {
				return fmt.Errorf("invalid connected server reference")
			}
			raw := old.Servers[copy.ID-1]
			copy.ID = 0
			for i, s := range next.Servers {
				if s.ServerObj.ExportToURL() == raw.ServerObj.ExportToURL() {
					copy.ID = i + 1
					break
				}
			}
			if copy.ID == 0 {
				next.Servers = append(next.Servers, raw)
				copy.ID = len(next.Servers)
			}
		}
		updated.Add(copy)
	}
	updated.Add(configure.Which{TYPE: configure.SubscriptionServerType, Sub: index, ID: best + 1, Outbound: "proxy"})
	next.Status = string(touch.NewUpdateStatus())
	next.Info = info
	wasRunning := v2ray.ProcessManager.Running() || (restart && configure.GetRunning())
	if err := configure.SetSubscriptionAndConnects(index, &next, updated); err != nil {
		return err
	}
	if wasRunning && (!unchanged || restart) {
		if err := v2ray.UpdateV2RayConfig(); err != nil {
			restoreErr := configure.SetSubscriptionAndConnects(index, old, previous)
			if restoreErr == nil {
				restoreErr = v2ray.UpdateV2RayConfig()
			}
			if restart && restoreErr != nil {
				_ = configure.SetRunning(true)
			}
			return errors.Join(fmt.Errorf("failed to apply selected server: %w", err), restoreErr)
		}
	}
	if restart && old.PreferFirst && results[0].err != nil {
		return fmt.Errorf("%w: first entry remains unavailable", ErrNoReachableSubscriptionServer)
	}
	log.Info("[AutoSelect] Subscription %d: selected server %d (%s), checked %d servers", index+1, best+1, next.Servers[best].Latency, len(servers))
	return nil
}
