package httpClient

import (
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func GetHttpClientWithProxy(proxyURL string) (client *http.Client, err error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return
	}
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(u)}}, nil
}

// The core runs inside v2rayA's own process tree, in the same network
// namespace, and its proxy inbounds listen on 127.0.0.1 unless port sharing
// is on, so the loopback address reaches them in every environment,
// including Docker with bridge or host networking.

func GetHttpClientWithv2rayAProxy() (client *http.Client, err error) {
	return GetHttpClientWithProxy("socks5://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(configure.GetPortsNotNil().Socks5)))
}

func GetHttpClientWithv2rayAPac() (client *http.Client, err error) {
	return GetHttpClientWithProxy("http://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(configure.GetPortsNotNil().HttpWithPac)))
}

func GetHttpClientAutomatically() (c *http.Client) {
	setting := configure.GetSettingNotNil()
	if !v2ray.ProcessManager.Running() || configure.GetConnectedServers() == nil || v2ray.IsTransparentOn(setting) {
		return http.DefaultClient
	}
	var err error
	switch setting.ProxyModeWhenSubscribe {
	case configure.ProxyModePac:
		c, err = GetHttpClientWithv2rayAPac()
		if err != nil {
			return http.DefaultClient
		}
	case configure.ProxyModeProxy:
		c, err = GetHttpClientWithv2rayAProxy()
		if err != nil {
			return http.DefaultClient
		}
	default:
		c = http.DefaultClient
	}
	return c
}
