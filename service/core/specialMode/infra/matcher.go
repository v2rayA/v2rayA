package infra

import (
	"github.com/v2fly/v2ray-core/v4/common/strmatcher"
	"regexp"
	"strings"
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
	g.g.Add(dm, g.id)
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
	g.g.Add(dm, g.id)
}

type RegexMatcher struct {
	Pattern *regexp.Regexp
}

func (m *RegexMatcher) Match(s string) bool {
	return m.Pattern.MatchString(s)
}
func (m *RegexMatcher) String() string {
	return "regexp:" + m.Pattern.String()
}

type SubstrMatcher string

func (m SubstrMatcher) Match(s string) bool {
	return strings.Contains(s, string(m))
}
func (m SubstrMatcher) String() string {
	return "contains:" + string(m)
}
