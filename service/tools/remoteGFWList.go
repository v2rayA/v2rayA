package tools

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

func GetRemoteGFWListUpdateTime(c *http.Client) (t time.Time, err error) {
	resp, err := HttpGetWithPreference(c, "https://github.com/ToutyRater/V2Ray-SiteDAT/blob/master/geofiles/h2y.dat")
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
