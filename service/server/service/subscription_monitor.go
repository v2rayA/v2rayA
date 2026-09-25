package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const monitorInterval = 10 * time.Second
const monitorFailureWindow = time.Minute

// There is one worker for the active subscription, never one worker per node.
// Requests are coalesced; a scheduled/manual refresh can request early recovery.
var subscriptionRecoveryRequests = make(chan string, 1)

var subscriptionScanMu sync.Mutex
var recoveryCancelMu sync.Mutex
var recoveryCancel context.CancelFunc

// Called before user mutations. Cancelling the job is non-blocking; its final
// identity check under ConfigurationMu also protects against stale results.
func CancelSubscriptionRecovery() {
	recoveryCancelMu.Lock()
	defer recoveryCancelMu.Unlock()
	if recoveryCancel != nil {
		recoveryCancel()
	}
}

type monitorTarget struct {
	key          string
	index        int
	port         int
	probeURL     string
	template     *v2ray.Template
	subscription *configure.SubscriptionRaw
	client       *http.Client
}

// Call with ConfigurationMu held. A manual stop or standalone selection is an
// explicit user choice, and must never be undone by monitoring.
func activeMonitorTarget() *monitorTarget {
	if !configure.GetRunning() {
		return nil
	}
	ws := configure.GetConnectedServersByOutbound("proxy").Get()
	if len(ws) != 1 || ws[0].TYPE != configure.SubscriptionServerType {
		return nil
	}
	w := ws[0]
	sub := configure.GetSubscription(w.Sub)
	if sub == nil || !sub.Monitor || w.ID < 1 || w.ID > len(sub.Servers) {
		return nil
	}
	t := v2ray.ProcessManager.GetRunningTemplate()
	if t != nil && t.Variant != where.Xray {
		return nil
	}
	url := configure.GetOutboundSetting("proxy").ProbeURL
	if url == "" {
		url = HttpTestURL
	}
	result := &monitorTarget{index: w.Sub, probeURL: url, template: t, subscription: sub,
		key: fmt.Sprintf("%d\n%s\n%s\n%s", w.Sub, sub.Address, sub.Servers[w.ID-1].ServerObj.ExportToURL(), url)}
	if t != nil {
		result.port = t.SubscriptionMonitorPort
	}
	return result
}

func requestSubscriptionRecovery() {
	if target := activeMonitorTarget(); target != nil {
		select {
		case subscriptionRecoveryRequests <- target.key:
		default:
		}
	}
}

func sameMonitorTarget(a, b *monitorTarget) bool {
	return a != nil && b != nil && a.key == b.key && a.template == b.template && reflect.DeepEqual(a.subscription, b.subscription)
}

type monitorState struct {
	key         string
	template    *v2ray.Template
	failedSince time.Time
	recovering  bool
	failures    int
	requested   string
}

func (s *monitorState) observe(now time.Time, healthy bool) bool {
	if healthy {
		s.failedSince = time.Time{}
		return false
	}
	if s.failedSince.IsZero() {
		s.failedSince = now
	}
	return now.Sub(s.failedSince) >= monitorFailureWindow
}

func (s *monitorState) retryDelay() time.Duration {
	s.failures++
	// One immediate retry handles a provider publishing an updated list. Then
	// leave increasing idle windows for the GUI and cap retries at twice/minute.
	switch s.failures {
	case 1:
		return 0
	case 2:
		return 5 * time.Second
	case 3:
		return 10 * time.Second
	case 4:
		return 20 * time.Second
	default:
		return 30 * time.Second
	}
}

