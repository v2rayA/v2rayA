package httpClient

import (
	"fmt"
	"github.com/v2rayA/v2rayA/global"
	"net/http"
)

func HttpGetUsingSpecificClient(c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	//shadowrocket会有可能不清楚alterid的情况，在后端为soga时aid不一致将无法连接
	//req.Header.Set("User-Agent", "v2rayA (like shadowrocket)")
	req.Header.Set("User-Agent", fmt.Sprintf("v2rayA/%v WebRequestHelper", global.Version))
	if resp, err = c.Do(req); err != nil {
		resp, err = http.DefaultClient.Do(req)
	}
	return
}
