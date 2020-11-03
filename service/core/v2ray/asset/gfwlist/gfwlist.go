package gfwlist

import (
	"bytes"
	sha2562 "crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/extra/copyfile"
	"github.com/v2rayA/v2rayA/extra/gopeed"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type GFWList struct {
	UpdateTime time.Time
	Tag        string
}

var g GFWList
var gMutex sync.Mutex

func GetRemoteGFWListUpdateTime(c *http.Client) (gfwlist GFWList, err error) {
	gMutex.Lock()
	defer gMutex.Unlock()
	if !g.UpdateTime.IsZero() {
		return g, nil
	}
	resp, err := httpClient.HttpGetUsingSpecificClient(c, "https://api.github.com/repos/v2rayA/dist-v2ray-rules-dat/tags")
	if err != nil {
		err = newError("failed to get latest version of GFWList").Base(err)
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	tag := gjson.GetBytes(b, "0.name").Str
	u := gjson.GetBytes(b, "0.commit.url").Str
	if tag == "" || u == "" {
		err = newError("failed to get latest version of GFWList: fail in getting latest tag")
		return
	}
	resp, err = httpClient.HttpGetUsingSpecificClient(c, u)
	if err != nil {
		err = newError("failed to get latest version of GFWList").Base(err)
		return
	}
	b, _ = ioutil.ReadAll(resp.Body)
	t := gjson.GetBytes(b, "commit.committer.date").Time()
	if t.IsZero() {
		err = newError("failed to get latest version of GFWList: fail in getting commit date of latest tag")
		return
	}
	g.Tag = tag
	g.UpdateTime = t
	return g, nil
}
func IsUpdate() (update bool, remoteTime time.Time, err error) {
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	remoteTime = gfwlist.UpdateTime
	if !asset.IsGFWListExists() {
		//本地文件不存在，那远端必定比本地新
		return false, remoteTime, nil
	}
	//本地文件存在，检查本地版本是否比远端还新
	t, err := asset.GetGFWListModTime()
	if err != nil {
		return
	}
	if !t.Before(remoteTime) {
		//那确实新
		update = true
		return
	}
	return
}

func LoyalsoldierSiteDatExists() bool {
	if info, err := os.Stat(filepath.Join(asset.GetV2rayLocationAsset(), "LoyalsoldierSite.dat")); err == nil && !info.IsDir() {
		return true
	}
	return false
}

func checkSha256(p string, sha256 string) bool {
	if b, err := ioutil.ReadFile(p); err == nil {
		hash := sha2562.Sum256(b)
		return hex.EncodeToString(hash[:]) == sha256
	} else {
		return false
	}
}

var (
	FailCheckSha = newError("failed to check sum256sum of GFWList file")
	DamagedFile  = newError("damaged GFWList file, update it again please")
)

func UpdateLocalGFWList() (localGFWListVersionAfterUpdate string, err error) {
	i := 0
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	//u := fmt.Sprintf(`https://cdn.jsdelivr.net/gh/v2rayA/dist-v2ray-rules-dat@%v/geoip.dat`, gfwlist.Tag)
	//err = gopeed.Down(&gopeed.Request{
	//	Method: "GET",
	//	URL:    u,
	//}, asset.GetV2rayLocationAsset()+"/LoyalsoldierIP.dat")
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	pathSiteDat := filepath.Join(asset.GetV2rayLocationAsset(), "LoyalsoldierSite.dat")
	backup := filepath.Join(asset.GetV2rayLocationAsset(), "LoyalsoldierSite.dat.bak")
	var sucBackup bool
	if _, err = os.Stat(pathSiteDat); err == nil {
		//backup
		err = copyfile.CopyFile(pathSiteDat, backup)
		if err != nil {
			err = newError("fail to backup gfwlist file").Base(err)
			return
		}
		sucBackup = true
	}
	u := fmt.Sprintf(`https://cdn.jsdelivr.net/gh/v2rayA/dist-v2ray-rules-dat@%v/geosite.dat`, gfwlist.Tag)
	err = gopeed.Down(&gopeed.Request{
		Method: "GET",
		URL:    u,
	}, pathSiteDat)
	if err != nil {
		log.Println(err)
		return
	}
	u2 := fmt.Sprintf(`https://cdn.jsdelivr.net/gh/v2rayA/dist-v2ray-rules-dat@%v/geosite.dat.sha256sum`, gfwlist.Tag)
	err = gopeed.Down(&gopeed.Request{
		Method: "GET",
		URL:    u2,
	}, pathSiteDat+".sha256sum")
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err != nil {
			if sucBackup {
				_ = copyfile.CopyFile(backup, pathSiteDat)
			} else {
				_ = os.Remove(pathSiteDat)
			}
		}
	}()
	var b []byte
	if b, err = ioutil.ReadFile(pathSiteDat + ".sha256sum"); err == nil {
		f := bytes.Fields(b)
		if len(f) < 2 {
			err = FailCheckSha
			return
		}
		if !checkSha256(pathSiteDat, string(f[0])) {
			err = newError(DamagedFile)
			return
		}
	} else {
		err = FailCheckSha
		return
	}
	_ = os.Chtimes(pathSiteDat, gfwlist.UpdateTime, gfwlist.UpdateTime)
	t, err := files.GetFileModTime(pathSiteDat)
	if err == nil {
		localGFWListVersionAfterUpdate = t.Local().Format("2006-01-02")
	}
	log.Printf("download[%v]: %v -> SUCCESS\n", i+1, u)
	return
}

func CheckAndUpdateGFWList() (localGFWListVersionAfterUpdate string, err error) {
	update, tRemote, err := IsUpdate()
	if err != nil {
		return
	}
	if update {
		return "", newError(
			"latest version is " + tRemote.Local().Format("2006-01-02") + ". GFWList is up to date",
		)
	}

	/* 更新LoyalsoldierSite.dat */
	localGFWListVersionAfterUpdate, err = UpdateLocalGFWList()
	if err != nil {
		return
	}
	setting := configure.GetSettingNotNil()
	if v2ray.IsV2RayRunning() && //正在使用GFWList模式再重启
		(setting.Transparent == configure.TransparentGfwlist ||
			setting.Transparent == configure.TransparentClose && setting.PacMode == configure.GfwlistMode) {
		err = v2ray.UpdateV2RayConfig(nil)
	}
	return
}