func (s *monitorState) step(ctx context.Context) time.Duration {
	if !ConfigurationMu.TryLock() {
		return time.Second
	}
	target := activeMonitorTarget()
	if target == nil {
		*s = monitorState{}
		ConfigurationMu.Unlock()
		return monitorInterval
	}
	if s.key != target.key || (!s.recovering && s.template != target.template) {
		*s = monitorState{key: target.key, template: target.template, requested: s.requested}
	}
	if s.requested == target.key {
		s.recovering = true
	}
	s.requested = ""
	ConfigurationMu.Unlock()
	if !s.recovering {
		var err error
		if target.port == 0 {
			err = fmt.Errorf("active core has no monitor listener")
		} else {
			_, err = probeHTTP(ctx, net.JoinHostPort("127.0.0.1", strconv.Itoa(target.port)), target.probeURL, subscriptionProbeTimeout)
		}
		if !s.observe(time.Now(), err == nil) {
			return monitorInterval
		}
	}
	if ctx.Err() != nil || !ConfigurationMu.TryLock() {
		return time.Second
	}
	current := activeMonitorTarget()
	if !sameMonitorTarget(current, target) {
		*s = monitorState{}
		ConfigurationMu.Unlock()
		return monitorInterval
	}
	target.client = httpClient.GetHttpClientAutomatically()
	ConfigurationMu.Unlock()
	if !s.recovering {
		log.Info("[Monitor] Subscription %d: connection failed for one minute; starting recovery", target.index+1)
	}
	s.recovering = true
	if err := recoverMonitoredSubscription(ctx, target); err != nil {
		// A refreshed first entry can change identity while still being dead.
		// Keep the recovery loop active across that change, without bypassing backoff.
		if errors.Is(err, ErrNoReachableSubscriptionServer) && ConfigurationMu.TryLock() {
			if current := activeMonitorTarget(); current != nil && current.index == target.index && current.subscription.PreferFirst {
				s.key, s.template = current.key, current.template
			}
			ConfigurationMu.Unlock()
		}
		delay := s.retryDelay()
		if ctx.Err() == nil {
			log.Warn("[Monitor] Subscription %d: recovery attempt %d failed; retry in %s: %v", target.index+1, s.failures, delay, err)
		}
		return delay
	}
	log.Info("[Monitor] Subscription %d: connection recovered", target.index+1)
	*s = monitorState{}
	return monitorInterval
}

func recoverMonitoredSubscription(parent context.Context, target *monitorTarget) error {
	if !subscriptionScanMu.TryLock() {
		return fmt.Errorf("another subscription scan is running")
	}
	defer subscriptionScanMu.Unlock()
	ctx, cancel := context.WithCancel(parent)
	if !ConfigurationMu.TryLock() {
		cancel()
		return fmt.Errorf("configuration is busy")
	}
	if !sameMonitorTarget(activeMonitorTarget(), target) {
		ConfigurationMu.Unlock()
		cancel()
		return fmt.Errorf("configuration changed before recovery")
	}
	recoveryCancelMu.Lock()
	recoveryCancel = cancel
	recoveryCancelMu.Unlock()
	ConfigurationMu.Unlock()
	defer func() {
		cancel()
		recoveryCancelMu.Lock()
		recoveryCancel = nil
		recoveryCancelMu.Unlock()
	}()
	index, old := target.index, target.subscription
	servers, info, downloadErr := resolveSubscriptionWithContext(ctx, old.Address, target.client)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if downloadErr != nil || len(servers) == 0 {
		// A subscription host can depend on the failed tunnel. Cached candidates
		// may restore connectivity without accepting an empty/broken replacement.
		log.Warn("[Monitor] Subscription %d: refresh unavailable; checking saved candidates", index+1)
		servers = make([]serverObj.ServerObj, len(old.Servers))
		for i := range old.Servers {
			servers[i] = old.Servers[i].ServerObj
		}
		info = old.Info
	}
	probeCandidates := policyProbe(old.PreferFirst, func(nodes []serverObj.ServerObj, url string) []subscriptionProbeResult {
		return probeSubscriptionWithContext(ctx, nodes, url)
	})
	results := probeCandidates(servers, target.probeURL)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !ConfigurationMu.TryLock() {
		return fmt.Errorf("configuration changed during recovery")
	}
	defer ConfigurationMu.Unlock()
	current := activeMonitorTarget()
	if ctx.Err() != nil || !sameMonitorTarget(current, target) {
		return fmt.Errorf("recovery cancelled by a configuration change")
	}
	probe := func([]serverObj.ServerObj, string) []subscriptionProbeResult { return results }
	return applySubscriptionSelection(index, old, servers, info, probe, true)
}

// Stop cancels in-flight HTTP requests/probe cores and waits for this one worker.
func StartSubscriptionMonitor() (stop func()) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		timer := time.NewTimer(monitorInterval)
		defer timer.Stop()
		state := monitorState{}
		for {
			select {
			case <-ctx.Done():
				return
			case key := <-subscriptionRecoveryRequests:
				state.requested = key
			case <-timer.C:
			}
			next := state.step(ctx)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(next)
		}
	}()
	return func() { cancel(); <-done }
}
