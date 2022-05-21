package service

import (
	"bytes"
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"net/http"
	"strings"
)

func CheckUpdate() (foundNew bool, remoteVersion string, err error) {
	resp, err := http.Get("https://apt.v2raya.org/dists/v2raya/main/binary-amd64/Packages")
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(resp.Body)
	if err != nil && n > 0 {
		return
	}
	defer resp.Body.Close()
	s := buf.String()
	l := strings.Index(s, "Package: v2raya")
	if l < 0 {
		return false, "", fmt.Errorf("failed to get latest version from Package file: 1")
	}
	s = s[l:]
	prefix := "Version: "
	l = strings.Index(s, prefix)
	if l < 0 {
		return false, "", fmt.Errorf("failed to get latest version from Package file: 2")
	}
	s = s[l+len(prefix):]
	r := strings.Index(s, "\n")
	if r < 0 { //没有换行就到末尾
		r = len(s)
	}
	s = s[:r]
	// 远端版本获取完毕
	ge, err := common.VersionGreaterEqual(conf.Version, s)
	return !ge, s, err
}
