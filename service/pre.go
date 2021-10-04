package main

import (
	"errors"
	"fmt"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/gfwlist"
	service2 "github.com/v2rayA/v2rayA/core/v2ray/service"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/copyfile"
	"github.com/v2rayA/v2rayA/pkg/util/gopeed"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/router"
	"github.com/v2rayA/v2rayA/server/service"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func checkEnvironment() {
	config := conf.GetEnvironmentConfig()
	if len(config.PrintReport) > 0 {
		config.Report()
		os.Exit(0)
	}
	if !config.PassCheckRoot || config.ResetPassword {
		if os.Getegid() != 0 {
			log.Fatal("Please execute this program with sudo or as a root user for the best experience.\n" +
				"If you are sure you are root user, use the --passcheckroot parameter to skip the check.\n" +
				"If you don't want to run as root or you are a non-linux user, use --lite please.\n" +
				"For example:\n" +
				"$ v2raya --lite",
			)
		}
	}
	if config.ResetPassword {
		err := configure.ResetAccounts()
		if err != nil {
			log.Fatal("checkEnvironment: %v", err)
		}
		fmt.Println("It will work after you restart v2rayA")
		os.Exit(0)
	}
	_, v2rayAListeningPort, err := net.SplitHostPort(config.Address)
	if err != nil {
		log.Fatal("checkEnvironment: %v", err)
	}
	if occupied, sockets, err := ports.IsPortOccupied([]string{v2rayAListeningPort + ":tcp"}); occupied {
		if err != nil {
			log.Fatal("netstat:", err)
		}
		for _, socket := range sockets {
			process, err := socket.Process()
			if err == nil {
				log.Fatal("Port %v is occupied by %v/%v", v2rayAListeningPort, process.Name, process.PID)
			}
		}
	}
}

func checkTProxySupportability() {
	if conf.GetEnvironmentConfig().Lite {
		return
	}
	//检查tproxy是否可以启用
	if err := service2.CheckAndProbeTProxy(); err != nil {
		log.Info("Cannot load TPROXY module: %v", err)
	}
}

func initDBValue() {
	log.Info("init DB")
	err := configure.SetConfigure(configure.New())
	if err != nil {
		log.Fatal("initDBValue: %v", err)
	}
}

func initConfigure() {
	//等待网络连通
	v2ray.CheckAndStopTransparentProxy()
	for {
		addrs, err := resolv.LookupHost("apple.com")
		if err == nil && len(addrs) > 0 {
			break
		}
		log.Alert("waiting for network connected")
		time.Sleep(5 * time.Second)
	}
	log.Alert("network is connected")
	//初始化配置
	jsonIteratorExtra.RegisterFuzzyDecoders()

	//db
	confPath := conf.GetEnvironmentConfig().Config
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(confPath), os.ModeDir|0750)
	}
	if configure.IsConfigureNotExists() {
		initDBValue()
	}
	//检查config.json是否存在
	if _, err := os.Stat(asset.GetV2rayConfigPath()); err != nil {
		//不存在就建一个。多数情况发生于docker模式挂载volume时覆盖了/etc/v2ray
		t := v2ray.Template{}
		_ = v2ray.WriteV2rayConfig(t.ToConfigBytes())
	}

	//首先确定v2ray是否存在
	if _, err := where.GetV2rayBinPath(); err == nil {
		//检查geoip、geosite是否存在
		if !asset.IsGeoipExists() || !asset.IsGeositeExists() {
			assetDir := asset.GetV2rayLocationAsset()
			dld := func(repo, remoteFilename, localFilename string) (err error) {
				log.Warn("downloading %v to %v", remoteFilename, assetDir)
				p := filepath.Join(os.TempDir(), remoteFilename)
				resp, err := http.Get("https://api.github.com/repos/" + repo + "/tags")
				if err != nil {
					return
				}
				defer resp.Body.Close()
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return
				}
				tag := gjson.GetBytes(b, "0.name").String()
				u := fmt.Sprintf("https://cdn.jsdelivr.net/gh/%v@%v/%v", repo, tag, remoteFilename)
				err = gopeed.Down(&gopeed.Request{
					Method: "GET",
					URL:    u,
				}, p)
				if err != nil {
					return errors.New("download<" + p + ">: " + err.Error())
				}
				err = os.Chmod(p, os.FileMode(0755))
				if err != nil {
					return errors.New("chmod: " + err.Error())
				}
				return copyfile.CopyFile(p, filepath.Join(assetDir, localFilename))
			}
			err := dld("v2rayA/dist-geoip", "geoip.dat", "geoip.dat")
			if err != nil {
				log.Warn("initConfigure: v2rayA/dist-geoip: %v", err)
			}
			err = dld("v2rayA/dist-domain-list-community", "dlc.dat", "geosite.dat")
			if err != nil {
				log.Warn("initConfigure: v2rayA/dist-domain-list-community: %v", err)
			}
		}
	}
}

