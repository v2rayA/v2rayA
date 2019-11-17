package main

import (
	"V2RayA/global"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/persistence/logs"
	"V2RayA/router"
	"V2RayA/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"log"
	"os"
	"runtime"
	"sync"
)

func checkEnvironment() {
	if runtime.GOOS == "windows" {
		fmt.Println("windows不支持直接运行，请配合docker使用。见https://github.com/mzz2017/V2RayA")
		fmt.Println("请按任意键继续...")
		_, _ = fmt.Scanf("\n")
		os.Exit(1)
	}
	if os.Getegid() != 0 {
		log.Fatal("请以sudo或root权限执行本程序")
	}
}

func initConfigure() {
	if !configure.IsConfigureExists() {
		err := configure.SetConfigure(configure.New())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkConnection() {
	//如果V2Ray正在运行，而配置文件中没有记录当前连接的节点是谁，就关掉V2Ray
	if v2ray.IsV2RayRunning() && configure.GetConnectedServer() == nil {
		err := v2ray.StopAndDisableV2rayService()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func hello() {
	color.Red.Println("V2RAY_LOCATION_ASSET is:", global.V2RAY_LOCATION_ASSET)
	if global.ServiceControlMode != global.DockerMode {
		wd, _ := os.Getwd()
		color.Red.Println("Service working directory is:", wd)
		conf := global.GetServiceConfig()
		color.Red.Println("Service listen: "+conf.Address+":"+conf.Port+",", "GUI demo: https://v2raya.mzz.pub")
	} else {
		fmt.Println("V2RayA is running in Docker. Compatibility mode starts up.")
	}
}
func checkUpdate() {
	setting := configure.GetSetting()
	if setting.GFWListAutoUpdateMode == configure.AutoUpdate {
		go func() {
			update, tRemote, err := service.IsUpdate()
			if err != nil {
				logs.Print("自动更新GFWList失败" + err.Error())
				return
			}
			if update {
				logs.Print("自动更新GFWList：目前最新版本为" + tRemote.Format("2006-01-02") + "，您的本地文件已最新，无需更新")
				return
			}
			/* 更新h2y.dat */
			localGFWListVersion, err := service.UpdateLocalGFWList()
			if err != nil {
				logs.Print("自动更新GFWList失败" + err.Error())
				return
			}
			logs.Print("自动更新GFWList完成，本地文件时间：" + localGFWListVersion)
		}()
	}
	if setting.CustomAutoUpdateMode == configure.AutoUpdate {
		//TODO
	}
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdate {
		subs := configure.GetSubscriptions()
		lenSubs := len(subs)
		go func() {
			control := make(chan struct{}, 2) //并发限制同时更新2个订阅
			wg := new(sync.WaitGroup)
			for i := 0; i < lenSubs; i++ {
				wg.Add(1)
				go func(i int) {
					control <- struct{}{}
					err := service.UpdateSubscription(i)
					if err != nil {
						logs.Print(fmt.Sprintf("自动更新订阅失败，id: %d，地址: %s，err: %v", i, subs[i].Address, err.Error()))
					} else {
						logs.Print(fmt.Sprintf("自动更新订阅成功，id: %d，地址: %s", i, subs[i].Address))
					}
					wg.Done()
					<-control
				}(i)
			}
			wg.Wait()
		}()
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	initConfigure()
	checkConnection()
	hello()
	checkUpdate()
	router.Run()
}
