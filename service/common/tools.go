package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

var (
	NotSameTypeErr    = fmt.Errorf("cannot fill empty: the two value have different type")
	NeedPassInPointer = fmt.Errorf("the structure passed in should be a pointer")
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

func PrefixListSatisfyString(prefixList []string, str string) int {
	for i, v := range prefixList {
		if strings.HasPrefix(str, v) {
			return i
		}
	}
	return -1
}

/* return if v1 is after v2 */
func VersionGreaterEqual(v1, v2 string) (is bool, err error) {
	var HighPriority = []string{"debug", "unstable"}
	if PrefixListSatisfyString(HighPriority, v1) != -1 {
		return true, nil
	}
	if PrefixListSatisfyString(HighPriority, v2) != -1 {
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

// IsDocker return true only if the environment SHOULD be docker
func IsDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
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

// FillEmpty fill the empty field of the struct with default value given
func FillEmpty(toFill interface{}, defaultVal interface{}) error {
	ta := reflect.TypeOf(toFill)
	if ta.Kind() != reflect.Ptr {
		return NeedPassInPointer
	}
	tb := reflect.TypeOf(defaultVal)
	va := reflect.ValueOf(toFill)
	vb := reflect.ValueOf(defaultVal)
	for ta.Kind() == reflect.Ptr {
		ta = ta.Elem()
		va = va.Elem()
	}
	for tb.Kind() == reflect.Ptr {
		tb = tb.Elem()
		vb = vb.Elem()
	}
	if ta != tb {
		return NotSameTypeErr
	}
	for i := 0; i < va.NumField(); i++ {
		v := va.Field(i)
		if v.Type().Name() == "bool" {
			continue
		}
		if v.IsZero() {
			v.Set(vb.Field(i))
		}
	}
	return nil
}

// IsOpenWrt return true only if the operating system SHOULD be openwrt
func IsOpenWrt() bool {
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/etc/openwrt_release"); err == nil {
			return true
		}
	}
	return false
}

func SliceSub(slice []string, toSub []string) []string {
	var res = make([]string, 0, len(slice))
	var m = make(map[string]struct{})
	for _, s := range toSub {
		m[s] = struct{}{}
	}
	for _, s := range slice {
		if _, ok := m[s]; !ok {
			res = append(res, s)
		}
	}
	return res
}

func SliceHas(slice []string, set []string) []string {
	var res = make([]string, 0, len(slice))
	var m = make(map[string]struct{})
	for _, s := range set {
		m[s] = struct{}{}
	}
	for _, s := range slice {
		if _, ok := m[s]; ok {
			res = append(res, s)
		}
	}
	return res
}

func SliceToSet(slice []string) map[string]struct{} {
	var m = make(map[string]struct{})
	for _, s := range slice {
		m[s] = struct{}{}
	}
	return m
}

func BytesCopy(b []byte) []byte {
	var a = make([]byte, len(b))
	copy(a, b)
	return a
}

func ToBytes(val interface{}) (b []byte, err error) {
	buf := new(bytes.Buffer)
	if err = gob.NewEncoder(buf).Encode(val); err != nil {
		return nil, err
	}
	return BytesCopy(buf.Bytes()), nil
}

func HomeExpand(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func HasAnyPrefix(s string, prefix []string) bool {
	for _, p := range prefix {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
