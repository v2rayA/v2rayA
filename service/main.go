package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/v2rayA/v2rayA/plugin/pingtunnel"
	_ "github.com/v2rayA/v2rayA/plugin/shadowsocksr"
	_ "github.com/v2rayA/v2rayA/plugin/simpleobfs"
	_ "github.com/v2rayA/v2rayA/plugin/ssrpluginSimpleobfs"
	_ "github.com/v2rayA/v2rayA/plugin/trojan"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	checkTProxySupportability()
	initConfigure()
	go checkUpdate()
	hello()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
