package gfwlist

import (
	"V2RayA/extra/copyfile"
	"V2RayA/extra/download"
	"V2RayA/model/v2ray"
	"V2RayA/model/v2ray/asset"
	"V2RayA/persistence/configure"
	"V2RayA/tools/files"
	"V2RayA/tools/httpClient"
	"fmt"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type GFWList struct {
	UpdateTime *time.Time
	Url        string
	Tag        string
	Size       uint
	sync.Mutex
}

var g GFWList

func GetRemoteGFWListUpdateTime(c *http.Client) (gfwlist GFWList, err error) {
	g.Lock()
	defer g.Unlock()
	if g.UpdateTime != nil {
		return g, nil
	}
	resp, err := httpClient.HttpGetUsingSpecificClient(c, "https://api.github.com/repos/Loyalsoldier/v2ray-rules-dat/releases/latest")
	if err != nil {
		err = errors.New("获取GFWList最新版本时间失败")
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var (
		t    time.Time
		u    string
		size uint
	)
	t, err = time.Parse(time.RFC3339, gjson.GetBytes(b, "published_at").String())
	arr := gjson.GetBytes(b, "assets").Array()
	for _, item := range arr {
		if item.Get("name").String() == "geosite.dat" {
			size = uint(item.Get("size").Uint())
			u = item.Get("browser_download_url").String()
			break
		}
	}
	if size == 0 && err == nil {
		err = errors.New("fail to get filesize")
	}
	g.Tag = gjson.GetBytes(b, "tag_name").String()
	g.UpdateTime = &t
	g.Url = u
	g.Size = size
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
	id, _ := gonanoid.Nanoid()
	i := 0
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	u := fmt.Sprintf(`https://cdn.jsdelivr.net/gh/mzz2017/dist-v2ray-rules-dat@%v/geosite.dat`, gfwlist.Tag)
	for {
		if i > 2 {
			break
		}
		err = download.Pget(u, "/tmp/LoyalsoldierSite.dat."+id)
		log.Printf("download[%v]: %v\n", i+1, u)
		if err != nil {
			//最多重试2次
			i++
			continue
		}
		break
	}
	if err != nil {
		log.Println(err)
		return
	}
	err = copyfile.CopyFile("/tmp/LoyalsoldierSite.dat."+id, asset.GetV2rayLocationAsset()+"/LoyalsoldierSite.dat")
	if err != nil {
		return
	}
	err = os.Chmod(asset.GetV2rayLocationAsset()+"/LoyalsoldierSite.dat", os.FileMode(0755))
	if err != nil {
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
			"目前最新版本为" + tRemote.Format("2006-01-02") + "，您的本地文件已最新，无需更新",
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
