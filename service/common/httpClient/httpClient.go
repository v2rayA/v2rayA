package httpClient

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

func GetHttpClientWithProxy(proxyURL string) (client *http.Client, err error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return
	}
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(u)}}, nil
}

func GetHttpClientWithv2rayAProxy() (client *http.Client, err error) {
	host := "127.0.0.1"
	//是否在docker环境
	if common.IsDocker() {
		//连接网关，即宿主机的端口
		out, err := exec.Command("sh", "-c", "ip route list default|head -n 1|awk '{print $3}'").Output()
		if err == nil {
			host = strings.TrimSpace(string(out))
		} else {
			return nil, fmt.Errorf("failed to get gateway: %v", err)
		}
	}
	return GetHttpClientWithProxy("socks5://" + net.JoinHostPort(host, strconv.Itoa(configure.GetPortsNotNil().Socks5)))
}

func GetHttpClientWithv2rayAPac() (client *http.Client, err error) {
	host := "127.0.0.1"
	//是否在docker环境
	if common.IsDocker() {
		//连接网关，即宿主机的端口
		out, err := exec.Command("sh", "-c", "ip route|grep default|awk '{print $3}'").Output()
		if err == nil {
			host = strings.TrimSpace(string(out))
		} else {
			return nil, fmt.Errorf("failed to get gateway: %v", err)
		}
	}
	return GetHttpClientWithProxy("http://" + net.JoinHostPort(host, strconv.Itoa(configure.GetPortsNotNil().HttpWithPac)))
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
