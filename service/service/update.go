package service

import (
	"V2RayA/global"
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func CheckUpdate() (foundNew bool, remoteVersion string, err error) {
	resp, err := http.Get("https://apt.v2raya.mzz.pub/dists/v2raya/main/binary-amd64/Packages")
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
		return false, "", errors.New("检查更新失败，没有从Packages文件中找到Package: v2raya")
	}
	s = s[l:]
	prefix := "Version: "
	l = strings.Index(s, prefix)
	if l < 0 {
		return false, "", errors.New("检查更新失败，没有从Packages文件中找到Version: ")
	}
	s = s[l+len(prefix):]
	r := strings.Index(s, "\n")
	if r < 0 { //没有换行就到末尾
		r = len(s)
	}
	s = s[:r]
	// 远端版本获取完毕
	if strings.ToLower(global.Version) == "debug" {
		return false, s, nil //debug模式无需检查更新
	}
	local := strings.Split(global.Version, ".")
	remote := strings.Split(s, ".")
	mlen := len(local)
	if len(remote) < mlen {
		mlen = len(remote)
	}
	for i := 0; i < mlen; i++ {
		lc, err := strconv.Atoi(local[i])
		if err != nil {
			return false, "", err
		}
		rm, err := strconv.Atoi(remote[i])
		if err != nil {
			return false, "", err
		}
		if lc < rm { //按节比较，某一节如果本地小于远端，则需要更新
			return true, s, nil
		} else if lc > rm { //奇怪的事情，本地版本高于远端
			return false, "", errors.New("奇怪的事情发生了，本地版本高于远端")
		}
	}
	return len(remote) > len(local), s, nil //如果前面都一样，远端版本号长度更长则需要更新
}
