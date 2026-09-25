package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var (
	automationCancelMu sync.Mutex
	automationCancel   context.CancelFunc
	automationWake     = make(chan struct{}, 1)
)

func CancelAutomation() {
	automationCancelMu.Lock()
	defer automationCancelMu.Unlock()
	if automationCancel != nil {
		automationCancel()
	}
}

func NotifyAutomation() {
	select {
	case automationWake <- struct{}{}:
	default:
	}
}

type subscriptionSchedule struct {
	signature                         string
	nextUpdate, nextHealth, nextRetry time.Time
}

type groupSchedule struct {
	signature string
	next      time.Time
}

type groupCandidate struct {
	ref      configure.NodeRef
	node     serverObj.ServerObj
	identity string
}

type subscriptionFetcher func(context.Context, string) ([]serverObj.ServerObj, string, error)

type automation struct {
	subscriptions map[int64]*subscriptionSchedule
	groups        map[string]*groupSchedule
	now           func() time.Time
	probe         func(context.Context, []serverObj.ServerObj, string) []subscriptionProbeResult
	fetch         subscriptionFetcher
	applyGroup    func(string, []configure.NodeRef) error
}

func newAutomation() *automation {
	return &automation{
		subscriptions: map[int64]*subscriptionSchedule{},
		groups:        map[string]*groupSchedule{},
		now:           time.Now,
		probe:         probeSubscriptionWithContext,
		fetch:         fetchSubscriptionForAutomation,
		applyGroup:    replaceManagedOutboundConnections,
	}
}

func ValidateOutboundSetting(setting configure.OutboundSetting) error {
	probeURL, err := url.Parse(setting.ProbeURL)
	if err != nil || probeURL.Host == "" || (probeURL.Scheme != "http" && probeURL.Scheme != "https") {
		return fmt.Errorf("probe URL must be HTTP or HTTPS")
	}
	interval, err := time.ParseDuration(setting.ProbeInterval)
	if err != nil || interval < time.Second || interval > 365*24*time.Hour {
		return fmt.Errorf("probe interval must be between 1 second and 365 days")
	}
	if setting.Type != configure.LeastPing {
		return fmt.Errorf("unsupported group type")
	}
	return nil
}

func subscriptionPolicySignature(sub *configure.SubscriptionRaw) string {
	b, _ := json.Marshal(struct {
		ID       int64
		Address  string
		Mode     configure.SubscriptionUpdateMode
		Regular  int
		Failsafe int
	}{sub.DatabaseID, sub.Address, sub.UpdateMode, sub.UpdateIntervalMinutes, sub.FailureIntervalMinutes})
	return string(b)
}

func fetchSubscriptionForAutomation(ctx context.Context, address string) ([]serverObj.ServerObj, string, error) {
	return resolveSubscriptionWithContext(ctx, address, subscriptionHTTPClient())
}

func findSubscriptionByDatabaseID(id int64) (int, *configure.SubscriptionRaw) {
	subscriptions := configure.GetSubscriptions()
	for index := range subscriptions {
		if subscriptions[index].DatabaseID == id {
			return index, &subscriptions[index]
		}
	}
	return -1, nil
}

func (a *automation) applyUpdate(ctx context.Context, observed *configure.SubscriptionRaw, nodes []serverObj.ServerObj, info string) (*configure.SubscriptionRaw, error) {
	ConfigurationMu.Lock()
	defer ConfigurationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	index, current := findSubscriptionByDatabaseID(observed.DatabaseID)
	if current == nil || subscriptionPolicySignature(current) != subscriptionPolicySignature(observed) {
		return nil, fmt.Errorf("subscription changed during update")
	}
	if err := storeSubscriptionUpdate(index, current, nodes, info, false); err != nil {
		return nil, err
	}
	_, updated := findSubscriptionByDatabaseID(observed.DatabaseID)
	return updated, nil
}