func hello() {
	log.Alert("V2RayLocationAsset is %v", asset.GetV2rayLocationAsset())
	v2rayPath, _ := where.GetV2rayBinPath()
	log.Alert("V2Ray binary is %v", v2rayPath)
	wd, _ := os.Getwd()
	log.Alert("v2rayA working directory is %v", wd)
	log.Alert("v2rayA configuration directory is %v", conf.GetEnvironmentConfig().Config)
	log.Alert("Golang: %v", runtime.Version())
	log.Alert("OS: %v", runtime.GOOS)
	log.Alert("Arch: %v", runtime.GOARCH)
	log.Alert("Lite: %v", conf.GetEnvironmentConfig().Lite)
	log.Alert("Version: %v", conf.Version)
	log.Alert("Starting...")
}

func updateSubscriptions() {
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
				log.Info("[AutoUpdate] Subscriptions: Failed to update subscription -- ID: %d，err: %v", i, err)
			} else {
				log.Info("[AutoUpdate] Subscriptions: Complete updating subscription -- ID: %d，Address: %s", i, subs[i].Address)
			}
			wg.Done()
			<-control
		}(i)
	}
	wg.Wait()
}

func initUpdatingTicker() {
	conf.TickerUpdateGFWList = time.NewTicker(24 * time.Hour * 365 * 100)
	conf.TickerUpdateSubscription = time.NewTicker(24 * time.Hour * 365 * 100)
	go func() {
		for range conf.TickerUpdateGFWList.C {
			_, err := gfwlist.CheckAndUpdateGFWList()
			if err != nil {
				log.Info("[AutoUpdate] GFWList: %v", err)
			}
		}
	}()
	go func() {
		for range conf.TickerUpdateSubscription.C {
			updateSubscriptions()
		}
	}()
}

func checkUpdate() {
	setting := service.GetSetting()

	//初始化ticker
	initUpdatingTicker()

	//检查PAC文件更新
	if setting.GFWListAutoUpdateMode == configure.AutoUpdate ||
		setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals ||
		setting.Transparent == configure.TransparentGfwlist {
		if setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals {
			conf.TickerUpdateGFWList.Reset(time.Duration(setting.GFWListAutoUpdateIntervalHour) * time.Hour)
		}
		switch setting.RulePortMode {
		case configure.GfwlistMode:
			go func() {
				/* 更新LoyalsoldierSite.dat */
				localGFWListVersion, err := gfwlist.CheckAndUpdateGFWList()
				if err != nil {
					log.Warn("Failed to update PAC file: %v", err.Error())
					return
				}
				log.Info("Complete updating PAC file. Localtime: %v", localGFWListVersion)
			}()
		case configure.CustomMode:
			// obsolete
		}
	}

	//检查订阅更新
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdate ||
		setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {

		if setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {
			conf.TickerUpdateSubscription.Reset(time.Duration(setting.SubscriptionAutoUpdateIntervalHour) * time.Hour)
		}
		go updateSubscriptions()
	}
	// 检查服务端更新
	go func() {
		f := func() {
			if foundNew, remote, err := service.CheckUpdate(); err == nil {
				conf.FoundNew = foundNew
				conf.RemoteVersion = remote
			}
		}
		f()
		c := time.Tick(7 * 24 * time.Hour)
		for range c {
			f()
		}
	}()
}

func run() (err error) {
	//判别需要启动v2ray吗
	if configure.GetRunning() {
		err := v2ray.UpdateV2RayConfig()
		if err != nil {
			log.Error("failed to start v2ray-core: %v", err)
		}
	}
	//w := configure.GetConnectedServers()
	//log.Println(err, ", which:", w)
	//_ = configure.ClearConnected()
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
		log.Fatal("run: %v", err)
	}
	fmt.Println("Quitting...")
	v2ray.CheckAndStopTransparentProxy()
	v2ray.ProcessManager.Stop(false)
	_ = db.DB().Close()
	return nil
}
