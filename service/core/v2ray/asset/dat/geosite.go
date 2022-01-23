package dat

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	gopeed2 "github.com/v2rayA/v2rayA/pkg/util/gopeed"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os"
	"path/filepath"
	"strings"
)

func UpdateLocalGeoSite() (err error) {
	assetDir := asset.GetV2rayLocationAsset()
	pathSiteDat := filepath.Join(assetDir, "geosite.dat")
	if err = gopeed2.Down(&gopeed2.Request{
		Method: "GET",
		URL:    "https://hubmirror.v2raya.org/v2fly/domain-list-community/releases/latest/download/dlc.dat",
	}, pathSiteDat+".new"); err != nil {
		log.Warn("UpdateLocalGeoSite: %v", err)
		return
	}
	siteDatSha256, err := httpGet("https://hubmirror.v2raya.org/v2fly/domain-list-community/releases/latest/download/dlc.dat.sha256sum")
	if err != nil {
		err = fmt.Errorf("%w: %v", FailCheckSha, err)
		log.Warn("UpdateLocalGeoSite: %v", err)
		return err
	}
	if !checkSha256(pathSiteDat+".new", strings.Fields(siteDatSha256)[0]) {
		err = fmt.Errorf("UpdateLocalGeoSite: %v", DamagedFile)
		return
	}
	if err := os.Rename(pathSiteDat+".new", pathSiteDat); err != nil {
		return err
	}
	log.Info("download[geosite.dat]: SUCCESS\n")
	return
}
