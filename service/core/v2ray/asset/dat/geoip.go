package dat

import (
	"fmt"
	"os"
	"strings"

	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func UpdateLocalGeoIP() (err error) {
	pathSiteDat, err := asset.GetV2rayLocationAsset("geoip.dat")
	if err != nil {
		return err
	}
	if err = asset.Download("https://github.com/v2fly/geoip/releases/latest/download/geoip.dat", pathSiteDat+".new"); err != nil {
		log.Warn("UpdateLocalGeoIP: %v", err)
		return
	}
	siteDatSha256, err := httpGet("https://github.com/v2fly/geoip/releases/latest/download/geoip.dat.sha256sum")
	if err != nil {
		err = fmt.Errorf("%w: %v", FailCheckSha, err)
		log.Warn("UpdateLocalGeoIP: %v", err)
		return err
	}
	var sha256 string
	if fields := strings.Fields(siteDatSha256); len(fields) != 0 {
		sha256 = fields[0]
	}
	if ok, actual := checkSha256(pathSiteDat+".new", sha256); !ok {
		err = fmt.Errorf("UpdateLocalGeoIP: %v (expected %s, got %s)", DamagedFile, sha256, actual)
		log.Warn("UpdateLocalGeoIP: sha mismatch, expected %s, got %s", sha256, actual)
		return
	}
	if err := os.Rename(pathSiteDat+".new", pathSiteDat); err != nil {
		return err
	}
	log.Info("download[geoip.dat]: SUCCESS\n")
	return
}
