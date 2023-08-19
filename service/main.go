package main

import (
	"runtime"

	"github.com/gin-gonic/gin"
	_ "github.com/v2rayA/v2rayA/conf/report"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/juicity"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/simpleobfs"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/socks5"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ss"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ssr"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/tcp"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/tls"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/trojanc"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/tuic"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ws"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	if runtime.GOOS == "linux" {
		checkTProxySupportability()
	}
	initConfigure()
	checkUpdate()
	hello()
	if err := run(); err != nil {
		log.Fatal("main: %v", err)
	}
}
