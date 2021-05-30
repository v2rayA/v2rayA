package asset

import (
	"github.com/golang/protobuf/proto"
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/core/dnsPoison"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/global"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"time"
	v2router "v2ray.com/core/app/router"
	"v2ray.com/core/common/strmatcher"
)

func GetV2rayLocationAsset() (s string) {
	var candidates = []string{`/usr/local/share/v2ray`, `/usr/share/v2ray`, `/usr/local/share/xray`, `/usr/share/xray`}
	var is bool
	if ver, err := where.GetV2rayServiceVersion(); err == nil {
		if is, err = common.VersionGreaterEqual(ver, "4.27.1"); is {
			for _, c := range candidates {
				if _, err := os.Stat(c); os.IsNotExist(err) {
					continue
				}
				if _, err := os.Stat(path.Join(c, "geoip.dat")); os.IsNotExist(err) {
					continue
				}
				s = c
				break
			}
		}
	}
	// old version of v2ray
	if s == "" {
		//maybe v2ray working directory
		v2rayPath, err := where.GetV2rayBinPath()
		if err != nil {
			s = "/etc/v2ray"
		}
		s = path.Dir(v2rayPath)
	}
	return
}

func IsGFWListExists() bool {
	_, err := os.Stat(GetV2rayLocationAsset() + "/LoyalsoldierSite.dat")
	if err != nil {
		return false
	}
	return true
}
func IsGeoipExists() bool {
	_, err := os.Stat(GetV2rayLocationAsset() + "/geoip.dat")
	if err != nil {
		return false
	}
	return true
}
func IsGeositeExists() bool {
	_, err := os.Stat(GetV2rayLocationAsset() + "/geosite.dat")
	if err != nil {
		return false
	}
	return true
}
func GetGFWListModTime() (time.Time, error) {
	return files.GetFileModTime(GetV2rayLocationAsset() + "/LoyalsoldierSite.dat")
}
func IsCustomExists() bool {
	_, err := os.Stat(GetV2rayLocationAsset() + "/custom.dat")
	if err != nil {
		return false
	}
	return true
}

func GetConfigBytes() (b []byte, err error) {
	b, err = os.ReadFile(GetV2rayConfigPath())
	if err != nil {
		log.Println(err)
		return
	}
	b = jsonc.ToJSON(b)
	return
}

func GetV2rayConfigPath() (p string) {
	return path.Join(global.GetEnvironmentConfig().Config, "config.json")
}

func GetV2rayConfigDirPath() (p string) {
	return global.GetEnvironmentConfig().V2rayConfigDirectory
}

func LoyalsoldierSiteDatExists() bool {
	if info, err := os.Stat(filepath.Join(GetV2rayLocationAsset(), "LoyalsoldierSite.dat")); err == nil && !info.IsDir() {
		return true
	}
	return false
}

var whitelistCn struct {
	domainMatcher *strmatcher.MatcherGroup
	sync.Mutex
}

func GetWhitelistCn(externIps []*v2router.CIDR, externDomains []*v2router.Domain) (wlDomains *strmatcher.MatcherGroup, err error) {
	whitelistCn.Lock()
	defer whitelistCn.Unlock()
	if whitelistCn.domainMatcher != nil {
		return whitelistCn.domainMatcher, nil
	}
	dir := GetV2rayLocationAsset()
	var siteList v2router.GeoSiteList
	b, err := os.ReadFile(path.Join(dir, "geosite.dat"))
	if err != nil {
		return nil, newError("GetWhitelistCn").Base(err)
	}
	err = proto.Unmarshal(b, &siteList)
	if err != nil {
		return nil, newError("GetWhitelistCn").Base(err)
	}
	wlDomains = new(strmatcher.MatcherGroup)
	domainMatcher := new(dnsPoison.DomainMatcherGroup)
	fullMatcher := new(dnsPoison.FullMatcherGroup)
	for _, e := range siteList.Entry {
		if e.CountryCode == "CN" {
			dms := e.Domain
			dms = append(dms, externDomains...)
			for _, dm := range dms {
				switch dm.Type {
				case v2router.Domain_Domain:
					domainMatcher.Add(dm.Value)
				case v2router.Domain_Full:
					fullMatcher.Add(dm.Value)
				case v2router.Domain_Plain:
					wlDomains.Add(dnsPoison.SubstrMatcher(dm.Value))
				case v2router.Domain_Regex:
					r, err := regexp.Compile(dm.Value)
					if err != nil {
						break
					}
					wlDomains.Add(&dnsPoison.RegexMatcher{Pattern: r})
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
