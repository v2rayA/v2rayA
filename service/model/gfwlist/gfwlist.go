package gfwlist

import (
	"V2RayA/extra/gopeed"
	"V2RayA/model/v2ray"
	"V2RayA/model/v2ray/asset"
	"V2RayA/persistence/configure"
	"V2RayA/tools/files"
	"V2RayA/tools/httpClient"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type GFWList struct {
	UpdateTime *time.Time
	Tag        string
	sync.Mutex
}

var g GFWList

func GetRemoteGFWListUpdateTime(c *http.Client) (gfwlist GFWList, err error) {
	g.Lock()
	defer g.Unlock()
	if g.UpdateTime != nil {
		return g, nil
	}
	resp, err := httpClient.HttpGetUsingSpecificClient(c, "https://api.github.com/repos/mzz2017/dist-v2ray-rules-dat/tags")
	if err != nil {
		err = errors.New("fail in get latest version of GFWList: " + err.Error())
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	tag := gjson.GetBytes(b, "0.name").Str
	u := gjson.GetBytes(b, "0.commit.url").Str
	if tag == "" || u == "" {
		err = errors.New("fail in get latest version of GFWList: fail in getting latest tag")
		return
	}
	resp, err = httpClient.HttpGetUsingSpecificClient(c, u)
	if err != nil {
		err = errors.New("fail in get latest version of GFWList: " + err.Error())
		return
	}
	b, _ = ioutil.ReadAll(resp.Body)
	t := gjson.GetBytes(b, "commit.committer.date").Time()
	g.Tag = tag
	g.UpdateTime = &t
	if t.IsZero() {
		err = errors.New("fail in get latest version of GFWList: fail in getting commit date of latest tag")
		return
	}
	return g, nil
}
func IsUpdate() (update bool, remoteTime time.Time, err error) {
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	remoteTime = *gfwlist.UpdateTime
	if !asset.IsGFWListExists() {
		//本地文件不存在，那远端必定比本地新
		return false, remoteTime, nil
	}
	//本地文件存在，检查本地版本是否比远端还新
	t, err := asset.GetGFWListModTime()
	if err != nil {
		return
	}
	if t.After(remoteTime) {
		//那确实新
		update = true
		return
	}
	return
}

func UpdateLocalGFWList() (localGFWListVersionAfterUpdate string, err error) {
	i := 0
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	//u := fmt.Sprintf(`https://cdn.jsdelivr.net/gh/mzz2017/dist-v2ray-rules-dat@%v/geoip.dat`, gfwlist.Tag)
	//err = gopeed.Down(&gopeed.Request{
	//	Method: "GET",
	//	URL:    u,
	//}, asset.GetV2rayLocationAsset()+"/LoyalsoldierIP.dat")
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	u := fmt.Sprintf(`https://cdn.jsdelivr.net/gh/mzz2017/dist-v2ray-rules-dat@%v/geosite.dat`, gfwlist.Tag)
	err = gopeed.Down(&gopeed.Request{
		Method: "GET",
		URL:    u,
	}, asset.GetV2rayLocationAsset()+"/LoyalsoldierSite.dat")
	if err != nil {
		log.Println(err)
		return
	}
	t, err := files.GetFileModTime(asset.GetV2rayLocationAsset() + "/LoyalsoldierSite.dat")
	if err == nil {
		localGFWListVersionAfterUpdate = t.Format("2006-01-02")
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
		return "", errors.New(
			"latest version is " + tRemote.Format("2006-01-02") + ". current GFWList is up to date",
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
