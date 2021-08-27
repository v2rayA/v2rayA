package configure

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type PortWhiteList struct {
	TCP []string `json:"tcp"`
	UDP []string `json:"udp"`
}

var r1 = regexp.MustCompile(`^\d+$`)
var r2 = regexp.MustCompile(`^\d+:\d+$`)

func (pwl *PortWhiteList) Valid() bool {
	for _, t := range pwl.TCP {
		if !r1.MatchString(t) && !r2.MatchString(t) {
			return false
		}
	}
	for _, t := range pwl.UDP {
		if !r1.MatchString(t) && !r2.MatchString(t) {
			return false
		}
	}
	return true
}

func (pwl *PortWhiteList) Compressed() (wl *PortWhiteList) {
	wl = new(PortWhiteList)
	solve := func(ps []string) (r []string) {
		var m [65537]bool
		r = make([]string, 0)
		for _, p := range ps {
			if r1.MatchString(p) {
				p, _ := strconv.Atoi(p)
				if p < 0 || p > 65535 {
					continue
				}
				m[p] = true
			} else if r2.MatchString(p) {
				arr := strings.Split(p, ":")
				l, _ := strconv.Atoi(arr[0])
				r, _ := strconv.Atoi(arr[1])
				for i := l; i <= r; i++ {
					m[i] = true
				}
			}
		}
		l := -1
		for i := 1; i <= 65536; i++ { //m[65536]一定是个false
			if m[i] && l == -1 {
				l = i
			} else if !m[i] && l > -1 {
				if i-l == 1 {
					r = append(r, strconv.Itoa(l))
				} else {
					r = append(r, fmt.Sprintf("%v:%v", l, i-1))
				}
				l = -1
			}
		}
		return
	}
	t := solve(pwl.TCP)
	if len(t) > 0 {
		wl.TCP = t
	}
	u := solve(pwl.UDP)
	if len(u) > 0 {
		wl.UDP = u
	}
	return
}

func (pwl *PortWhiteList) Has(port string, protocol string) (has bool) {
	iPort, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	has = false
	var list []string
	switch strings.ToLower(protocol) {
	case "tcp":
		list = pwl.TCP
	case "udp":
		list = pwl.UDP
	default:
		return false
	}
	for _, t := range list {
		if t == port {
			has = true
			break
		} else if strings.Contains(t, ":") {
			arr := strings.Split(t, ":")
			l, _ := strconv.Atoi(arr[0])
			r, _ := strconv.Atoi(arr[1])
			if iPort >= l && iPort <= r {
				has = true
				break
			}
		}
	}
	return has
}