func subscriptionNodes(sub *configure.SubscriptionRaw) []serverObj.ServerObj {
	nodes := make([]serverObj.ServerObj, len(sub.Servers))
	for i := range sub.Servers {
		nodes[i] = sub.Servers[i].ServerObj
	}
	return nodes
}

func anyHealthy(results []subscriptionProbeResult) bool {
	for _, result := range results {
		if result.err == nil {
			return true
		}
	}
	return false
}

func isIntervalMode(mode configure.SubscriptionUpdateMode) bool {
	return mode == configure.SubscriptionUpdateAtInterval || mode == configure.SubscriptionUpdateIntervalFailsafe
}

func (a *automation) process(ctx context.Context, sub *configure.SubscriptionRaw, probeURL string) error {
	now := a.now()
	signature := subscriptionPolicySignature(sub)
	state := a.subscriptions[sub.DatabaseID]
	if state == nil || state.signature != signature {
		state = &subscriptionSchedule{signature: signature, nextUpdate: now}
		a.subscriptions[sub.DatabaseID] = state
	}

	regularDue := !state.nextUpdate.IsZero() && !now.Before(state.nextUpdate)
	retryDue := !state.nextRetry.IsZero() && !now.Before(state.nextRetry)
	healthDue := !state.nextHealth.IsZero() && !now.Before(state.nextHealth)
	if healthDue && !regularDue && !retryDue {
		results := a.probe(ctx, cloneProbeNodes(subscriptionNodes(sub)), probeURL)
		if err := ctx.Err(); err != nil {
			return err
		}
		if anyHealthy(results) {
			state.nextHealth = a.now().Add(time.Duration(sub.FailureIntervalMinutes) * time.Minute)
			return nil
		}
		retryDue = true
		state.nextHealth = time.Time{}
	}

	if !regularDue && !retryDue {
		return nil
	}
	nodes, info, err := a.fetch(ctx, sub.Address)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err == nil {
		updated, applyErr := a.applyUpdate(ctx, sub, nodes, info)
		if applyErr == nil {
			sub = updated
			v2ray.ApiFeed.ProductMessage("catalog_changed", nil)
		} else {
			err = applyErr
		}
	}
	if err != nil {
		log.Warn("[Subscriptions] Update %d failed; saved candidates retained: %v", sub.DatabaseID, err)
	}

	finished := a.now()
	if regularDue {
		state.nextUpdate = time.Time{}
		if isIntervalMode(sub.UpdateMode) {
			state.nextUpdate = finished.Add(time.Duration(sub.UpdateIntervalMinutes) * time.Minute)
		}
	}
	if sub.UpdateMode != configure.SubscriptionUpdateIntervalFailsafe {
		state.nextHealth, state.nextRetry = time.Time{}, time.Time{}
		return nil
	}
	results := a.probe(ctx, cloneProbeNodes(subscriptionNodes(sub)), probeURL)
	if err := ctx.Err(); err != nil {
		return err
	}
	state.nextHealth, state.nextRetry = time.Time{}, time.Time{}
	interval := time.Duration(sub.FailureIntervalMinutes) * time.Minute
	if anyHealthy(results) {
		state.nextHealth = a.now().Add(interval)
	} else {
		state.nextRetry = a.now().Add(interval)
	}
	return nil
}

func (a *automation) nextDeadline() time.Time {
	var next time.Time
	for _, state := range a.subscriptions {
		for _, candidate := range []time.Time{state.nextUpdate, state.nextHealth, state.nextRetry} {
			if !candidate.IsZero() && (next.IsZero() || candidate.Before(next)) {
				next = candidate
			}
		}
	}
	for _, state := range a.groups {
		if !state.next.IsZero() && (next.IsZero() || state.next.Before(next)) {
			next = state.next
		}
	}
	return next
}

