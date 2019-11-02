package main

import (
	"V2RayA/config"
	"V2RayA/router"
	"V2RayA/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	if runtime.GOOS == "windows" {
		fmt.Println("windows不支持直接运行，请配合docker使用。见https://github.com/mzz2017/V2RayA")
		fmt.Println("请按任意键继续...")
		_, _ = fmt.Scanf("\n")
		return
	}
	_, err := exec.Command("ls", "/.dockerenv").Output()
	if os.Getegid() != 0 {
		fmt.Println("请以sudo或root权限执行本程序")
		return
	}
	if err != nil {
		color.BgLightRed.Println("检测到未运行在docker环境中，请注意会有以下问题:")
		color.Warn.Println("1. 本程序会修改v2ray配置文件")
		color.Warn.Println("2. 本程序生成的v2ray配置文件将占用10800(socks)、10801(http)、10802(http)端口，如果端口已被占用将会出错，如需修改端口请按照v2ray文档修改models/template.go的templateJson变量")
		color.Warn.Println("强烈建议在生产环境下将程序运行于docker中! 具体方法参见: https://github.com/mzz2017/V2RayA/blob/master/README.md")
		fmt.Println("============================================================")
		wd, _ := os.Getwd()
		color.Red.Println("Working directory is:", wd)
		conf := config.GetServiceConfig()
		color.Red.Println("Configuration file is at:", conf.ConfigPath)
		color.Red.Println("Service listen: "+conf.Address+":"+conf.Port+",", "GUI demo: https://v2raya.mzz.pub")
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
