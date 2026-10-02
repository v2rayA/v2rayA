package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
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
	automationEnabled  atomic.Bool
)

func init() {
	automationEnabled.Store(true)
}

func PauseAutomation() {
	automationEnabled.Store(false)
	CancelAutomation()
	// A stopped core must not leave a temporary probe core behind.
	probeCoreSlot <- struct{}{}
	<-probeCoreSlot
}

func ResumeAutomation() {
	automationEnabled.Store(true)
	NotifyAutomation()
}

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
	// retryFromOutage records that the reserved nextRetry was armed by an
	// actual all-unavailable health result. A retry reserved before a regular
	// fetch, or re-armed after its cancellation, must not grant the direct
	// fallback: no outage was detected for it.
	retryFromOutage   bool
	warnedUnprobeable bool
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

type subscriptionFetcher func(context.Context, string, bool) ([]serverObj.ServerObj, string, error)

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
	switch setting.Type {
	case configure.LeastPing, configure.LeastLoad, configure.RoundRobin, configure.Random:
	default:
		return fmt.Errorf("unsupported group type")
	}
	return nil
}

func subscriptionPolicySignature(sub *configure.SubscriptionRaw) string {
	b, _ := json.Marshal(struct {
		ID          int64
		Address     string
		Mode        configure.SubscriptionUpdateMode
		Regular     int
		AllowDirect bool
		Failsafe    int
	}{sub.DatabaseID, sub.Address, sub.UpdateMode, sub.UpdateIntervalMinutes, sub.AllowDirectRecovery, sub.FailureIntervalMinutes})
	return string(b)
}

func fetchSubscriptionForAutomation(ctx context.Context, address string, allowDirectFallback bool) ([]serverObj.ServerObj, string, error) {
	client, err := subscriptionHTTPClient()
	var nodes []serverObj.ServerObj
	var info string
	if err == nil {
		nodes, info, err = resolveSubscriptionWithContext(ctx, address, client)
	}
	if err != nil && allowDirectFallback && ctx.Err() == nil && configure.GetSettingNotNil().ProxyModeWhenSubscribe != configure.ProxyModeDirect {
		log.Warn("[Subscriptions] Fail-safe recovery for %s: configured download route failed; retrying directly", subscriptionHost(address))
		return resolveSubscriptionWithContext(ctx, address, directSubscriptionClient())
	}
	return nodes, info, err
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

func (a *automation) subscriptionUnavailable(ctx context.Context, sub *configure.SubscriptionRaw, probeURL string, state *subscriptionSchedule) (bool, error) {
	nodes := cloneProbeNodes(subscriptionNodes(sub))
	for _, node := range nodes {
		if node == nil {
			if !state.warnedUnprobeable {
				log.Warn("[Subscriptions] Subscription %d contains candidates that cannot be probed in isolation; fail-safe recovery is suspended, regular updates remain enabled", sub.DatabaseID)
				state.warnedUnprobeable = true
			}
			// An unsupported probe is unknown health, not evidence that every
			// candidate is down. Do not repeatedly download plugin subscriptions.
			return false, ctx.Err()
		}
	}
	state.warnedUnprobeable = false
	results := a.probe(ctx, nodes, probeURL)
	return !anyHealthy(results), ctx.Err()
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
	interval := time.Duration(sub.FailureIntervalMinutes) * time.Minute
	if healthDue && !regularDue && !retryDue {
		state.nextHealth = now.Add(interval)
		unavailable, err := a.subscriptionUnavailable(ctx, sub, probeURL, state)
		if err != nil {
			state.nextHealth = a.now().Add(interval)
			return err
		}
		if !unavailable {
			state.nextHealth = a.now().Add(time.Duration(sub.FailureIntervalMinutes) * time.Minute)
			return nil
		}
		retryDue = true
		state.retryFromOutage = true
		state.nextHealth = time.Time{}
	}

	if !regularDue && !retryDue {
		return nil
	}
	if sub.UpdateMode == configure.SubscriptionUpdateIntervalFailsafe {
		// Reserve a retry before cancellable work; a dashboard mutation must
		// not consume the only deadline that can resume recovery. The
		// reservation never touches retryFromOutage: only an actual
		// all-unavailable health result arms the direct fallback, and a
		// cancelled regular fetch therefore never promotes itself into a
		// direct recovery attempt when its retry is re-armed below.
		state.nextHealth = time.Time{}
		state.nextRetry = a.now().Add(interval)
		if regularDue {
			state.nextUpdate = a.now().Add(time.Duration(sub.UpdateIntervalMinutes) * time.Minute)
		}
		defer func() {
			if ctx.Err() != nil {
				state.nextRetry = a.now().Add(interval)
				if regularDue {
					state.nextUpdate = a.now().Add(time.Duration(sub.UpdateIntervalMinutes) * time.Minute)
				}
			}
		}()
	}
	nodes, info, err := a.fetch(ctx, sub.Address, retryDue && state.retryFromOutage && sub.AllowDirectRecovery)
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
		state.retryFromOutage = false
		return nil
	}
	unavailable, err := a.subscriptionUnavailable(ctx, sub, probeURL, state)
	if err != nil {
		return err
	}
	state.nextHealth, state.nextRetry = time.Time{}, time.Time{}
	state.retryFromOutage = false
	if !unavailable {
		state.nextHealth = a.now().Add(interval)
	} else {
		state.nextRetry = a.now().Add(interval)
		state.retryFromOutage = true
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
	if !automationEnabled.Load() {
		return time.Time{}
	}
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
	if !automationEnabled.Load() {
		return time.Time{}
	}

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
				return a.nextDeadline()
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
	automationEnabled.Store(true)
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
