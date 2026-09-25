package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var automationCancelMu sync.Mutex
var automationCancel context.CancelFunc

// A user mutation cancels network work before taking ConfigurationMu. No
// network request or candidate probe holds that mutex.
func CancelAutomation() {
	automationCancelMu.Lock()
	defer automationCancelMu.Unlock()
	if automationCancel != nil {
		automationCancel()
	}
}

func ValidateOutboundSetting(s configure.OutboundSetting) error {
	u, err := url.Parse(s.ProbeURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("probe URL must be HTTP or HTTPS")
	}
	d, err := time.ParseDuration(s.ProbeInterval)
	if err != nil || d < time.Second || d > 365*24*time.Hour {
		return fmt.Errorf("probe interval must be between 1 second and 365 days")
	}
	if s.Type != configure.LeastPing {
		return fmt.Errorf("unsupported group type")
	}
	return nil
}

type automationSnapshot struct {
	Subscriptions []configure.SubscriptionRaw
	Servers       []configure.ServerRaw
	Groups        map[string]configure.OutboundSetting
}

func readAutomationSnapshot() automationSnapshot {
	s := automationSnapshot{Subscriptions: configure.GetSubscriptions(), Servers: configure.GetServers(), Groups: map[string]configure.OutboundSetting{}}
	for _, name := range configure.GetOutbounds() {
		s.Groups[name] = configure.GetOutboundSetting(name)
	}
	return s
}

func fingerprint(v interface{}) string { b, _ := json.Marshal(v); return string(b) }

type subscriptionSchedule struct {
	snapshot                            string
	nextRegular, nextFailure, nextCheck time.Time
}

func (s *subscriptionSchedule) due(now time.Time) (regular, failure bool) {
	return !s.nextRegular.IsZero() && !now.Before(s.nextRegular), !s.nextFailure.IsZero() && !now.Before(s.nextFailure)
}

func (s *subscriptionSchedule) health(now time.Time, healthy bool, minutes int) {
	s.nextCheck = now.Add(300 * time.Second)
	if healthy {
		s.nextFailure = time.Time{}
	} else if s.nextFailure.IsZero() {
		s.nextFailure = now.Add(time.Duration(minutes) * time.Minute)
	}
}

type groupSchedule struct {
	snapshot string
	next     time.Time
}
type automation struct {
	subscriptions map[string]*subscriptionSchedule
	groups        map[string]groupSchedule
	now           func() time.Time
	probe         func(context.Context, []serverObj.ServerObj, string) []subscriptionProbeResult
}

func newAutomation() *automation {
	return &automation{map[string]*subscriptionSchedule{}, map[string]groupSchedule{}, time.Now, probeSubscriptionWithContext}
}

