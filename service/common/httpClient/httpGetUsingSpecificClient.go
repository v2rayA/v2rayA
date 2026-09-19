package httpClient

import (
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"

	"github.com/v2rayA/v2rayA/conf"
)

func HttpGetUsingSpecificClient(c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	//shadowrocket会有可能不清楚alterid的情况，影响aead是否启用的问题
	req.Header.Set("User-Agent", fmt.Sprintf("v2rayA/%v WebRequestHelper", conf.Version))
	if resp, err = c.Do(req); err != nil {
		fallback := *http.DefaultClient
		if c.Timeout > 0 && (fallback.Timeout == 0 || c.Timeout < fallback.Timeout) {
			fallback.Timeout = c.Timeout
		}
		resp, err = fallback.Do(req)
	}
	if resp == nil {
		cause := err
		var urlErr *neturl.Error
		if errors.As(cause, &urlErr) {
			cause = urlErr.Err
		}
		if cause == nil {
			cause = errors.New("request returned no response")
		}
		return nil, fmt.Errorf("could not reach %s: %w", req.URL.Hostname(), cause)
	}
	return
}