func automaticGroupCandidates() []groupCandidate {
	servers := configure.GetServers()
	subscriptions := configure.GetSubscriptions()
	candidates := make([]groupCandidate, 0, len(servers))
	for index, raw := range servers {
		identity := fmt.Sprintf("server/%d/", index)
		if raw.ServerObj != nil {
			identity += raw.ServerObj.ExportToURL()
		}
		candidates = append(candidates, groupCandidate{
			ref:      configure.NodeRef{TYPE: configure.ServerType, ID: index + 1},
			node:     raw.ServerObj,
			identity: identity,
		})
	}
	for subIndex, sub := range subscriptions {
		for nodeIndex, raw := range sub.Servers {
			identity := fmt.Sprintf("subscription/%d/%d/%d/", sub.DatabaseID, subIndex, nodeIndex)
			if raw.ServerObj != nil {
				identity += raw.ServerObj.ExportToURL()
			}
			candidates = append(candidates, groupCandidate{
				ref:      configure.NodeRef{TYPE: configure.SubscriptionServerType, Sub: subIndex, ID: nodeIndex + 1},
				node:     raw.ServerObj,
				identity: identity,
			})
		}
	}
	return candidates
}

func automaticGroupSignature(setting configure.OutboundSetting, candidates []groupCandidate) string {
	identities := make([]string, len(candidates))
	for i := range candidates {
		identities[i] = candidates[i].identity
	}
	b, _ := json.Marshal(struct {
		AutoAdd       bool
		ProbeURL      string
		ProbeInterval string
		Type          configure.ObservatoryType
		Candidates    []string
	}{setting.AutoAdd, setting.ProbeURL, setting.ProbeInterval, setting.Type, identities})
	return string(b)
}

