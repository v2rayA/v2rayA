package infra

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/common/strmatcher"
	"github.com/v2rayA/v2ray-lib/router/routercommon"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"os"
	"regexp"
	"sync"
)

var whitelistCn struct {
	domainMatcher *strmatcher.MatcherGroup
	sync.Mutex
}

func GetWhitelistCn(externDomains []*routercommon.Domain) (wlDomains *strmatcher.MatcherGroup, err error) {
	whitelistCn.Lock()
	defer whitelistCn.Unlock()
	if whitelistCn.domainMatcher != nil {
		return whitelistCn.domainMatcher, nil
	}
	datpath, err := asset.GetV2rayLocationAsset("geosite.dat")
	if err != nil {
		return nil, err
	}
	var siteList routercommon.GeoSiteList
	b, err := os.ReadFile(datpath)
	if err != nil {
		return nil, fmt.Errorf("GetWhitelistCn: %w", err)
	}
	err = proto.Unmarshal(b, &siteList)
	if err != nil {
		return nil, fmt.Errorf("GetWhitelistCn: %w", err)
	}
	wlDomains = new(strmatcher.MatcherGroup)
	domainMatcher := new(DomainMatcherGroup)
	fullMatcher := new(FullMatcherGroup)
	for _, e := range siteList.Entry {
		if e.CountryCode == "CN" {
			dms := e.Domain
			dms = append(dms, externDomains...)
			for _, dm := range dms {
				switch dm.Type {
				case routercommon.Domain_RootDomain:
					domainMatcher.Add(dm.Value)
				case routercommon.Domain_Full:
					fullMatcher.Add(dm.Value)
				case routercommon.Domain_Plain:
					wlDomains.Add(SubstrMatcher(dm.Value))
				case routercommon.Domain_Regex:
					r, err := regexp.Compile(dm.Value)
					if err != nil {
						break
					}
					wlDomains.Add(&RegexMatcher{Pattern: r})
				}
			}
			break
		}
	}
	domainMatcher.Add("lan")
	wlDomains.Add(domainMatcher)
	wlDomains.Add(fullMatcher)
	whitelistCn.domainMatcher = wlDomains
	return
}
