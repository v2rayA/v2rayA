package main

import (
	"V2RayA/config"
	"V2RayA/router"
	"V2RayA/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/exec"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	wd, _ := os.Getwd()
	fmt.Println("Working directory is", wd)
	_, err := exec.Command("ls", "/.dockerenv").Output()
	if err != nil {
		fmt.Println("检测到未运行在docker环境中，请注意会有以下问题:")
		fmt.Println("1. 本程序会修改v2ray和privoxy配置文件")
		fmt.Println("2. 本程序生成的v2ray配置文件只占用1080端口，如果1080端口已被占用将会出错，如需修改端口请按照v2ray文档修改models/template.go的templateJson变量")
		fmt.Println("3. 如果程序未以root权限运行，可能会运行失败")
		fmt.Println("强烈建议在生产环境下将程序运行于docker中! 具体方法参见: https://github.com/mzz2017/V2RayA/blob/master/README.md")
	} else {
		fmt.Println("Running in Docker")
	}
	if tools.IsV2RayRunning() && config.GetTouchRaw().ConnectedServer == nil {
		err = tools.StopV2rayService()
		if err != nil {
			log.Fatal(err)
		}
		err = tools.DisableV2rayService()
		if err != nil {
			log.Fatal(err)
		}
	}
	router.Run()
}
