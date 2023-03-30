package report

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CurlReporter struct {
}

var DefaultCurlReporter CurlReporter

func (r *CurlReporter) PresetPortReport() (ok bool, report string) {
	defer func() {
		report = "Preset Port(socks5): " + report
	}()
	ports := configure.GetPortsNotNil()
	if ports.Socks5 == 0 {
		return false, "Preset HTTP Port is closed"
	}
	cli, err := httpClient.GetHttpClientWithProxy("socks5://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(ports.Socks5)))
	if err != nil {
		return false, err.Error()
	}
	resp, err := cli.Get("https://www.apple.com")
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			resp.Body.Close()
			return false, resp.Status
		}
		return false, err.Error()
	}
	resp.Body.Close()
	return true, resp.Status
}

func (r *CurlReporter) TransparentReport() (ok bool, report string) {
	setting := configure.GetSettingNotNil()
	defer func() {
		report = fmt.Sprintf("Transparent Proxy(%v): %v", setting.TransparentType, report)
	}()
	if !v2ray.IsTransparentOn(nil) {
		return true, "Transparent Proxy is not enabled"
	}
	cli := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := cli.Get("https://ipv4.appspot.com/")
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			resp.Body.Close()
			return false, resp.Status
		}
		return false, err.Error()
	}
	b, _ := io.ReadAll(resp.Body)
	ip := strings.TrimSpace(string(b))
	if net.ParseIP(ip) == nil {
		return false, "UNKNOWN PROBLEM"
	}
	return true, "Your remote IP is: " + ip
}
