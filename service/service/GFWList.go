package service

import (
	"V2RayA/extra/copyfile"
	"V2RayA/extra/quickdown"
	"V2RayA/global"
	"V2RayA/model/v2ray"
	"V2RayA/tools"
	"github.com/PuerkitoBio/goquery"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func GetRemoteGFWListUpdateTime(c *http.Client) (t time.Time, err error) {
	resp, err := tools.HttpGetUsingCertainClient(c, "https://github.com/ToutyRater/V2Ray-SiteDAT/contributors/master/geofiles/h2y.dat")
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	timeRaw, ok := doc.Find("relative-time").First().Attr("datetime")
	if !ok {
		log.Println(doc.Html())
		return time.Time{}, errors.New("获取最新GFWList版本失败")
	}
	log.Println("timeraw", timeRaw)
	return time.Parse(time.RFC3339, timeRaw)
}
func IsUpdate() (update bool, remoteTime time.Time, err error) {
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		return
	}
	remoteTime, err = GetRemoteGFWListUpdateTime(c)
	if err != nil {
		return
	}

	if !v2ray.IsH2yExists() {
		//本地文件不存在，那远端必定比本地新
		return false, remoteTime, nil
	}
	//本地文件存在，检查本地版本是否比远端还新
	t, err := v2ray.GetH2yModTime()
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
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		return
	}
	quickdown.SetHttpClient(c)
	id, _ := gonanoid.Nanoid()
	i := 0
	for {
		err = quickdown.DownloadWithWorkersTo("https://github.com/ToutyRater/V2Ray-SiteDAT/raw/master/geofiles/h2y.dat", 10, "/tmp/"+id)
		if err != nil && i < 3 && strings.Contains(err.Error(), "head fail") {
			//建立连接问题，最多重试3次
			i++
			continue
		}
		break
	}
	if err != nil {
		return
	}
	err = copyfile.CopyFile("/tmp/"+id, global.V2RAY_LOCATION_ASSET+"/h2y.dat")
	if err != nil {
		return
	}
	err = os.Chmod(global.V2RAY_LOCATION_ASSET+"/h2y.dat", os.FileMode(0755))
	if err != nil {
		return
	}
	t, err := tools.GetFileModTime(global.V2RAY_LOCATION_ASSET + "/h2y.dat")
	if err == nil {
		localGFWListVersionAfterUpdate = t.Format("2006-01-02")
	}
	return
}
