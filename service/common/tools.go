package common

import (
	"github.com/mzz2017/go-engine/src/common"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func BoolToInt(a bool) int {
	if a {
		return 1
	}
	return 0
}

func BoolToString(a bool) string {
	if a {
		return "true"
	}
	return "false"
}

func VersionMustGreaterEqual(v1, v2 string) (is bool) {
	is, _ = VersionGreaterEqual(v1, v2)
	return
}

func Deduplicate(list []string) []string {
	res := make([]string, 0, len(list))
	m := make(map[string]struct{})
	for _, v := range list {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		res = append(res, v)
	}
	return res
}

/* return if v1 is after v2 */
func VersionGreaterEqual(v1, v2 string) (is bool, err error) {
	if v1 == "UnknownClient" {
		return true, nil
	}
	if v2 == "UnknownClient" {
		return false, nil
	}
	var HighPriority = []string{"debug", "unstable"}
	if common.ArrayContainString(HighPriority, v1) {
		return true, nil
	}
	if common.ArrayContainString(HighPriority, v2) {
		return false, nil
	}
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	v1 = strings.ReplaceAll(v1, "-", ".")
	v2 = strings.ReplaceAll(v2, "-", ".")
	a1 := strings.Split(v1, ".")
	a2 := strings.Split(v2, ".")
	l := Min(len(a1), len(a2))
	var vv1, vv2 int
	for i := 0; i < l; i++ {
		vv1, err = strconv.Atoi(a1[i])
		if err != nil {
			return
		}
		vv2, err = strconv.Atoi(a2[i])
		if err != nil {
			return
		}
		if vv1 < vv2 {
			return false, nil
		} else if vv1 > vv2 {
			return true, nil
		}
	}
	return len(a1) >= len(a2), nil
}

func IsInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return !os.IsNotExist(err)
}

// UrlEncoded encodes a string like Javascript's encodeURIComponent()
func UrlEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		return str
	}
	return u.String()
}

func TrimLineContains(parent, sub string) string {
	lines := strings.Split(parent, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, sub) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
