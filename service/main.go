package main

import (
	"V2RayA/extra/quickdown"
	"V2RayA/global"
	"V2RayA/model/ipforward"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/router"
	"V2RayA/service"
	"V2RayA/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
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
	_, port, err := net.SplitHostPort(global.GetEnvironmentConfig().Address)
	if err != nil {
		log.Fatal(err)
	}
	if occupied, which := tools.IsPortOccupied(port); occupied {
		log.Fatalf("V2RayA启动失败，%v端口已被%v占用", port, which)
	}
}

func initConfigure() {
	//初始化配置
	jsonIteratorExtra.RegisterFuzzyDecoders()
	if !configure.IsConfigureExists() {
		_ = os.MkdirAll(path.Dir(global.GetEnvironmentConfig().Config), os.ModeDir|0755)
		err := configure.SetConfigure(configure.New())
		if err != nil {
			log.Fatal(err)
		}
	}
	//配置文件描述符上限
	if global.ServiceControlMode == global.ServiceMode || global.ServiceControlMode == global.SystemctlMode {
		_ = v2ray.LiberalizeProcFile()
	}
	//配置ip转发
	setting := configure.GetSettingNotNil()
	if setting.Transparent != configure.TransparentClose {
		if setting.IpForward != ipforward.IsIpForwardOn() {
			_ = ipforward.WriteIpForward(setting.IpForward)
		}
	}
	//检查geoip、geosite是否存在
	if !v2ray.IsGeoipExists() {
		wg := new(sync.WaitGroup)
		wg.Add(2)
		dld := func(filename string) {
			color.Red.Println("正在安装" + filename)
			defer wg.Done()
			u := "https://cdn.jsdelivr.net/gh/v2ray/v2ray-core@master/release/config/" + filename
			p := v2ray.GetV2rayLocationAsset() + "/" + filename
			err := quickdown.DownloadWithWorkersTo(u, 5, p)
			if err != nil {
				return
			}
			err = os.Chmod(p, os.FileMode(0755))
		}
		go dld("geoip.dat")
		go dld("geosite.dat")
		wg.Wait()
	}
}

func checkConnection() {
	//如果V2Ray正在运行，而配置文件中没有记录当前连接的节点是谁，就关掉V2Ray
	if v2ray.IsV2RayRunning() && configure.GetConnectedServerNotNil() == nil {
		err := v2ray.StopAndDisableV2rayService()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func hello() {
	color.Red.Println("V2RayLocationAsset is", v2ray.GetV2rayLocationAsset())
	wd, _ := v2ray.GetV2rayWorkingDir()
	color.Red.Println("V2Ray binary is", wd+"/v2ray")
	if global.ServiceControlMode != global.DockerMode {
		wd, _ = os.Getwd()
		color.Red.Println("V2RayA working directory is", wd)
		color.Red.Println("Version:", global.Version)
	} else {
		fmt.Println("V2RayA is running in Docker. Compatible mode starts up.")
	}
	color.Red.Println("V2RayA is running at", global.GetEnvironmentConfig().Address)
}
func checkUpdate() {
	setting := service.GetSetting()

	//检查PAC文件更新
	if setting.PacAutoUpdateMode == configure.AutoUpdate {
		switch setting.PacMode {
		case configure.GfwlistMode:
			go func() {
				/* 更新h2y.dat */
				localGFWListVersion, err := service.CheckAndUpdateGFWList()
				if err != nil {
					log.Println("自动更新PAC文件失败" + err.Error())
					return
				}
				log.Println("自动更新PAC文件完成，本地文件时间：" + localGFWListVersion)
			}()
		case configure.CustomMode:
			//TODO
		}
	}

	//检查订阅更新
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdate {
		go func() {
			subs := configure.GetSubscriptions()
			lenSubs := len(subs)
			control := make(chan struct{}, 2) //并发限制同时更新2个订阅
			wg := new(sync.WaitGroup)
			for i := 0; i < lenSubs; i++ {
				wg.Add(1)
				go func(i int) {
					control <- struct{}{}
					err := service.UpdateSubscription(i, false)
					if err != nil {
						log.Println(fmt.Sprintf("自动更新订阅失败，id: %d，err: %v", i, err.Error()))
					} else {
						log.Println(fmt.Sprintf("自动更新订阅成功，id: %d，地址: %s", i, subs[i].Address))
					}
					wg.Done()
					<-control
				}(i)
			}
			wg.Wait()
		}()
	}
	// 检查服务端更新
	go func() {
		if foundNew, remote, err := service.CheckUpdate(); err == nil {
			global.FoundNew = foundNew
			global.RemoteVersion = remote
		}
	}()
}
func run() (err error) {
	//docker模式下把transparent纠正一下
	if global.ServiceControlMode == global.DockerMode {
		if err = configure.SetTransparent(configure.TransparentClose); err != nil {
			return
		}
	}
	err = service.CheckAndSetupTransparentProxy(true)
	if err != nil {
		return
	}
	errch := make(chan error)
	go func() {
		errch <- router.Run()
	}()
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGILL)
		<-sigs
		errch <- nil
	}()
	if err = <-errch; err != nil {
		return
	}
	fmt.Println("Quitting...")
	_ = service.CheckAndStopTransparentProxy()
	return nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	initConfigure()
	checkConnection()
	hello()
	checkUpdate()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
