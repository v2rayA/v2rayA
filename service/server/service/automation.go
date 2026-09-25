package service

import (
	"context"
	"encoding/json"
	"fmt"
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
	warnedUnprobeable                 bool
}

type subscriptionFetcher func(context.Context, string, bool) ([]serverObj.ServerObj, string, error)

type automation struct {
	subscriptions map[int64]*subscriptionSchedule
	now           func() time.Time
	probe         func(context.Context, []serverObj.ServerObj, string) []subscriptionProbeResult
	fetch         subscriptionFetcher
}

func newAutomation() *automation {
	return &automation{
		subscriptions: map[int64]*subscriptionSchedule{},
		now:           time.Now,
		probe:         probeSubscriptionWithContext,
		fetch:         fetchSubscriptionForAutomation,
	}
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
	if current.AutoSelect {
		if err := SelectServersFromSubscription(index, true); err != nil {
			return nil, err
		}
	}
	if err := storeSubscriptionUpdate(index, current, nodes, info, false); err != nil {
		if current.AutoSelect {
			_ = SelectServersFromSubscription(index, false)
		}
		return nil, err
	}
	if current.AutoSelect {
		if err := SelectServersFromSubscription(index, false); err != nil {
			return nil, err
		}
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
	nodes, info, err := a.fetch(ctx, sub.Address, retryDue && sub.AllowDirectRecovery)
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
	return next
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
