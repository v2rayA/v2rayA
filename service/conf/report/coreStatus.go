package report

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"net"
	"strconv"
	"strings"
	"time"
)

type CoreStatusReporter struct {
}

var DefaultCoreStatusReporter CoreStatusReporter

func (r *CoreStatusReporter) FromDatabase() (ok bool, report string) {
	running := configure.GetRunning()
	if !running {
		return false, "Core Running Status(from database): v2ray-core is not running"
	}
	return true, "Core Running Status(from database): v2ray-core is running"
}

func (r *CoreStatusReporter) FromApiListening() (ok bool, report string) {
	defer func() {
		report = "Core Running Status(from api listening): " + report
	}()
	b, err := asset.GetConfigBytes()
	if err != nil {
		return false, fmt.Sprintf("failed to read config: %v", err)
	}
	var t v2ray.Template
	_ = jsoniter.Unmarshal(b, &t)
	var apiPort int
	for _, inbound := range t.Inbounds {
		if strings.HasPrefix(inbound.Tag, "api-in") {
			apiPort = inbound.Port
			break
		}
	}
	if apiPort == 0 {
		return false, "cannot get api port from config file"
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(apiPort)), 5*time.Second)
	if err != nil {
		return false, "v2ray-core is NOT running"
	}
	conn.Close()
	return true, "v2ray-core is running"
}
