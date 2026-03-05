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

	// 尝试作为服务运行（在 Windows 上有实现，其他平台返回 false）
	if tryRunAsService() {
		return
	}

	// 正常启动（非服务模式）
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
