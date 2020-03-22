package httpClient

import "net/http"

func HttpGetUsingSpecificClient(c *http.Client, url string) (resp *http.Response, err error) {
	if resp, err = c.Get(url); err != nil {
		resp, err = http.Get(url)
	}
	return
}
