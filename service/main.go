package main

import (
	"github.com/gin-gonic/gin"
	"log"
	_ "v2rayA/plugins/pingtunnel"
	_ "v2rayA/plugins/shadowsocksr"
	_ "v2rayA/plugins/trojan"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	checkTProxySupportability()
	initConfigure()
	checkConnection()
	go checkUpdate()
	hello()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
