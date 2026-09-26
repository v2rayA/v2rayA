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
	applyState    func(string, configure.OutboundSetting, []configure.NodeRef, bool) error
}

func newAutomation() *automation {
	return &automation{
		subscriptions: map[int64]*subscriptionSchedule{},
		groups:        map[string]*groupSchedule{},
		now:           time.Now,
		probe:         probeSubscriptionWithContext,
		fetch:         fetchSubscriptionForAutomation,
		applyGroup:    replaceManagedOutboundConnections,
		applyState:    applyManagedGroupState,
	}
}

func writeGroupMembers(outbound string, members []configure.NodeRef) error {
	if len(members) == 0 {
		return configure.ClearConnects(outbound)
	}
	refs := new(configure.NodeRefs)
	for i := range members {
		member := members[i]
		member.Outbound = outbound
		refs.Add(member)
	}
	return configure.OverwriteConnects(refs)
}

// applyManagedGroupState changes the worker-owned strategy state and, for an
// automatic group, its membership before one preserved-interception reload.
func applyManagedGroupState(outbound string, next configure.OutboundSetting, members []configure.NodeRef, replaceMembers bool) error {
	previousSetting := configure.GetOutboundSetting(outbound)
	previousMembers := configure.GetConnectedServersByOutbound(outbound)
	restoreMembers := func() error {
		if previousMembers == nil || previousMembers.Len() == 0 {
			return configure.ClearConnects(outbound)
		}
		return configure.OverwriteConnects(previousMembers)
	}
	return ApplyGroupConfig(func() func() error {
		return func() error {
			if replaceMembers {
				if err := restoreMembers(); err != nil {
					return err
				}
			}
			return configure.SetOutboundSetting(outbound, previousSetting)
		}
	}, func() error {
		if replaceMembers {
			if err := writeGroupMembers(outbound, members); err != nil {
				return err
			}
		}
		if err := configure.SetOutboundSetting(outbound, next); err != nil {
			if replaceMembers {
				_ = restoreMembers()
			}
			return err
		}
		return nil
	})
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
	case configure.LeastPing, configure.KeepCurrent, configure.RoundRobin, configure.Random:
	default:
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

func connectedGroupCandidates(name string) []groupCandidate {
	refs := configure.GetConnectedServersByOutbound(name)
	if refs == nil {
		return nil
	}
	candidates := make([]groupCandidate, 0, refs.Len())
	for _, rawRef := range refs.Get() {
		ref := *rawRef
		located, err := ref.LocateServerRaw()
		if err != nil || located.ServerObj == nil {
			continue
		}
		candidates = append(candidates, groupCandidate{
			ref:      ref,
			node:     located.ServerObj,
			identity: fmt.Sprintf("%s/%d/%d/%s", ref.TYPE, ref.Sub, ref.ID, located.ServerObj.ExportToURL()),
		})
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
		Selected      string
		Candidates    []string
	}{setting.AutoAdd, setting.ProbeURL, setting.ProbeInterval, setting.Type, setting.Selected, identities})
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

func (a *automation) processGroup(ctx context.Context, name string, setting configure.OutboundSetting, candidates []groupCandidate, probe func([]serverObj.ServerObj, string) ([]subscriptionProbeResult, error)) {
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
		healthy := make(map[string]bool, len(candidates))
		for i, result := range results {
			if result.err == nil {
				if candidates[i].node != nil {
					healthy[configure.NodeFingerprint(candidates[i].node.ExportToURL())] = true
				}
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
			currentCandidates := connectedGroupCandidates(name)
			if currentSetting.AutoAdd {
				currentCandidates = automaticGroupCandidates()
			}
			if automaticGroupSignature(currentSetting, currentCandidates) != signature {
				err = fmt.Errorf("catalog or group setting changed during membership check")
			} else {
				nextSetting := currentSetting
				stickyChanged := false
				if currentSetting.Type == configure.KeepCurrent && currentSetting.Selected == "" {
					nextCurrent := currentSetting.StickyCurrent
					if !healthy[nextCurrent] {
						nextCurrent = ""
						for i, result := range results {
							if result.err == nil && candidates[i].node != nil {
								nextCurrent = configure.NodeFingerprint(candidates[i].node.ExportToURL())
								break
							}
						}
					}
					stickyChanged = nextCurrent != currentSetting.StickyCurrent
					nextSetting.StickyCurrent = nextCurrent
				}
				membershipChanged := currentSetting.AutoAdd && !sameMembers(configure.GetConnectedServersByOutbound(name).Get(), members)
				switch {
				case stickyChanged || (membershipChanged && currentSetting.Type == configure.KeepCurrent):
					err = a.applyState(name, nextSetting, members, currentSetting.AutoAdd)
				case membershipChanged:
					err = a.applyGroup(name, members)
				}
				if err == nil && (stickyChanged || membershipChanged) {
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
	catalogCandidates := automaticGroupCandidates()
	groups := make(map[string]configure.OutboundSetting)
	groupCandidates := make(map[string][]groupCandidate)
	for _, name := range configure.GetOutbounds() {
		setting := configure.GetOutboundSetting(name)
		groups[name] = setting
		if !setting.AutoAdd && setting.Type == configure.KeepCurrent {
			groupCandidates[name] = connectedGroupCandidates(name)
		}
	}
	ConfigurationMu.Unlock()
	activeGroups := map[string]bool{}
	for name, setting := range groups {
		if !setting.AutoAdd && setting.Type != configure.KeepCurrent {
			continue
		}
		activeGroups[name] = true
		if err := ValidateOutboundSetting(setting); err != nil {
			log.Warn("[Groups] %s: invalid automatic membership setting: %v", name, err)
			continue
		}
		candidates := groupCandidates[name]
		if setting.AutoAdd {
			candidates = catalogCandidates
		}
		a.processGroup(ctx, name, setting, candidates, probe)
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
