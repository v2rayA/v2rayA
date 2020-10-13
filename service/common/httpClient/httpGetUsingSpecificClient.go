package httpClient

import "net/http"

func HttpGetUsingSpecificClient(c *http.Client, url string) (resp *http.Response, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "v2rayA (like shadowrocket)")
	if resp, err = c.Do(req); err != nil {
		resp, err = http.DefaultClient.Do(req)
	}
	return
}
