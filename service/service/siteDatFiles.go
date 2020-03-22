package service

import (
	"V2RayA/core/siteDat"
	"V2RayA/core/v2ray/asset"
	"github.com/gogo/protobuf/proto"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"v2ray.com/core/app/router"
)

func GetSiteDatFiles() (siteDats []siteDat.SiteDat) {
	dir := asset.GetV2rayLocationAsset()
	fis, err := ioutil.ReadDir(dir)
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
			b, err := ioutil.ReadFile(path.Join(dir, f.Name()))
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
