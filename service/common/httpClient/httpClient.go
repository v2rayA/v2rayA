package httpClient

import (
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/core/v2ray"
	"github.com/mzz2017/v2rayA/extra/proxyWithHttp"
	"github.com/mzz2017/v2rayA/db/configure"
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
	httpTransport := &http.Transport{
		Dial: dialer.Dial,
	}
	client = &http.Client{Transport: httpTransport}
	return
}

func GetHttpClientWithv2rayAProxy() (client *http.Client, err error) {
	host := "127.0.0.1"
	//是否在docker环境
	if common.IsInDocker() {
		//连接网关，即宿主机的端口，失败则用同网络下v2ray容器的
		out, err := exec.Command("sh", "-c", "ip route list default|head -n 1|awk '{print $3}'").Output()
		if err == nil {
			host = strings.TrimSpace(string(out))
		} else {
			host = "v2ray"
		}
	}
	return GetHttpClientWithProxy("socks5://" + host + ":20170")
}

func GetHttpClientWithv2rayAPac() (client *http.Client, err error) {
	host := "127.0.0.1"
	//是否在docker环境
	if common.IsInDocker() {
		//连接网关，即宿主机的端口，失败则用同网络下v2ray容器的
		out, err := exec.Command("sh", "-c", "ip route|grep default|awk '{print $3}'").Output()
		if err == nil {
			host = strings.TrimSpace(string(out))
		} else {
			host = "v2ray"
		}
	}
	return GetHttpClientWithProxy("http://" + host + ":20172")
}

func GetHttpClientAutomatically() (c *http.Client, err error) {
	if s := configure.GetSettingNotNil(); !v2ray.IsV2RayRunning() || configure.GetConnectedServer() == nil || s.Transparent != configure.TransparentClose {
		return http.DefaultClient, nil
	}
	setting := configure.GetSettingNotNil()
	switch setting.ProxyModeWhenSubscribe {
	case configure.ProxyModePac:
		c, err = GetHttpClientWithv2rayAPac()
	case configure.ProxyModeProxy:
		c, err = GetHttpClientWithv2rayAProxy()
	default:
		c = http.DefaultClient
	}
	return
}
