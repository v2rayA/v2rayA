package service

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
)

func CheckUpdate() (foundNew bool, remoteVersion string, err error) {
	arch := runtime.GOARCH
	switch arch {
	case "386":
		arch = "i386"
	case "arm":
		arch = "armhf"
	case "mipsle":
		arch = "mips32le"
	}
	resp, err := http.Get("https://raw.githubusercontent.com/v2rayA/v2raya-apt/master/dists/v2raya/main/binary-" + arch + "/Packages")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}
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
