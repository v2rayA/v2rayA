package global

import (
	"fmt"
	"github.com/stevenroose/gonfig"
	"log"
	"os"
	"sync"
)

type Params struct {
	Address       string `id:"address" short:"a" default:"0.0.0.0:2017" desc:"Listening address"`
	Config        string `id:"config" short:"c" default:"/etc/v2ray/v2raya.json" desc:"V2RayA configure path"`
	Mode          string `id:"mode" short:"m" desc:"Options: systemctl, service, universal. Auto-detect if not set"`
	SSRListenPort int    `short:"s" default:"12346" desc:"SSR outbound port"`
	PassCheckRoot bool   `desc:"Skip privilege checking. Use it only when you cannot start v2raya but confirm you have root privilege"`
	ResetPassword bool   `id:"reset-password"`
	ShowVersion   bool   `id:"version"`
}

var params Params

func initFunc() {
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
