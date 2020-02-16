package main

import (
	"V2RayA/extra/download"
	"V2RayA/global"
	"V2RayA/model/ipforward"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/router"
	"V2RayA/service"
	"V2RayA/tools/ports"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

func checkEnvironment() {
	if runtime.GOOS == "windows" {
		fmt.Println("windows不支持直接运行，请配合docker使用。见https://github.com/mzz2017/V2RayA")
		fmt.Println("请按任意键继续...")
		_, _ = fmt.Scanf("\n")
		os.Exit(1)
	}
	if !global.GetEnvironmentConfig().PassCheckRoot {
		if os.Getegid() == -1 {
			log.Println("[warning] 无法判断当前是否sudo或拥有root权限，请确保V2RayA以sudo或root权限运行")
		} else if os.Getegid() != 0 {
			log.Fatal("请以sudo或root权限执行本程序. 或使用--passcheckroot参数跳过检查")
		}
	}
	_, port, err := net.SplitHostPort(global.GetEnvironmentConfig().Address)
	if err != nil {
		log.Fatal(err)
	}
	if occupied, which := ports.IsPortOccupied(port, "tcp", true); occupied {
		log.Fatalf("V2RayA启动失败，%v端口已被%v占用", port, which)
	}
}

func initConfigure() {
	//初始化配置
	jsonIteratorExtra.RegisterFuzzyDecoders()
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if configure.IsConfigureNotExists() {
		_ = os.MkdirAll(path.Dir(global.GetEnvironmentConfig().Config), os.ModeDir|0755)
		err := configure.SetConfigure(configure.New())
		if err != nil {
			log.Fatal(err)
		}
	}
	//配置文件描述符上限
	if global.ServiceControlMode == global.ServiceMode || global.ServiceControlMode == global.SystemctlMode {
		err := v2ray.LiberalizeProcFile()
		if err != nil && strings.Contains(err.Error(), "not be found") {
			log.Fatal(err.Error() + " 您可能未正确安装v2ray-core. 参考：https://github.com/mzz2017/V2RayA#%E4%BD%BF%E7%94%A8")
		}
	}
	//配置ip转发
	setting := configure.GetSettingNotNil()
	if setting.Transparent != configure.TransparentClose {
		if setting.IpForward != ipforward.IsIpForwardOn() {
			_ = ipforward.WriteIpForward(setting.IpForward)
		}
	}
	//检查geoip、geosite是否存在
	if !v2ray.IsGeoipExists() || !v2ray.IsGeositeExists() {
		dld := func(filename string) (err error) {
			color.Red.Println("正在安装" + filename)
			//jsdelivr经常版本落后，但这俩文件版本落后一点也没关系
			u := "https://cdn.jsdelivr.net/gh/v2ray/v2ray-core@master/release/config/" + filename
			p := v2ray.GetV2rayLocationAsset() + "/" + filename
			err = download.Pget(u, p)
			if err != nil {
				return errors.New("download(" + u + ")(" + p + "): " + err.Error())
			}
			err = os.Chmod(p, os.FileMode(0755))
			if err != nil {
				return errors.New("chmod: " + err.Error())
			}
			return
		}
		err := dld("geoip.dat")
		if err != nil {
			log.Println(err)
		}
		err = dld("geosite.dat")
		if err != nil {
			log.Println(err)
		}
	}
	//检查config.json是否存在
	if _, err := os.Stat(v2ray.GetConfigPath()); err != nil {
		//不存在就建一个。多数情况发生于docker模式挂载volume时覆盖了/etc/v2ray
		t := v2ray.NewTemplate()
		_ = v2ray.WriteV2rayConfig(t.ToConfigBytes())
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
	color.Red.Println("V2RayLocationAsset is", v2ray.GetV2rayLocationAsset())
	wd, _ := v2ray.GetV2rayWorkingDir()
	color.Red.Println("V2Ray binary is at", wd+"/v2ray")
	if global.ServiceControlMode != global.DockerMode {
		wd, _ = os.Getwd()
		color.Red.Println("V2RayA working directory is", wd)
		color.Red.Println("Version:", global.Version)
	} else {
		fmt.Println("V2RayA is running in Docker. Compatible mode starts up.")
		fmt.Printf("%v\n", "Waiting for container v2raya_v2ray's running. Refer: https://github.com/mzz2017/V2RayA#docker%E6%96%B9%E5%BC%8F")
		for !v2ray.IsV2RayProcessExists() {
			time.Sleep(1 * time.Second)
		}
		fmt.Println("Container v2raya_v2ray is ready.")
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
				time.Sleep(2 * time.Second)
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
			time.Sleep(2 * time.Second)
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
		//等待网络连通
		for {
			c := http.DefaultClient
			c.Timeout = 10 * time.Second
			resp, err := http.Get("http://www.gstatic.com/generate_204")
			if err == nil {
				_ = resp.Body.Close()
				break
			}
			time.Sleep(10 * time.Second)
		}
		if foundNew, remote, err := service.CheckUpdate(); err == nil {
			global.FoundNew = foundNew
			global.RemoteVersion = remote
		}
	}()
}
func run() (err error) {
	//判别是否common模式，需要启动v2ray吗
	if global.ServiceControlMode == global.CommonMode && configure.GetConnectedServer() != nil && !v2ray.IsV2RayProcessExists() {
		_ = v2ray.RestartV2rayService()
	}
	//刷新配置以刷新透明代理、ssr server
	v2ray.CheckAndStopTransparentProxy()
	err = v2ray.UpdateV2rayWithConnectedServer()
	if err != nil {
		w := configure.GetConnectedServer()
		log.Println("which:", w)
		_ = configure.ClearConnected()
	}
	errch := make(chan error)
	//启动服务端
	go func() {
		errch <- router.Run()
	}()
	//监听信号，处理透明代理的关闭
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
	v2ray.CheckAndStopTransparentProxy()
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
