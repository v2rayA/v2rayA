package global

import (
	"github.com/stevenroose/gonfig"
	"log"
	"sync"
)

type Param struct {
	Address string `id:"address" short:"a" default:"0.0.0.0" desc:"监听地址"`
	Port    string `id:"port" short:"p" default:"2017" desc:"监听端口"`
	Config  string `id:"config" short:"c" default:"/etc/v2ray/v2raya.json" desc:"V2RayA配置文件路径"`
}

var param Param

func initFunc() {
	err := gonfig.Load(&param, gonfig.Conf{
		FileDisable:       true,
		FlagIgnoreUnknown: false,
		EnvPrefix:         "V2RAYA_",
	})
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log.Fatal(err)
		}
	}
}

func GetServiceConfig() *Param {
	var once sync.Once
	once.Do(initFunc)
	return &param
}
