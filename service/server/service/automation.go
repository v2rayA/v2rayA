package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strings"
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

// RefreshAutomaticGroups only synchronizes membership. Selection is scheduled separately.
func RefreshAutomaticGroups(ctx context.Context, name string) error {
	return newAutomation().forceGroups(ctx, name)
}

type subscriptionSchedule struct {
	signature                         string
	nextUpdate, nextHealth, nextRetry time.Time
	warnedUnprobeable                 bool
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
	ping          func(context.Context, []serverObj.ServerObj) []subscriptionProbeResult
	fetch         subscriptionFetcher
	applyGroup    func(string, []configure.NodeRef) error
	applyMembers  func(string, []configure.NodeRef) error
	applyState    func(string, configure.OutboundSetting, []configure.NodeRef, bool) error
	groupErrors   map[string]error
	choose        func(int) int
}

func newAutomation() *automation {
	return &automation{
		subscriptions: map[int64]*subscriptionSchedule{},
		groups:        map[string]*groupSchedule{},
		now:           time.Now,
		probe:         probeSubscriptionWithContext,
		ping:          pingGroupCandidates,
		fetch:         fetchSubscriptionForAutomation,
		applyGroup:    replaceManagedOutboundConnections,
		applyMembers:  writeGroupMembers,
		applyState:    applyManagedGroupState,
		groupErrors:   map[string]error{},
		choose:        rand.Intn,
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
	case configure.Fixed:
		if setting.Selected == "" {
			return fmt.Errorf("fixed group requires a selected server")
		}
		if setting.AutoAdd {
			return fmt.Errorf("fixed group cannot manage membership automatically")
		}
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
	for _, node := range nodes {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		results := a.probe(ctx, []serverObj.ServerObj{node}, probeURL)
		if len(results) != 1 {
			return false, fmt.Errorf("incomplete subscription check")
		}
		if results[0].err == nil {
			return false, nil
		}
	}
	return true, ctx.Err()
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
		state.nextHealth = time.Time{}
	}

	if !regularDue && !retryDue {
		return nil
	}
	if sub.UpdateMode == configure.SubscriptionUpdateIntervalFailsafe {
		// Reserve a retry before cancellable work; a dashboard mutation must
		// not consume the only deadline that can resume recovery.
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
	unavailable, err := a.subscriptionUnavailable(ctx, sub, probeURL, state)
	if err != nil {
		return err
	}
	state.nextHealth, state.nextRetry = time.Time{}, time.Time{}
	if !unavailable {
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

func eligibleProbeResults(results []subscriptionProbeResult) []bool {
	eligible := make([]bool, len(results))
	for i, result := range results {
		eligible[i] = result.err == nil && result.speedMeasured && result.throughput >= subscriptionMinSpeed
	}
	return eligible
}

// Probe only as far as the strategy needs. TCP ordering does not start a core.
func (a *automation) selectGroup(ctx context.Context, setting configure.OutboundSetting, candidates []groupCandidate, probe func([]serverObj.ServerObj, string) ([]subscriptionProbeResult, error)) ([]subscriptionProbeResult, error) {
	results := make([]subscriptionProbeResult, len(candidates))
	nodes := make([]serverObj.ServerObj, len(candidates))
	for i := range candidates {
		nodes[i] = candidates[i].node
		results[i].err = fmt.Errorf("not selected")
	}
	checked := make(map[int]bool)
	check := func(i int) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		checked[i] = true
		one, err := probe([]serverObj.ServerObj{nodes[i]}, setting.ProbeURL)
		if err != nil {
			return false, err
		}
		if len(one) != 1 {
			return false, fmt.Errorf("incomplete candidate check")
		}
		results[i] = one[0]
		return eligibleProbeResults(one)[0], nil
	}
	if setting.Type == configure.KeepCurrent {
		for i, node := range nodes {
			if node != nil && configure.NodeFingerprint(node.ExportToURL()) == setting.StickyCurrent {
				ok, err := check(i)
				if err != nil || ok {
					return results, err
				}
				break
			}
		}
	}
	pings := a.ping(ctx, nodes)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if len(pings) != len(nodes) {
		return nil, fmt.Errorf("incomplete TCP ping results")
	}
	order := make([]int, 0, len(nodes))
	for i := range nodes {
		if nodes[i] != nil && !checked[i] && (pings[i].err == nil || udpOnlyCandidate(nodes[i])) {
			order = append(order, i)
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return pings[order[i]].latency < pings[order[j]].latency })
	if setting.Type == configure.Random {
		for i := len(order) - 1; i > 0; i-- {
			j := a.choose(i + 1)
			order[i], order[j] = order[j], order[i]
		}
	}
	for _, i := range order {
		ok, err := check(i)
		if err != nil {
			return results, err
		}
		if ok && setting.Type != configure.RoundRobin {
			break
		}
	}
	return results, nil
}

func selectedGroupCandidate(candidates []groupCandidate, results []subscriptionProbeResult) string {
	for i, ok := range eligibleProbeResults(results) {
		if ok && candidates[i].node != nil {
			return configure.NodeFingerprint(candidates[i].node.ExportToURL())
		}
	}
	return ""
}

func (a *automation) syncGroupMembers(name string, candidates []groupCandidate) error {
	members := make([]configure.NodeRef, len(candidates))
	for i := range candidates {
		members[i] = candidates[i].ref
		members[i].Outbound = name
	}
	if sameMembers(configure.GetConnectedServersByOutbound(name).Get(), members) {
		return nil
	}
	if err := a.applyMembers(name, members); err != nil {
		return err
	}
	v2ray.ApiFeed.ProductMessage("catalog_changed", nil)
	return nil
}

func (a *automation) processGroup(ctx context.Context, name string, setting configure.OutboundSetting, candidates []groupCandidate, probe func([]serverObj.ServerObj, string) ([]subscriptionProbeResult, error)) {
	if setting.AutoAdd {
		ConfigurationMu.Lock()
		err := a.syncGroupMembers(name, candidates)
		ConfigurationMu.Unlock()
		if err != nil {
			a.groupErrors[name] = err
			return
		}
	}
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

	results, err := a.selectGroup(ctx, setting, candidates, probe)
	if err == nil {
		usable := eligibleProbeResults(results)
		members := make([]configure.NodeRef, 0, len(candidates))
		eligibleFingerprints := make([]string, 0, len(candidates))
		for i, result := range results {
			if result.err == nil && usable[i] {
				ref := candidates[i].ref
				ref.Outbound = name
				members = append(members, ref)
				if candidates[i].node != nil {
					eligibleFingerprints = append(eligibleFingerprints, configure.NodeFingerprint(candidates[i].node.ExportToURL()))
				}
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
				if configure.UsesWorkerSelection(currentSetting.Type) && currentSetting.Selected == "" {
					nextCurrent := selectedGroupCandidate(candidates, results)
					stickyChanged = nextCurrent != currentSetting.StickyCurrent
					nextSetting.StickyCurrent = nextCurrent
				}
				// Cache measured members for every strategy. Switching to round
				// robin can then use the already verified set in its first reload,
				// instead of restarting the core again after a second probe.
				eligibleMembers := strings.Join(eligibleFingerprints, " ")
				eligibleChanged := currentSetting.EligibleMembers != eligibleMembers
				nextSetting.EligibleMembers = eligibleMembers
				switch {
				case stickyChanged || (eligibleChanged && currentSetting.Type == configure.RoundRobin):
					err = a.applyState(name, nextSetting, nil, false)
				case eligibleChanged:
					err = configure.SetOutboundSetting(name, nextSetting)
				}
				if err == nil && (stickyChanged || eligibleChanged) {
					log.Info("[Groups] %s: %d/%d candidates available", name, len(members), len(candidates))
					v2ray.ApiFeed.ProductMessage("catalog_changed", nil)
				}
			}
		}
		ConfigurationMu.Unlock()
	}
	if err != nil {
		a.groupErrors[name] = err
		state.next = a.now().Add(backoff)
		if ctx.Err() == nil {
			log.Warn("[Groups] %s: automatic membership failed: %v", name, err)
		}
		return
	}
	delete(a.groupErrors, name)
	state.next = a.now().Add(interval)
}

func (a *automation) forceGroups(ctx context.Context, name string) error {
	ConfigurationMu.Lock()
	defer ConfigurationMu.Unlock()
	found := name == ""
	candidates := automaticGroupCandidates()
	for _, outbound := range configure.GetOutbounds() {
		if !configure.GetOutboundSetting(outbound).AutoAdd || (name != "" && name != outbound) {
			continue
		}
		found = true
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := a.syncGroupMembers(outbound, candidates); err != nil {
			return err
		}
	}
	if !found {
		return fmt.Errorf("group %q does not manage membership automatically", name)
	}
	return nil
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
		if !setting.AutoAdd && configure.UsesWorkerProbe(setting.Type) {
			groupCandidates[name] = connectedGroupCandidates(name)
		}
	}
	ConfigurationMu.Unlock()
	activeGroups := map[string]bool{}
	for name, setting := range groups {
		if !setting.AutoAdd && !configure.UsesWorkerProbe(setting.Type) {
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
