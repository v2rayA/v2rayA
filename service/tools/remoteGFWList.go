package tools

import (
	"github.com/PuerkitoBio/goquery"
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
	timeRaw, _ := doc.Find("relative-time").First().Attr("datetime")
	return time.Parse(time.RFC3339, timeRaw)
}