// step is called by one worker. Probes use at most two temporary cores and
// share results for identical candidates and URLs within this pass.
func (a *automation) step(parent context.Context) {
	ConfigurationMu.Lock()
	ctx, cancel := context.WithCancel(parent)
	automationCancelMu.Lock()
	automationCancel = cancel
	automationCancelMu.Unlock()
	snapshot := readAutomationSnapshot()
	ConfigurationMu.Unlock()
	defer func() { cancel(); automationCancelMu.Lock(); automationCancel = nil; automationCancelMu.Unlock() }()

	// Apply only against the exact catalog/settings that the job observed.
	apply := func(f func() error) error {
		ConfigurationMu.Lock()
		defer ConfigurationMu.Unlock()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !reflect.DeepEqual(snapshot, readAutomationSnapshot()) {
			cancel()
			return fmt.Errorf("catalog changed during automation")
		}
		err := f()
		snapshot = readAutomationSnapshot()
		return err
	}
	cache := map[string]subscriptionProbeResult{}
	probe := func(nodes []serverObj.ServerObj, probeURL string) ([]subscriptionProbeResult, error) {
		var missing []serverObj.ServerObj
		keys := make([]string, len(nodes))
		seen := map[string]bool{}
		for i, node := range nodes {
			keys[i] = probeURL + "\n" + fingerprint(node)
			if _, ok := cache[keys[i]]; !ok && !seen[keys[i]] {
				missing = append(missing, node)
				seen[keys[i]] = true
			}
		}
		var results []subscriptionProbeResult
		if len(missing) > 0 {
			results = a.probe(ctx, missing, probeURL)
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if len(results) != len(missing) {
			return nil, fmt.Errorf("incomplete reachability results")
		}
		for i, node := range missing {
			cache[probeURL+"\n"+fingerprint(node)] = results[i]
		}
		results = make([]subscriptionProbeResult, len(nodes))
		for i, key := range keys {
			results[i] = cache[key]
		}
		return results, nil
	}
	probeURL := snapshot.Groups["proxy"].ProbeURL
	if probeURL == "" {
		probeURL = HttpTestURL
	}
	activeSubs := map[string]bool{}
	for index := range snapshot.Subscriptions {
		sub := &snapshot.Subscriptions[index]
		if !sub.AutoUpdate {
			continue
		}
		key := fmt.Sprintf("%d\n%s", index, sub.Address)
		activeSubs[key] = true
		state := a.subscriptions[key]
		signature := fingerprint(sub) + probeURL
		if state == nil || state.snapshot != signature {
			state = &subscriptionSchedule{snapshot: signature}
			if sub.UpdateIntervalMinutes > 0 {
				state.nextRegular = a.now().Add(time.Duration(sub.UpdateIntervalMinutes) * time.Minute)
			}
			a.subscriptions[key] = state
		}
		regular, failure := state.due(a.now())
		if regular || failure {
			ConfigurationMu.Lock()
			client := subscriptionHTTPClient()
			ConfigurationMu.Unlock()
			nodes, info, err := resolveSubscriptionWithContext(ctx, sub.Address, client)
			if ctx.Err() != nil {
				return
			}
			if err == nil {
				err = apply(func() error { return storeSubscriptionUpdate(index, sub, nodes, info, false) })
				if ctx.Err() != nil {
					return
				}
				sub = &snapshot.Subscriptions[index]
				if err == nil {
					v2ray.ApiFeed.ProductMessage("catalog_changed", nil)
				}
			}
			if err != nil {
				log.Warn("[Subscriptions] Update %d failed; saved candidates retained: %v", index+1, err)
			}
			if regular {
				state.nextRegular = a.now().Add(time.Duration(sub.UpdateIntervalMinutes) * time.Minute)
			}
			// A failed fetch is bounded by the same retry interval; it never
			// replaces the catalog with an empty or malformed response.
			if failure {
				state.nextFailure = a.now().Add(time.Duration(max(1, sub.FailureIntervalMinutes)) * time.Minute)
			}
			state.snapshot = fingerprint(sub) + probeURL
		}
		if regular || failure || !a.now().Before(state.nextCheck) {
			nodes := make([]serverObj.ServerObj, len(sub.Servers))
			for i := range sub.Servers {
				nodes[i] = sub.Servers[i].ServerObj
			}
			results, err := probe(nodes, probeURL)
			if err != nil {
				return
			}
			if err := apply(func() error { return nil }); err != nil {
				return
			}
			healthy := false
			for _, result := range results {
				healthy = healthy || result.err == nil
			}
			state.health(a.now(), healthy, max(1, sub.FailureIntervalMinutes))
		}
	}
	for key := range a.subscriptions {
		if !activeSubs[key] {
			delete(a.subscriptions, key)
		}
	}
	activeGroups := map[string]bool{}
	for name, setting := range snapshot.Groups {
		if !setting.AutoAdd {
			continue
		}
		activeGroups[name] = true
		if err := ValidateOutboundSetting(setting); err != nil {
			continue
		}
		signature := fingerprint(struct {
			S interface{}
			C interface{}
			G interface{}
		}{snapshot.Subscriptions, snapshot.Servers, setting})
		state := a.groups[name]
		if state.snapshot == signature && a.now().Before(state.next) {
			continue
		}
		var refs []configure.NodeRef
		var nodes []serverObj.ServerObj
		for i, raw := range snapshot.Servers {
			refs = append(refs, configure.NodeRef{TYPE: configure.ServerType, ID: i + 1, Outbound: name})
			nodes = append(nodes, raw.ServerObj)
		}
		for i, sub := range snapshot.Subscriptions {
			for j, raw := range sub.Servers {
				refs = append(refs, configure.NodeRef{TYPE: configure.SubscriptionServerType, Sub: i, ID: j + 1, Outbound: name})
				nodes = append(nodes, raw.ServerObj)
			}
		}
		results, err := probe(nodes, setting.ProbeURL)
		if err != nil {
			return
		}
		members := make([]configure.NodeRef, 0, len(refs))
		for i, result := range results {
			if result.err == nil {
				members = append(members, refs[i])
			}
		}
		err = apply(func() error {
			current := configure.GetConnectedServersByOutbound(name).Get()
			if sameMembers(current, members) {
				return nil
			}
			if err := ReplaceOutboundConnections(name, members); err != nil {
				return err
			}
			log.Info("[Groups] %s: %d/%d candidates available", name, len(members), len(nodes))
			v2ray.ApiFeed.ProductMessage("catalog_changed", nil)
			return nil
		})
		if ctx.Err() != nil {
			return
		}
		interval, _ := time.ParseDuration(setting.ProbeInterval)
		if err != nil {
			log.Warn("[Groups] %s: could not apply membership: %v", name, err)
			interval = max(interval, 30*time.Second)
		}
		a.groups[name] = groupSchedule{signature, a.now().Add(interval)}
	}
	for name := range a.groups {
		if !activeGroups[name] {
			delete(a.groups, name)
		}
	}
}

func sameMembers(current []*configure.NodeRef, desired []configure.NodeRef) bool {
	if len(current) != len(desired) {
		return false
	}
	seen := map[configure.NodeRef]bool{}
	for _, ref := range current {
		seen[*ref] = true
	}
	for _, ref := range desired {
		if !seen[ref] {
			return false
		}
	}
	return true
}

func StartAutomation() func() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker := newAutomation()
		for {
			worker.step(ctx)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
	return func() { cancel(); <-done }
}
