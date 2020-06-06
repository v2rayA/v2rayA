package main

import (
	"fmt"
	"github.com/gookit/color"
	jsoniter "github.com/json-iterator/go"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"v2ray.com/core/common/errors"
	"github.com/mzz2017/v2rayA/common/netTools/ports"
	"github.com/mzz2017/v2rayA/core/ipforward"
	"github.com/mzz2017/v2rayA/core/iptables"
	"github.com/mzz2017/v2rayA/core/v2ray"
	"github.com/mzz2017/v2rayA/core/v2ray/asset"
	"github.com/mzz2017/v2rayA/core/v2ray/asset/gfwlist"
	"github.com/mzz2017/v2rayA/db"
	"github.com/mzz2017/v2rayA/db/configure"
	"github.com/mzz2017/v2rayA/extra/gopeed"
	"github.com/mzz2017/v2rayA/global"
	"github.com/mzz2017/v2rayA/router"
	"github.com/mzz2017/v2rayA/service"
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
		fmt.Println("ok")
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
		log.Println("[INFO] Cannot load TPROXY module:", err, ". Switch to DNSPoison module")
	}
	v2ray.CheckAndStopTransparentProxy()
	preprocess := func(c *iptables.SetupCommands) {
		commands := string(*c)
		lines := strings.Split(commands, "\n")
		reg := regexp.MustCompile(`{{.+}}`)
		for i, line := range lines {
			if len(reg.FindString(line)) > 0 {
				lines[i] = ""
			}
		}
		commands = strings.Join(lines, "\n")
		*c = iptables.SetupCommands(commands)
	}
	err := iptables.Tproxy.GetSetupCommands().Setup(&preprocess)
	if err != nil {
		log.Println(err)
		global.SupportTproxy = false
	}
	iptables.Tproxy.GetCleanCommands().Clean()
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
	b, err := ioutil.ReadFile(jsonConfPath)
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
		_ = os.MkdirAll(path.Dir(confPath), os.ModeDir|0644)
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

	//配置文件描述符上限
	if global.ServiceControlMode == global.ServiceMode || global.ServiceControlMode == global.SystemctlMode {
		err := v2ray.OptimizeServiceFile()
		if err != nil {
			log.Println(err)
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
	if !asset.IsGeoipExists() || !asset.IsGeositeExists() {
		dld := func(repo, filename, localname string) (err error) {
			color.Red.Println("installing" + filename)
			p := asset.GetV2rayLocationAsset() + "/" + filename
			resp, err := http.Get("https://api.github.com/repos/" + repo + "/tags")
			if err != nil {
				return
			}
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
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
			return
		}
		err := dld("mzz2017/dist-geoip", "geoip.dat", "geoip.dat")
		if err != nil {
			log.Println(err)
		}
		err = dld("mzz2017/dist-domain-list-community", "dlc.dat", "geosite.dat")
		if err != nil {
			log.Println(err)
		}
	}
	//检查config.json是否存在
	if _, err := os.Stat(asset.GetConfigPath()); err != nil {
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
	color.Red.Println("V2RayLocationAsset is", asset.GetV2rayLocationAsset())
	wd, _ := asset.GetV2rayWorkingDir()
	color.Red.Println("V2Ray binary is at", wd+"/v2ray")
	wd, _ = os.Getwd()
	color.Red.Println("v2rayA working directory is", wd)
	color.Red.Println("Version:", global.Version)
}

func checkUpdate() {
	//等待网络连通
	for {
		c := http.DefaultClient
		c.Timeout = 5 * time.Second
		resp, err := http.Get("http://www.gstatic.com/generate_204")
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(c.Timeout)
	}

	setting := service.GetSetting()
	//检查PAC文件更新
	if setting.PacAutoUpdateMode == configure.AutoUpdate || setting.Transparent == configure.TransparentGfwlist {
		switch setting.PacMode {
		case configure.GfwlistMode:
			go func() {
				/* 更新LoyalsoldierSite.dat */
				localGFWListVersion, err := gfwlist.CheckAndUpdateGFWList()
				if err != nil {
					log.Println("Fail in updating PAC file: " + err.Error())
					return
				}
				log.Println("Complete updating PAC file. Localtime: " + localGFWListVersion)
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
						log.Println(fmt.Sprintf("Fail in updating subscription -- ID: %d，err: %v", i, err.Error()))
					} else {
						log.Println(fmt.Sprintf("Complete updating subscription -- ID: %d，地址: %s", i, subs[i].Address))
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
	if configure.GetConnectedServer() != nil {
		_ = v2ray.RestartV2rayService()
	}
	//刷新配置以刷新透明代理、ssr server
	err = v2ray.UpdateV2RayConfig(nil)
	if err != nil {
		w := configure.GetConnectedServer()
		log.Println(err, ", which:", w)
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
		log.Fatal(err)
	}
	fmt.Println("Quitting...")
	v2ray.CheckAndStopTransparentProxy()
	_ = v2ray.StopV2rayService()
	_ = db.DB().Close()
	return nil
}
