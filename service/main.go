package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/v2rayA/v2rayA/conf/report"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/anytls"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/http"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/hysteria2"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/juicity"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/simpleobfs"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/socks5"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ss"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ssr"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/tcp"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/tls"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/trojan"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/trojanc"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/tuic"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ws"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func main() {
	println("[DEBUG] main.main started")
	gin.SetMode(gin.ReleaseMode)

	// Try running as a service (implemented on Windows, returns false on other platforms)
	if tryRunAsService() {
		return
	}

	// Normal startup (non-service mode)
	checkEnvironment()
	if err := checkPlatformSpecific(); err != nil {
		log.Fatal("Platform check failed: %v", err)
	}
	initConfigure()
	checkUpdate()
	hello()
	if err := run(); err != nil {
		log.Fatal("main: %v", err)
	}
}
