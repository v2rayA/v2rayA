package httpClient

import "net/http"

func HttpGetUsingSpecificClient(c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	//FIXME: shadowrocket会有可能不清楚alterid的情况，在后端为soga时aid不一致将无法连接
	//req.Header.Set("User-Agent", "v2rayA (like shadowrocket)")
	if resp, err = c.Do(req); err != nil {
		resp, err = http.DefaultClient.Do(req)
	}
	return
}
