package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/pingtunnel"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/shadowsocksr"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/simpleobfs"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/ssrpluginSimpleobfs"
	_ "github.com/v2rayA/v2rayA/pkg/plugin/trojan-go"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	checkTProxySupportability()
	initConfigure()
	checkUpdate()
	hello()
	if err := run(); err != nil {
		log.Fatal("main: %v", err)
	}
}
