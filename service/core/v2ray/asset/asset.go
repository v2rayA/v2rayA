package asset

import (
	"github.com/golang/protobuf/proto"
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/common/files"
	"github.com/mzz2017/v2rayA/core/dnsPoison"
	"github.com/mzz2017/v2rayA/core/v2ray/where"
	"github.com/mzz2017/v2rayA/global"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
	v2router "v2ray.com/core/app/router"
	"v2ray.com/core/common/strmatcher"
)

var v2rayLocationAsset *string

func GetV2rayLocationAsset() (s string) {
	if v2rayLocationAsset != nil {
		return *v2rayLocationAsset
	}
	switch global.ServiceControlMode {
	case global.SystemctlMode, global.ServiceMode:
		p, _ := where.GetV2rayServiceFilePath()
		out, err := exec.Command("sh", "-c", "cat "+p+"|grep Environment=V2RAY_LOCATION_ASSET").CombinedOutput()
		if err != nil {
			break
		}
		s = strings.TrimSpace(string(out))
		s = s[len("Environment=V2RAY_LOCATION_ASSET="):]
	}
	var err error
	if s == "" {
		//fine, guess one
		var ver string
		var is bool
		if ver, err = where.GetV2rayServiceVersion(); err == nil {
			if is, err = common.VersionGreaterEqual(ver, "4.27.1"); is {
				s = "/usr/share/v2ray"
			}
		}
		if s == "" {
			//maybe v2ray working directory
			s, err = where.GetV2rayWorkingDir()
			if err != nil {
				s = "/etc/v2ray"
			}
		}
	} else {
		//save the result if not by guess
		v2rayLocationAsset = &s
	}
	return
}

func GetV2ctlDir() (string, error) {
	d, err := where.GetV2rayWorkingDir()
	if err == nil {
		_, err := os.Stat(d + "/v2ctl")
		if err != nil {
			return "", err
		}
		return d, nil
	}
	out, err := exec.Command("sh", "-c", "which v2ctl").Output()
	if err != nil {
		err = newError(string(out)).Base(err)
		return "", err
	}
	return path.Dir(strings.TrimSpace(string(out))), nil
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
	b, err = ioutil.ReadFile(GetConfigPath())
	if err != nil {
		log.Println(err)
		return
	}
	b = jsonc.ToJSON(b)
	return
}

func GetConfigPath() (p string) {
	p = "/etc/v2ray/config.json"
	switch global.ServiceControlMode {
	case global.SystemctlMode, global.ServiceMode:
		//从systemd的启动参数里找
		pa, _ := where.GetV2rayServiceFilePath()
		out, e := exec.Command("sh", "-c", "cat "+pa+"|grep ExecStart=").CombinedOutput()
		if e != nil {
			return
		}
		pa = strings.TrimSpace(string(out))[len("ExecStart="):]
		indexConfigBegin := strings.Index(pa, "-config")
		if indexConfigBegin == -1 {
			return
		}
		indexConfigBegin += len("-config") + 1
		indexConfigEnd := strings.Index(pa[indexConfigBegin:], " ")
		if indexConfigEnd == -1 {
			indexConfigEnd = len(pa)
		} else {
			indexConfigEnd += indexConfigBegin
		}
		p = pa[indexConfigBegin:indexConfigEnd]
	case global.UniversalMode:
		p = GetV2rayLocationAsset() + "/config.json"
	default:
	}
	return
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
	b, err := ioutil.ReadFile(path.Join(dir, "geosite.dat"))
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

