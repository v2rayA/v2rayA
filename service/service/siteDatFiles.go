package service

import (
	"github.com/golang/protobuf/proto"
	"github.com/v2rayA/v2rayA/core/siteDat"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"log"
	"os"
	"path"
	"strings"
	"v2ray.com/core/app/router"
)

func GetSiteDatFiles() (siteDats []siteDat.SiteDat) {
	dir := asset.GetV2rayLocationAsset()
	fis, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, f := range fis {
		if f.IsDir() {
			continue
		}
		if path.Ext(strings.ToLower(f.Name())) == ".dat" {
			if f.Name() == "geoip.dat" {
				//暂不支持IPDat
				continue
			}
			var sd siteDat.SiteDat
			sd.Filename = f.Name()
			b, err := os.ReadFile(path.Join(dir, f.Name()))
			if err != nil {
				log.Println(err)
				continue
			}
			var siteList router.GeoSiteList
			err = proto.Unmarshal(b, &siteList)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, e := range siteList.Entry {
				sd.Tags = append(sd.Tags, strings.ToLower(e.CountryCode))
			}
			siteDats = append(siteDats, sd)
		}
	}
	return
}
