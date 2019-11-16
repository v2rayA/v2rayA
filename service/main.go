package main

import (
	"V2RayA/global"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/router"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"log"
	"os"
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
	color.Red.Println("V2RAY_LOCATION_ASSET is:", global.V2RAY_LOCATION_ASSET)
	if !configure.IsConfigureExists() {
		err := configure.SetConfigure(configure.New())
		if err != nil {
			log.Fatal(err)
		}
	}
	if global.ServiceControlMode != global.DockerMode {
		wd, _ := os.Getwd()
		color.Red.Println("Service working directory is:", wd)
		conf := global.GetServiceConfig()
		color.Red.Println("Service listen: http://"+conf.Address+":"+conf.Port+",", "GUI demo: https://v2raya.mzz.pub")

		//如果V2Ray正在运行，而配置文件中没有记录当前连接的节点是谁，就关掉V2Ray
		if v2ray.IsV2RayRunning() && configure.GetConnectedServer() == nil {
			err := v2ray.StopV2rayService()
			if err != nil {
				log.Fatal(err)
			}
			err = v2ray.DisableV2rayService()
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		fmt.Println("V2RayA is running in Docker. Compatibility mode starts up.") //TODO
	}
	router.Run()
}
