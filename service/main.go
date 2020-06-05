package main

import (
	"github.com/gin-gonic/gin"
	"log"
	_ "v2rayA/plugin/pingtunnel"
	_ "v2rayA/plugin/shadowsocksr"
	_ "v2rayA/plugin/trojan"
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
