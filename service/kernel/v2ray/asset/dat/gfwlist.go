package dat

import (
	libSha256 "crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/files"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"google.golang.org/protobuf/proto"
)

type GFWList struct {
	UpdateTime time.Time
	Tag        string
}

var g GFWList
var gFetchedAt time.Time
var gMutex sync.Mutex

func GetRemoteGFWListUpdateTime(c *http.Client) (gfwlist GFWList, err error) {
	const host = "api.github.com"
	status := ""
	// The reason the request never got an answer, without our own wrapping.
	detail := ""
	defer func() {
		if err != nil {
			var coded *common.CodedError
			if !errors.As(err, &coded) {
				if status == "" {
					// Nothing was answered, so there is no HTTP status to
					// report; give the reason instead.
					if detail == "" {
						detail = err.Error()
					}
					err = common.Coded("ASSET_UNREACHABLE", err, map[string]interface{}{
						"host":   host,
						"detail": detail,
					})
				} else {
					err = common.Coded("ASSET_DOWNLOAD_FAILED", err, map[string]interface{}{
						"host":   host,
						"status": status,
					})
				}
			}
		}
	}()

	gMutex.Lock()
	defer gMutex.Unlock()
	if !g.UpdateTime.IsZero() && time.Since(gFetchedAt) < time.Hour {
		return g, nil
	}
	resp, err := httpClient.HttpGetUsingSpecificClient(c, "https://api.github.com/repos/v2rayA/dist-v2ray-rules-dat/tags")
	if err != nil {
		detail = unwrappedReason(err)
		err = fmt.Errorf("failed to get latest version of GFWList: %w", err)
		return
	}
	status = resp.Status
	b, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	tag := gjson.GetBytes(b, "0.name").Str
	if tag == "" {
		err = fmt.Errorf("GitHub returned no GFWList tag list (HTTP %s); it may be rate-limiting this IP, try again later or use a custom download link", resp.Status)
		return
	}
	t, err := time.Parse("200601021504", tag)
	if err != nil {
		err = fmt.Errorf("latest GFWList tag %q has an invalid timestamp: %w", tag, err)
		return
	}
	g.Tag = tag
	g.UpdateTime = t
	gFetchedAt = time.Now()
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

func checkSha256(p string, sha256 string) (bool, string) {
	if b, err := os.ReadFile(p); err == nil {
		hash := libSha256.Sum256(b)
		actual := hex.EncodeToString(hash[:])
		return actual == sha256, actual
	}
	return false, ""
}

var (
	FailCheckSha = fmt.Errorf("could not fetch checksum")
	DamagedFile  = fmt.Errorf("downloaded file checksum does not match")
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

func installGeoSiteFile(downloadedPath, targetPath string) error {
	data, err := os.ReadFile(downloadedPath)
	if err != nil {
		return err
	}
	var sites routercommon.GeoSiteList
	if err := proto.Unmarshal(data, &sites); err != nil {
		return fmt.Errorf("downloaded GFWList is not a valid geosite database: %w", err)
	}
	if len(sites.Entry) == 0 {
		return fmt.Errorf("downloaded GFWList contains no geosite entries")
	}
	return os.Rename(downloadedPath, targetPath)
}

func UpdateLocalGFWListByCustomLink(downloadLink string) (localGFWListVersionAfterUpdate string, err error) {
	pathSiteDat, err := asset.GetV2rayLocationAsset("LoyalsoldierSite.dat")
	if err != nil {
		return "", err
	}
	if err = asset.Download(downloadLink, pathSiteDat+".new"); err != nil {
		log.Warn("UpdateLocalGFWList: %v", err)
		return "", err
	}
	gfwlist, err := GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err == nil {
		_ = os.Chtimes(pathSiteDat+".new", gfwlist.UpdateTime, gfwlist.UpdateTime)
	}
	t, err := files.GetFileModTime(pathSiteDat + ".new")
	if err != nil {
		return "", err
	}
	localGFWListVersionAfterUpdate = t.Local().Format("2006-01-02")
	if err := installGeoSiteFile(pathSiteDat+".new", pathSiteDat); err != nil {
		return "", err
	}
	log.Info("download: %v -> SUCCESS\n", downloadLink)
	return localGFWListVersionAfterUpdate, nil
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
	u := fmt.Sprintf(`https://github.com/v2rayA/dist-v2ray-rules-dat/raw/%v/geosite.dat`, gfwlist.Tag)
	if err = asset.Download(u, pathSiteDat+".new"); err != nil {
		log.Warn("UpdateLocalGFWList: %v", err)
		return
	}
	u2 := fmt.Sprintf(`https://github.com/v2rayA/dist-v2ray-rules-dat/raw/%v/geosite.dat.sha256sum`, gfwlist.Tag)
	siteDatSha256, err := httpGet(u2)
	if err != nil {
		err = common.Coded("ASSET_DOWNLOAD_FAILED", fmt.Errorf("%w for GFWList: %w", FailCheckSha, err), map[string]interface{}{
			"host":   "github.com",
			"status": "",
		})
		log.Warn("UpdateLocalGFWList: %v", err)
		return "", err
	}
	var sha256 string
	if fields := strings.Fields(siteDatSha256); len(fields) != 0 {
		sha256 = fields[0]
	}
	if ok, actual := checkSha256(pathSiteDat+".new", sha256); !ok {
		err = fmt.Errorf("%w for GFWList (expected %s, got %s); try again", DamagedFile, sha256, actual)
		log.Warn("UpdateLocalGFWList: sha mismatch, expected %s, got %s", sha256, actual)
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

// ErrGFWListUpToDate reports that the local GFWList already matches the latest
// release, so no download was needed.
var ErrGFWListUpToDate = errors.New("GFWList is already up to date")

func CheckAndUpdateGFWList(downloadLink string) (localGFWListVersionAfterUpdate string, err error) {
	if downloadLink == "" {
		update, tRemote, err := IsGFWListUpdate()
		if err != nil {
			return "", err
		}
		if update {
			// Nothing to download. This is not a failure, and reporting it as
			// one made the GUI say "could not update GFWList: GFWList is up to
			// date"; the caller decides how to present it.
			return tRemote.Local().Format("2006-01-02"), ErrGFWListUpToDate
		}

		/* 更新LoyalsoldierSite.dat */
		localGFWListVersionAfterUpdate, err = UpdateLocalGFWList()
		if err != nil {
			return "", err
		}
	} else {
		/* 手动更新LoyalsoldierSite.dat */
		localGFWListVersionAfterUpdate, err = UpdateLocalGFWListByCustomLink(downloadLink)
		if err != nil {
			return "", err
		}
	}

	setting := configure.GetSettingNotNil()
	if v2ray.ProcessManager.Running() && //正在使用GFWList模式再重启
		(setting.Transparent == configure.TransparentGfwlist ||
			!v2ray.IsTransparentOn(setting) && setting.RulePortMode == configure.GfwlistMode) {
		err = v2ray.UpdateV2RayConfig()
	}
	return
}

func DeleteGFWList() error {
	pathSiteDat, err := asset.GetV2rayLocationAsset("LoyalsoldierSite.dat")
	if err != nil {
		return err
	}
	// Deleting what is already gone is the result the caller asked for; it
	// used to answer "remove …: no such file or directory".
	if err := os.Remove(pathSiteDat); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// unwrappedReason returns the innermost cause of err, which is what a user can
// act on: the DNS or connection failure, not the layers that wrap it.
func unwrappedReason(err error) string {
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err.Error()
		}
		err = inner
	}
}
