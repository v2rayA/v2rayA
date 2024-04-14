package tun

import "github.com/v2fly/v2ray-core/v5/common/strmatcher"

type Matcher []strmatcher.Matcher

func (mg Matcher) Match(input string) bool {
	for _, m := range mg {
		if m.Match(input) {
			return true
		}
	}
	return false
}
