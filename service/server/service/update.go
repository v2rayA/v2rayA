package service

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
)

func CheckUpdate() (foundNew bool, remoteVersion string, err error) {
	resp, err := http.Get("https://raw.githubusercontent.com/v2rayA/v2raya-apt/master/dists/v2raya/main/binary-amd64/Packages")
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
		return false, "", fmt.Errorf("the APT package index has no v2rayA entry; cannot tell whether a newer version exists")
	}
	s = s[l:]
	prefix := "Version: "
	l = strings.Index(s, prefix)
	if l < 0 {
		return false, "", fmt.Errorf("the v2rayA entry in the APT package index has no Version field")
	}
	s = s[l+len(prefix):]
	r := strings.Index(s, "\n")
	if r < 0 { // Go to end if no newline
		r = len(s)
	}
	s = s[:r]
	// Remote version obtained
	ge, err := common.VersionGreaterEqual(conf.Version, s)
	return !ge, s, err
}
