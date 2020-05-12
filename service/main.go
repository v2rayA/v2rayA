package main

import (
	"github.com/gin-gonic/gin"
	"log"
	_ "v2rayA/plugins/pingtunnel"
	_ "v2rayA/plugins/shadowsocksr"
	_ "v2rayA/plugins/trojan"
)

func main() {
	//u, _ := url.Parse("trojan://CLT5Ge@45.93.216.206:8443?allowInsecure=1&peer=jpo206.ovod.me#%E5%85%8D%E8%B4%B9%C2%B7T%C2%B7%E6%97%A5%E6%9C%ACO%C2%B7206")
	//log.Println(u.Scheme)
	//log.Println(u.Hostname())
	//log.Println(u.Port())
	//log.Println(u.User.String())
	//log.Println(u.Query())
	//log.Println(u.Fragment)
	//log.Println(u.Opaque)
	//return
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
