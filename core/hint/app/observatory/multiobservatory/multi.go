// Package multiobservatory provides a multi-group observatory for the merged v2xray core.
// It creates one Observatory instance per group tag and aggregates their results,
// enabling v2ray-compatible MultiObservatory semantics on top of xray-core.
package multiobservatory

import (
	"context"
	"sync"
	"time"

	xray_obs "github.com/xtls/xray-core/app/observatory"
	"github.com/xtls/xray-core/app/observatory/burst"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/features/extension"
	"google.golang.org/protobuf/proto"
)

// MultiObservatory holds multiple Observatory instances keyed by group tag.
// It implements extension.Observatory for aggregated results and provides
// GetObservationByTag for per-group queries.
type MultiObservatory struct {
	mu       sync.RWMutex
	children map[string]extension.Observatory // tag -> Observatory
	ctx      context.Context
}

// Type implements features.Feature.
func (m *MultiObservatory) Type() interface{} {
	return extension.ObservatoryType()
}

// Start implements features.Feature by starting all child observers.
func (m *MultiObservatory) Start() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, child := range m.children {
		if err := child.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Close implements features.Feature by closing all child observers.
func (m *MultiObservatory) Close() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, child := range m.children {
		_ = child.Close()
	}
	return nil
}

// GetObservation implements extension.Observatory.
// It returns the aggregated ObservationResult of all child observers.
func (m *MultiObservatory) GetObservation(ctx context.Context) (proto.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var combined []*xray_obs.OutboundStatus
	for _, child := range m.children {
		result, err := child.GetObservation(ctx)
		if err != nil {
			continue
		}
		if obs, ok := result.(*xray_obs.ObservationResult); ok {
			combined = append(combined, obs.GetStatus()...)
		}
	}
	return &xray_obs.ObservationResult{Status: combined}, nil
}

// GetObservationByTag returns the ObservationResult for a specific group tag.
// If the tag is empty or not found, the aggregated result is returned.
func (m *MultiObservatory) GetObservationByTag(tag string, ctx context.Context) (proto.Message, error) {
	m.mu.RLock()
	child, ok := m.children[tag]
	m.mu.RUnlock()

	if !ok || tag == "" {
		return m.GetObservation(ctx)
	}
	return child.GetObservation(ctx)
}

// New creates a MultiObservatory from Config, registering each group as a child Observer.
func New(ctx context.Context, config *Config) (*MultiObservatory, error) {
	mo := &MultiObservatory{
		children: make(map[string]extension.Observatory),
		ctx:      ctx,
	}

	for _, obs := range config.GetObservers() {
		pingCfg := &burst.HealthPingConfig{
			Destination: obs.GetProbeUrl(),
			Interval:    obs.GetProbeInterval(),
		}
		// Keep legacy 10s probing behavior when interval isn't configured.
		if pingCfg.Interval == 0 {
			pingCfg.Interval = int64(10 * time.Second)
		}

		childCfg := &burst.Config{
			SubjectSelector: obs.GetSubjectSelector(),
			PingConfig:      pingCfg,
		}
		child, err := burst.New(ctx, childCfg)
		if err != nil {
			return nil, err
		}
		mo.children[obs.GetTag()] = child
	}
	return mo, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
