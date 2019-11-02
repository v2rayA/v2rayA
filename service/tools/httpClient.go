package tools

import (
	"V2RayA/global"
	"V2RayA/models/touch"
	"V2RayA/proxyWithHttp"
	"net/http"
	"net/url"
)

func GetHttpClientWithProxy(proxyURL string) (client *http.Client, err error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return
	}
	dialer, err := proxyWithHttp.FromURL(u, proxyWithHttp.Direct)
	if err != nil {
		return
	}
	httpTransport := &http.Transport{}
	httpTransport.Dial = dialer.Dial
	client = &http.Client{Transport: httpTransport}
	return
}

func GetHttpClientWithV2RayAProxy() (client *http.Client, err error) {
	return GetHttpClientWithProxy("socks5://localhost:10800")
}

func GetHttpClientWithV2RayAPac() (client *http.Client, err error) {
	return GetHttpClientWithProxy("http://localhost:10802")
}

func GetHttpClientAutomatically() (c *http.Client, err error) {
	switch global.GetTouchRaw().Setting.ProxyModeWhenSubscribe {
	case touch.ProxyModePac:
		c, err = GetHttpClientWithV2RayAPac()
	case touch.ProxyModeProxy:
		c, err = GetHttpClientWithV2RayAProxy()
	default:
		c = http.DefaultClient
	}
	return
}
