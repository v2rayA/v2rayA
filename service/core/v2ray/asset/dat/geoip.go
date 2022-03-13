package dat

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os"
	"strings"
)

func UpdateLocalGeoIP() (err error) {
	pathSiteDat, err := asset.GetV2rayLocationAsset("geoip.dat")
	if err != nil {
		return err
	}
	if err = asset.Download("https://hubmirror.v2raya.org/v2fly/geoip/releases/latest/download/geoip.dat", pathSiteDat+".new"); err != nil {
		log.Warn("UpdateLocalGeoIP: %v", err)
		return
	}
	siteDatSha256, err := httpGet("https://hubmirror.v2raya.org/v2fly/geoip/releases/latest/download/geoip.dat.sha256sum")
	if err != nil {
		err = fmt.Errorf("%w: %v", FailCheckSha, err)
		log.Warn("UpdateLocalGeoIP: %v", err)
		return err
	}
	if !checkSha256(pathSiteDat+".new", strings.Fields(siteDatSha256)[0]) {
		err = fmt.Errorf("UpdateLocalGeoIP: %v", DamagedFile)
		return
	}
	if err := os.Rename(pathSiteDat+".new", pathSiteDat); err != nil {
		return err
	}
	log.Info("download[geoip.dat]: SUCCESS\n")
	return
}