func sameMembers(current []*configure.NodeRef, desired []configure.NodeRef) bool {
	if len(current) != len(desired) {
		return false
	}
	seen := make(map[configure.NodeRef]bool, len(current))
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

func (a *automation) processAutomaticGroup(ctx context.Context, name string, setting configure.OutboundSetting, candidates []groupCandidate, probe func([]serverObj.ServerObj, string) ([]subscriptionProbeResult, error)) {
	signature := automaticGroupSignature(setting, candidates)
	state := a.groups[name]
	now := a.now()
	if state != nil && state.signature == signature && now.Before(state.next) {
		return
	}
	if state == nil {
		state = &groupSchedule{}
		a.groups[name] = state
	}
	interval, err := time.ParseDuration(setting.ProbeInterval)
	if err != nil {
		interval = 30 * time.Second
	}
	backoff := max(interval, 30*time.Second)
	state.signature = signature

	nodes := make([]serverObj.ServerObj, len(candidates))
	for i := range candidates {
		nodes[i] = candidates[i].node
	}
	results, err := probe(nodes, setting.ProbeURL)
	if err == nil {
		members := make([]configure.NodeRef, 0, len(candidates))
		for i, result := range results {
			if result.err == nil {
				ref := candidates[i].ref
				ref.Outbound = name
				members = append(members, ref)
			}
		}
		ConfigurationMu.Lock()
		if ctx.Err() != nil {
			err = ctx.Err()
		} else {
			currentSetting := configure.GetOutboundSetting(name)
			currentCandidates := automaticGroupCandidates()
			if automaticGroupSignature(currentSetting, currentCandidates) != signature {
				err = fmt.Errorf("catalog or group setting changed during membership check")
			} else if !sameMembers(configure.GetConnectedServersByOutbound(name).Get(), members) {
				err = a.applyGroup(name, members)
				if err == nil {
					log.Info("[Groups] %s: %d/%d candidates available", name, len(members), len(candidates))
					v2ray.ApiFeed.ProductMessage("catalog_changed", nil)
				}
			}
		}
		ConfigurationMu.Unlock()
	}
	if err != nil {
		state.next = a.now().Add(backoff)
		if ctx.Err() == nil {
			log.Warn("[Groups] %s: automatic membership failed: %v", name, err)
		}
		return
	}
	state.next = a.now().Add(interval)
}

// step runs due jobs serially. If a pass takes longer than its interval, a
// second pass is never started in parallel; the next deadline starts at the
// completion time.
func (a *automation) step(parent context.Context) time.Time {
	ctx, cancel := context.WithCancel(parent)
	automationCancelMu.Lock()
	automationCancel = cancel
	automationCancelMu.Unlock()
	defer func() {
		cancel()
		automationCancelMu.Lock()
		automationCancel = nil
		automationCancelMu.Unlock()
	}()

	ConfigurationMu.Lock()
	subs := configure.GetSubscriptions()
	probeURL := configure.GetOutboundSetting(configure.DefaultOutboundName).ProbeURL
	ConfigurationMu.Unlock()
	if probeURL == "" {
		probeURL = HttpTestURL
	}
	active := map[int64]bool{}
	for index := range subs {
		sub := &subs[index]
		if sub.UpdateMode == configure.SubscriptionUpdateDisabled {
			continue
		}
		active[sub.DatabaseID] = true
		if err := a.process(ctx, sub, probeURL); err != nil {
			if ctx.Err() != nil {
				return time.Time{}
			}
			log.Warn("[Subscriptions] automation failed for %d: %v", sub.DatabaseID, err)
		}
	}
	for id := range a.subscriptions {
		if !active[id] {
			delete(a.subscriptions, id)
		}
	}

	cache := map[string]subscriptionProbeResult{}
	probe := func(nodes []serverObj.ServerObj, probeURL string) ([]subscriptionProbeResult, error) {
		keys := make([]string, len(nodes))
		missingKeys := make([]string, 0, len(nodes))
		missingNodes := make([]serverObj.ServerObj, 0, len(nodes))
		seen := map[string]bool{}
		for i, node := range nodes {
			identity := fmt.Sprintf("nil/%d", i)
			if node != nil {
				identity = node.ExportToURL()
			}
			keys[i] = probeURL + "\n" + identity
			if _, ok := cache[keys[i]]; !ok && !seen[keys[i]] {
				seen[keys[i]] = true
				missingKeys = append(missingKeys, keys[i])
				missingNodes = append(missingNodes, node)
			}
		}
		if len(missingNodes) > 0 {
			results := a.probe(ctx, cloneProbeNodes(missingNodes), probeURL)
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if len(results) != len(missingNodes) {
				return nil, fmt.Errorf("incomplete reachability results")
			}
			for i := range results {
				cache[missingKeys[i]] = results[i]
			}
		}
		results := make([]subscriptionProbeResult, len(nodes))
		for i := range keys {
			results[i] = cache[keys[i]]
		}
		return results, nil
	}

	ConfigurationMu.Lock()
	candidates := automaticGroupCandidates()
	groups := make(map[string]configure.OutboundSetting)
	for _, name := range configure.GetOutbounds() {
		groups[name] = configure.GetOutboundSetting(name)
	}
	ConfigurationMu.Unlock()
	activeGroups := map[string]bool{}
	for name, setting := range groups {
		if !setting.AutoAdd {
			continue
		}
		activeGroups[name] = true
		if err := ValidateOutboundSetting(setting); err != nil {
			log.Warn("[Groups] %s: invalid automatic membership setting: %v", name, err)
			continue
		}
		a.processAutomaticGroup(ctx, name, setting, candidates, probe)
		if ctx.Err() != nil {
			return time.Time{}
		}
	}
	for name := range a.groups {
		if !activeGroups[name] {
			delete(a.groups, name)
		}
	}
	return a.nextDeadline()
}

func StartAutomation() func() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker := newAutomation()
		for {
			next := worker.step(ctx)
			if ctx.Err() != nil {
				return
			}
			var timer *time.Timer
			var timerC <-chan time.Time
			if !next.IsZero() {
				delay := time.Until(next)
				if delay < 0 {
					delay = 0
				}
				timer = time.NewTimer(delay)
				timerC = timer.C
			}
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				return
			case <-automationWake:
				if timer != nil {
					timer.Stop()
				}
			case <-timerC:
			}
		}
	}()
	return func() { cancel(); <-done }
}
