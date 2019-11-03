package main

import (
	"V2RayA/global"
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
	if os.Getegid() != 0 {
		fmt.Println("请以sudo或root权限执行本程序")
		return
	}
	_, err := exec.Command("ls", "/.dockerenv").Output()
	if err != nil {
		wd, _ := os.Getwd()
		color.Red.Println("Working directory is:", wd)
		conf := global.GetServiceConfig()
		color.Red.Println("Configuration file is at:", conf.ConfigPath)
		color.Red.Println("Service listen: "+conf.Address+":"+conf.Port+",", "GUI demo: https://v2raya.mzz.pub")

		//如果V2Ray正在运行，而配置文件中没有记录当前连接的节点是谁，就关掉V2Ray
		if tools.IsV2RayRunning() && global.GetTouchRaw().ConnectedServer == nil {
			err = tools.StopV2rayService()
			if err != nil {
				log.Fatal(err)
			}
			err = tools.DisableV2rayService()
			if err != nil {
				log.Fatal(err)
			}
		}

	} else {
		global.IsInDocker = true
		fmt.Println("V2RayA is running in Docker. Compatibility mode starts up.") //TODO
	}
	router.Run()
}
