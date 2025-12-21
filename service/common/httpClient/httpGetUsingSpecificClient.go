package httpClient

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"net/http"
)

func HttpGetUsingSpecificClient(c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	//shadowrocket会有可能不清楚alterid的情况，影响aead是否启用的问题
	req.Header.Set("User-Agent", fmt.Sprintf("v2rayA/%v WebRequestHelper", conf.Version))
	if resp, err = c.Do(req); err != nil {
		resp, err = http.DefaultClient.Do(req)
	}
	return
}

// HttpGetSubscriptionWithClient performs HTTP GET request for subscription with HWID headers
func HttpGetSubscriptionWithClient(c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	
	// Set User-Agent
	req.Header.Set("User-Agent", fmt.Sprintf("v2rayA/%v WebRequestHelper", conf.Version))
	
	// Get HWID and system info
	hwid := conf.GetHWID()
	sysInfo := common.GetSystemInfo()
	
	// Set HWID headers
	req.Header.Set("x-hwid", hwid)
	req.Header.Set("x-device-os", sysInfo.DeviceOS)
	req.Header.Set("x-ver-os", sysInfo.VersionOS)
	req.Header.Set("x-device-model", sysInfo.DeviceModel)
	
	if resp, err = c.Do(req); err != nil {
		resp, err = http.DefaultClient.Do(req)
	}
	return
}
