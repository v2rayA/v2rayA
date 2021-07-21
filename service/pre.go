package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gookit/color"
	jsoniter "github.com/json-iterator/go"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/gfwlist"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/extra/gopeed"
	"github.com/v2rayA/v2rayA/global"
	"github.com/v2rayA/v2rayA/server/router"
	"github.com/v2rayA/v2rayA/server/service"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func checkEnvironment() {
	if runtime.GOOS == "windows" {
		fmt.Println("v2rayA cannot run on windows")
		fmt.Println("Press any key to continue...")
		_, _ = fmt.Scanf("\n")
		os.Exit(1)
	}
	conf := global.GetEnvironmentConfig()
	if !conf.PassCheckRoot || conf.ResetPassword {
		if os.Getegid() != 0 {
			log.Fatal("Please execute this program with sudo or as a root user. If you are sure that you have root privileges, you can use the --passcheckroot parameter to skip the check")
		}
	}
	if conf.ResetPassword {
		err := configure.ResetAccounts()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("It will work after you restart v2rayA")
		os.Exit(0)
	}
	_, port, err := net.SplitHostPort(conf.Address)
	if err != nil {
		log.Fatal(err)
	}
	if occupied, socket, err := ports.IsPortOccupied([]string{port + ":tcp"}); occupied {
		if err != nil {
			log.Fatal("netstat:", err)
		}
		process, err := socket.Process()
		if err == nil {
			log.Fatalf("Port %v is occupied by %v/%v", port, process.Name, process.PID)
		}
	}
}

func checkTProxySupportability() {
	//检查tproxy是否可以启用
	if err := v2ray.CheckAndProbeTProxy(); err != nil {
		log.Println("[INFO] Cannot load TPROXY module:", err)
	}
}

func migrate(jsonConfPath string) (err error) {
	log.Println("[info] Migrating json to nutsdb...")
	defer func() {
		if err != nil {
			log.Println("[info] Migrating failed: ", err.Error())
		} else {
			log.Println("[info] Migrating complete")
		}
	}()
	b, err := os.ReadFile(jsonConfPath)
	if err != nil {
		return
	}
	var cfg configure.Configure
	if err = jsoniter.Unmarshal(b, &cfg); err != nil {
		return
	}
	if err = configure.SetConfigure(&cfg); err != nil {
		return
	}
	return nil
}

func initDBValue() {
	err := configure.SetConfigure(configure.New())
	if err != nil {
		log.Fatal(err)
	}
}

func initConfigure() {
	//初始化配置
	jsonIteratorExtra.RegisterFuzzyDecoders()
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//db
	confPath := global.GetEnvironmentConfig().Config
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		_ = os.MkdirAll(path.Dir(confPath), os.ModeDir|0750)
	}
	if configure.IsConfigureNotExists() {
		// need to migrate?
		camp := []string{path.Join(path.Dir(confPath), "v2raya.json"), "/etc/v2ray/v2raya.json", "/etc/v2raya/v2raya.json"}
		var ok bool
		for _, jsonConfPath := range camp {
			if _, err := os.Stat(jsonConfPath); err == nil {
				err = migrate(jsonConfPath)
				if err == nil {
					ok = true
					break
				}
			}
		}
		if !ok {
			initDBValue()
		}
	}
	//检查config.json是否存在
	if _, err := os.Stat(asset.GetV2rayConfigPath()); err != nil {
		//不存在就建一个。多数情况发生于docker模式挂载volume时覆盖了/etc/v2ray
		t := v2ray.NewTemplate()
		_ = v2ray.WriteV2rayConfig(t.ToConfigBytes())
	}

	//首先确定v2ray是否存在
	if _, err := where.GetV2rayBinPath(); err == nil {
		//检查geoip、geosite是否存在
		if !asset.IsGeoipExists() || !asset.IsGeositeExists() {
			dld := func(repo, filename, localname string) (err error) {
				color.Red.Println("installing " + filename)
				p := asset.GetV2rayLocationAsset() + "/" + filename
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
				u := fmt.Sprintf("https://cdn.jsdelivr.net/gh/%v@%v/%v", repo, tag, filename)
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
				os.Rename(p, asset.GetV2rayLocationAsset()+"/"+localname)
				return
			}
			err := dld("v2rayA/dist-geoip", "geoip.dat", "geoip.dat")
			if err != nil {
				log.Println(err)
			}
			err = dld("v2rayA/dist-domain-list-community", "dlc.dat", "geosite.dat")
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func hello() {
	color.Red.Println("V2RayLocationAsset is", asset.GetV2rayLocationAsset())
	v2rayPath, _ := where.GetV2rayBinPath()
	color.Red.Println("V2Ray binary is", v2rayPath)
	wd, _ := os.Getwd()
	color.Red.Println("v2rayA working directory is", wd)
	color.Red.Println("Version:", global.Version)
	color.Red.Println("Starting...")
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
				log.Println(fmt.Sprintf("[AutoUpdate] Subscriptions: Failed to update subscription -- ID: %d，err: %v", i, err.Error()))
			} else {
				log.Println(fmt.Sprintf("[AutoUpdate] Subscriptions: Complete updating subscription -- ID: %d，Address: %s", i, subs[i].Address))
			}
			wg.Done()
			<-control
		}(i)
	}
	wg.Wait()
}

