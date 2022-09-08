package infra

import (
	"github.com/v2fly/v2ray-core/v5/common/strmatcher"
	"sync/atomic"
)

type DomainMatcherGroup struct {
	id uint32
	g  strmatcher.DomainMatcherGroup
	strmatcher.Matcher
}

func (g *DomainMatcherGroup) Match(dm string) bool {
	return g.g.Match(dm) != nil
}

func (g *DomainMatcherGroup) Add(dm string) {
	atomic.AddUint32(&g.id, 1)
	g.g.AddDomainMatcher(strmatcher.DomainMatcher(dm), g.id)
}

type FullMatcherGroup struct {
	id uint32
	g  strmatcher.FullMatcherGroup
	strmatcher.Matcher
}

func (g *FullMatcherGroup) Match(dm string) bool {
	return g.g.Match(dm) != nil
}

func (g *FullMatcherGroup) Add(dm string) {
	atomic.AddUint32(&g.id, 1)
	g.g.AddFullMatcher(strmatcher.FullMatcher(dm), g.id)
}
