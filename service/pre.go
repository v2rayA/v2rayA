package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/dat"
	service2 "github.com/v2rayA/v2rayA/core/v2ray/service"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/copyfile"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/pkg/util/privilege"
	"github.com/v2rayA/v2rayA/server/router"
	"github.com/v2rayA/v2rayA/server/service"

	confv4 "github.com/v2rayA/v2rayA-lib4/conf"

	touchv4 "github.com/v2rayA/v2rayA-lib4/core/touch"

	configurev4 "github.com/v2rayA/v2rayA-lib4/db/configure"

	servicev4 "github.com/v2rayA/v2rayA-lib4/server/service"
)

func checkEnvironment() {
	config := conf.GetEnvironmentConfig()
	if len(config.PrintReport) > 0 {
		db.SetReadOnly()
		config.Report()
		os.Exit(0)
	}
	if !config.PassCheckRoot {
		switch runtime.GOOS {
		case "linux":
			if !privilege.IsRootOrAdmin() && !config.Lite {
				log.Fatal("Please execute this program with sudo or as a root user for the best experience.\n" +
					"If you are sure you are root user, use the --passcheckroot parameter to skip the check.\n" +
					"If you don't want to run as root or you are a non-linux user, use --lite please.\n" +
					"For example:\n" +
					"$ v2raya --lite",
				)
			}
		case "windows":
			if !privilege.IsRootOrAdmin() && !config.Lite {
				log.Fatal("Please run v2rayA as Administrator (or SYSTEM) with elevation, or start with --lite to skip privilege checks.")
			}
		}
	}
	if config.ResetPassword {
		fmt.Println("Config directory:", config.Config)
		fmt.Println("Resetting password...\nIf no response for a long time, please stop other v2rayA instances and try again.")
		err := configure.ResetAccounts()
		if err != nil {
			log.Fatal("checkEnvironment: %v", err)
		}
		fmt.Println("Succeed. It will work after you restart v2rayA.")
		os.Exit(0)
	}
	_, v2rayAListeningPort, err := net.SplitHostPort(config.Address)
	if err != nil {
		log.Fatal("checkEnvironment: %v", err)
	}
	if occupied, sockets, err := ports.IsPortOccupied([]string{v2rayAListeningPort + ":tcp"}); occupied {
		if err != nil {
			log.Fatal("netstat: %v", err)
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
	// check if tproxy can be enabled
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
	// initialize configuration
	jsonIteratorExtra.RegisterFuzzyDecoders()

	//db
	dbPath := filepath.Join(conf.GetEnvironmentConfig().Config, "bolt.db")
	if _, e := os.Lstat(dbPath); os.IsNotExist(e) {
		//confv4.SetConfig(confv4.Params{Config: conf.GetEnvironmentConfig().Config})
		// need to migrate?
		if !configurev4.IsConfigureNotExists() {
			// There is different format in server and subscription.
			// So we keep other content and reimport servers and subscriptions.
			log.Warn("Migrating from v4 to main")
			if err := copyfile.CopyFileContent(filepath.Join(
				confv4.GetEnvironmentConfig().Config,
				"boltv4.db",
			), filepath.Join(
				conf.GetEnvironmentConfig().Config,
				"bolt.db",
			)); err != nil {
				log.Fatal("Failed to copy boltv4.db to bolt.db: %v", err)
			}

			// clear connects of outbounds
			for _, out := range configure.GetOutbounds() {
				_ = configure.ClearConnects(out)
			}
			var indexes []int
			for i := 0; i < configurev4.GetLenServers(); i++ {
				indexes = append(indexes, i)
			}
			_ = configure.RemoveServers(indexes)

			indexes = nil
			for i := 0; i < configurev4.GetLenSubscriptions(); i++ {
				indexes = append(indexes, i)
			}
			_ = configure.RemoveSubscriptions(indexes)

			// migrate servers and subscriptions
			t := touchv4.GenerateTouch()
			subs := configurev4.GetSubscriptionsV2()
			for _, sub := range subs {
				log.Info("Importing subscription: %v", sub.Address)
				if e := service.Import(sub.Address, nil); e != nil {
					log.Warn("Failed to migrate subscription: %v", sub.Address)
				}
			}
			for iSvr := range t.Servers {
				if addr, e := servicev4.GetSharingAddress(&configurev4.Which{
					TYPE: configurev4.ServerType,
					ID:   iSvr + 1,
				}); e == nil {
					if e := service.Import(addr, nil); e != nil {
						log.Warn("Failed to migrate server: %v", addr)
					}
				}
			}

			log.Warn("Migration is done")
		} else {
			initDBValue()
		}
	}

	if len(configure.GetTproxyWhiteIpGroups().CountryCodes) == 0 {
		configure.SetTproxyWhiteIpGroups([]string{"PRIVATE"}, []string{})
	}

	// check if config.json exists
	if _, err := os.Stat(asset.GetV2rayConfigPath()); err != nil {
		// if not exists, create one. This mostly happens when mounting a volume in docker mode and it covers /etc/v2ray.
		t := v2ray.Template{}
		_ = v2ray.WriteV2rayConfig(t.ToConfigBytes())
	}

	// first determine if v2ray exists
	if _, err := where.GetV2rayBinPath(); err == nil {
		// check if geoip, geosite exist
		if !asset.DoesV2rayAssetExist("geoip.dat") || !asset.DoesV2rayAssetExist("geosite.dat") {
			log.Alert("downloading missing geoip.dat and geosite.dat")
			var l net.Listener
			if l, err = net.Listen("tcp", conf.GetEnvironmentConfig().Address); err != nil {
				log.Fatal("net.Listen: %v", err)
			}
			e := gin.New()
			e.GET("/", func(c *gin.Context) {
				c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.String(200, "Downloading missing geoip.dat and geosite.dat; refresh the page later.")
			})
			go e.RunListener(l)
			if !asset.DoesV2rayAssetExist("geoip.dat") {
				err := dat.UpdateLocalGeoIP()
				if err != nil {
					log.Fatal("UpdateLocalGeoIP: %v", err)
				}
			}
			if !asset.DoesV2rayAssetExist("geosite.dat") {
				err = dat.UpdateLocalGeoSite()
				if err != nil {
					log.Fatal("UpdateLocalGeoSite: %v", err)
				}
			}
			if l != nil {
				l.Close()
			}
			log.Alert("geoip.dat and geosite.dat are ready")
		}
	}
}

func hello() {
	v2rayPath, _ := where.GetV2rayBinPath()
	log.Alert("V2Ray binary is %v", v2rayPath)
	log.Alert("V2Ray asset directory is %v", asset.GetV2rayLocationAssetOverride())
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
	control := make(chan struct{}, 2) // concurrency limit: update 2 subscriptions at a time
	// Disconnect from subscriptions before auto-selecting servers from them
	// to limit the number of connected servers and avoid hitting the limit
	shouldDisconnect := true
	err := service.AutoSelectServersFromSubscriptions(shouldDisconnect)
	if err != nil {
		log.Error("[AutoSelect] Failed to disconnect servers from subscriptions -- err: %v", err)
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < lenSubs; i++ {
		wg.Add(1)
		go func(i int) {
			control <- struct{}{}
			err := service.UpdateSubscription(i, false)
			if err != nil {
				log.Info("[AutoUpdate] Subscriptions: Failed to update subscription -- ID: %d, err: %v", i, err)
			} else {
				log.Info("[AutoUpdate] Subscriptions: Complete updating subscription -- ID: %d, Address: %s", i, subs[i].Address)
			}
			wg.Done()
			<-control
		}(i)
	}
	wg.Wait()
	shouldDisconnect = false
	err2 := service.AutoSelectServersFromSubscriptions(shouldDisconnect)
	if err2 != nil {
		log.Error("[AutoSelect] Failed to auto-select servers from subscriptions -- err: %v", err2)
	}

}

func initUpdatingTicker() {
	conf.TickerUpdateGFWList = time.NewTicker(24 * time.Hour * 365 * 100)
	conf.TickerUpdateSubscription = time.NewTicker(24 * time.Hour * 365 * 100)
	go func() {
		for range conf.TickerUpdateGFWList.C {
			_, err := dat.CheckAndUpdateGFWList("")
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

	// initialize ticker
	initUpdatingTicker()

	// check for PAC file updates
	if setting.GFWListAutoUpdateMode == configure.AutoUpdate ||
		setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals ||
		setting.Transparent == configure.TransparentGfwlist {
		if setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals {
			conf.TickerUpdateGFWList.Reset(time.Duration(setting.GFWListAutoUpdateIntervalHour) * time.Hour)
		}
		switch setting.RulePortMode {
		case configure.GfwlistMode:
			go func() {
				/* Update LoyalsoldierSite.dat */
				localGFWListVersion, err := dat.CheckAndUpdateGFWList("")
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

	// check for subscription updates
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdate ||
		setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {

		if setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {
			conf.TickerUpdateSubscription.Reset(time.Duration(setting.SubscriptionAutoUpdateIntervalHour) * time.Hour)
		}
		go updateSubscriptions()
	}
	// check for server updates
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
	// check if v2ray should be started
	if configure.GetRunning() {
		//configure the ip forward
		setting := service.GetSetting()
		if setting.IpForward != ipforward.IsIpForwardOn() {
			e := ipforward.WriteIpForward(setting.IpForward)
			if e != nil {
				log.Warn("Connect: %v", e)
			}
		}
		// Start.
		err := v2ray.UpdateV2RayConfig()
		if err != nil {
			log.Error("failed to start v2ray-core: %v", err)
		}
	} else {
		log.Info("the core was not running the last time v2rayA exited")
	}
	//w := configure.GetConnectedServers()
	//log.Println(err, ", which:", w)
	//_ = configure.ClearConnected()
	errch := make(chan error)
	// start server
	go func() {
		errch <- router.Run()
	}()
	// listen for signals to handle transparent proxy shutdown
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
	v2ray.ProcessManager.CheckAndStopTransparentProxy(nil)
	v2ray.ProcessManager.Stop(false)
	_ = db.DB().Close()
	return nil
}
