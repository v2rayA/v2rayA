package global

import (
	"fmt"
	"github.com/stevenroose/gonfig"
	"log"
	"os"
	"sync"
)

type Params struct {
	Address          string `id:"address" short:"a" default:"0.0.0.0:2017" desc:"Listening address"`
	Config           string `id:"config" short:"c" default:"/etc/v2ray/v2raya.json" desc:"v2rayA configure path"`
	Mode             string `id:"mode" short:"m" desc:"Options: systemctl, service, universal. Auto-detect if not set"`
	PluginListenPort int    `short:"s" default:"32346" desc:"Plugin outbound port"`
	PassCheckRoot    bool   `desc:"Skip privilege checking. Use it only when you cannot start v2raya but confirm you have root privilege"`
	ResetPassword    bool   `id:"reset-password"`
	ShowVersion      bool   `id:"version"`
}

var params Params

var dontLoadConfig bool

func initFunc() {
	defer SetServiceControlMode(params.Mode)
	if dontLoadConfig {
		return
	}
	err := gonfig.Load(&params, gonfig.Conf{
		FileDisable:       true,
		FlagIgnoreUnknown: false,
		EnvPrefix:         "V2RAYA_",
	})
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log.Fatal(err)
		}
	}
	if params.ShowVersion {
		fmt.Println(Version)
		os.Exit(0)
	}
}

var once sync.Once

func GetEnvironmentConfig() *Params {
	once.Do(initFunc)
	return &params
}

func SetConfig(config Params) {
	params = config
}

func DontLoadConfig() {
	dontLoadConfig = true
}