func initUpdatingTicker() {
	global.TickerUpdateGFWList = time.NewTicker(24 * time.Hour * 365 * 100)
	global.TickerUpdateSubscription = time.NewTicker(24 * time.Hour * 365 * 100)
	go func() {
		for range global.TickerUpdateGFWList.C {
			_, err := gfwlist.CheckAndUpdateGFWList()
			if err != nil {
				log.Println("[AutoUpdate] GFWList:", err)
			}
		}
	}()
	go func() {
		for range global.TickerUpdateSubscription.C {
			updateSubscriptions()
		}
	}()
}

func checkUpdate() {
	setting := service.GetSetting()
	//等待网络连通
	resolver := net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			address = "114.114.114.114:53"
			return d.DialContext(ctx, network, address)
		},
	}
	for {
		c := http.DefaultClient
		c.Timeout = 5 * time.Second
		addrs, err := resolver.LookupHost(context.Background(), "apple.com")
		if err == nil && len(addrs) > 0 {
			break
		}
		log.Println("[info] waiting for network connected")
		time.Sleep(c.Timeout)
	}
	log.Println("[info] network is connected")

	//初始化ticker
	initUpdatingTicker()

	//检查PAC文件更新
	if setting.GFWListAutoUpdateMode == configure.AutoUpdate ||
		setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals ||
		setting.Transparent == configure.TransparentGfwlist {
		if setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals {
			global.TickerUpdateGFWList.Reset(time.Duration(setting.GFWListAutoUpdateIntervalHour) * time.Hour)
		}
		switch setting.RulePortMode {
		case configure.GfwlistMode:
			go func() {
				/* 更新LoyalsoldierSite.dat */
				localGFWListVersion, err := gfwlist.CheckAndUpdateGFWList()
				if err != nil {
					log.Println("Failed to update PAC file: " + err.Error())
					return
				}
				log.Println("Complete updating PAC file. Localtime: " + localGFWListVersion)
			}()
		case configure.CustomMode:
			// obsolete
		}
	}

	//检查订阅更新
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdate ||
		setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {

		if setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {
			global.TickerUpdateSubscription.Reset(time.Duration(setting.SubscriptionAutoUpdateIntervalHour) * time.Hour)
		}
		go updateSubscriptions()
	}
	// 检查服务端更新
	go func() {
		f := func() {
			if foundNew, remote, err := service.CheckUpdate(); err == nil {
				global.FoundNew = foundNew
				global.RemoteVersion = remote
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
	if w := configure.GetConnectedServer(); w != nil {
		_ = service.Connect(w)
	}
	//w := configure.GetConnectedServer()
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
		log.Fatal(err)
	}
	fmt.Println("Quitting...")
	v2ray.CheckAndStopTransparentProxy()
	_ = v2ray.StopV2rayService()
	_ = db.DB().Close()
	return nil
}
