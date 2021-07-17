package infra

import (
	"regexp"
	"strings"
	"sync/atomic"
	"v2ray.com/core/common/strmatcher"
)

type DomainMatcherGroup struct {
	id uint32
	g  strmatcher.DomainMatcherGroup
	strmatcher.Matcher
}

func (g *DomainMatcherGroup) Match(dm string) bool {
	return g.g.Match(dm) > 0
}

func (g *DomainMatcherGroup) Add(dm string) {
	atomic.AddUint32(&g.id, 1)
	g.g.Add(dm, g.id)
}

type FullMatcherGroup struct {
	id uint32
	g  strmatcher.FullMatcherGroup
	strmatcher.Matcher
}

func (g *FullMatcherGroup) Match(dm string) bool {
	return g.g.Match(dm) > 0
}

func (g *FullMatcherGroup) Add(dm string) {
	atomic.AddUint32(&g.id, 1)
	g.g.Add(dm, g.id)
}

type RegexMatcher struct {
	Pattern *regexp.Regexp
}

func (m *RegexMatcher) Match(s string) bool {
	return m.Pattern.MatchString(s)
}

type SubstrMatcher string

func (m SubstrMatcher) Match(s string) bool {
	return strings.Contains(s, string(m))
}