package service

import (
	"V2RayA/model/siteDat"
	"V2RayA/model/v2ray"
	"github.com/gogo/protobuf/proto"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"v2ray.com/core/app/router"
)

func GetSiteDatFiles() (siteDats []siteDat.SiteDat) {
	dir := v2ray.GetV2rayLocationAsset()
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, f := range fis {
		if f.IsDir() {
			continue
		}
		if path.Ext(strings.ToLower(f.Name())) == ".dat" {
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
