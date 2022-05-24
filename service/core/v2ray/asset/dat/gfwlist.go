package dat

import (
	libSha256 "crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"io"
	"net/http"
	"os"
	"strings"
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
	resp, err := httpClient.HttpGetUsingSpecificClient(c, "https://hubmirror.v2raya.org/api/v2rayA/dist-v2ray-rules-dat/tags")
	if err != nil {
		err = fmt.Errorf("failed to get latest version of GFWList: %w", err)
		return
	}
	b, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	tag := gjson.GetBytes(b, "0.name").Str
	t, err := time.Parse("200601021504", tag)
	if err != nil {
		err = fmt.Errorf("failed to get latest version of GFWList: fail in getting commit date of latest tag: %w", err)
		return
	}
	g.Tag = tag
	g.UpdateTime = t
	return g, nil
}
func IsGFWListUpdate() (update bool, remoteTime time.Time, err error) {
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	remoteTime = gfwlist.UpdateTime
	if !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
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

func checkSha256(p string, sha256 string) bool {
	if b, err := os.ReadFile(p); err == nil {
		hash := libSha256.Sum256(b)
		return hex.EncodeToString(hash[:]) == sha256
	} else {
		return false
	}
}

var (
	FailCheckSha = fmt.Errorf("failed to check sum256sum of GFWList file")
	DamagedFile  = fmt.Errorf("damaged GFWList file, update it again please")
)

func httpGet(url string) (data string, err error) {
	resp, err := httpClient.GetHttpClientAutomatically().Get(url)
	if err != nil {
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func UpdateLocalGFWList() (localGFWListVersionAfterUpdate string, err error) {
	i := 0
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		return
	}
	pathSiteDat, err := asset.GetV2rayLocationAsset("LoyalsoldierSite.dat")
	if err != nil {
		return "", err
	}
	u := fmt.Sprintf(`https://hubmirror.v2raya.org/v2rayA/dist-v2ray-rules-dat/raw/%v/geosite.dat`, gfwlist.Tag)
	if err = asset.Download(u, pathSiteDat+".new"); err != nil {
		log.Warn("UpdateLocalGFWList: %v", err)
		return
	}
	u2 := fmt.Sprintf(`https://hubmirror.v2raya.org/v2rayA/dist-v2ray-rules-dat/raw/%v/geosite.dat.sha256sum`, gfwlist.Tag)
	siteDatSha256, err := httpGet(u2)
	if err != nil {
		err = fmt.Errorf("%w: %v", FailCheckSha, err)
		log.Warn("UpdateLocalGFWList: %v", err)
		return "", err
	}
	var sha256 string
	if fields := strings.Fields(siteDatSha256); len(fields) != 0 {
		sha256 = fields[0]
	}
	if !checkSha256(pathSiteDat+".new", sha256) {
		err = fmt.Errorf("UpdateLocalGFWList: %v", DamagedFile)
		return
	}
	_ = os.Chtimes(pathSiteDat+".new", gfwlist.UpdateTime, gfwlist.UpdateTime)
	t, err := files.GetFileModTime(pathSiteDat + ".new")
	if err == nil {
		localGFWListVersionAfterUpdate = t.Local().Format("2006-01-02")
	}
	if err := os.Rename(pathSiteDat+".new", pathSiteDat); err != nil {
		return "", err
	}
	log.Info("download[%v]: %v -> SUCCESS\n", i+1, u)
	return
}

func CheckAndUpdateGFWList() (localGFWListVersionAfterUpdate string, err error) {
	update, tRemote, err := IsGFWListUpdate()
	if err != nil {
		return
	}
	if update {
		return "", fmt.Errorf(
			"latest version is " + tRemote.Local().Format("2006-01-02") + ". GFWList is up to date",
		)
	}

	/* 更新LoyalsoldierSite.dat */
	localGFWListVersionAfterUpdate, err = UpdateLocalGFWList()
	if err != nil {
		return
	}
	setting := configure.GetSettingNotNil()
	if v2ray.ProcessManager.Running() && //正在使用GFWList模式再重启
		(setting.Transparent == configure.TransparentGfwlist ||
			!v2ray.IsTransparentOn() && setting.RulePortMode == configure.GfwlistMode) {
		err = v2ray.UpdateV2RayConfig()
	}
	return
}
