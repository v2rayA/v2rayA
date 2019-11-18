package global

import (
	"github.com/stevenroose/gonfig"
	"log"
)

type Param struct {
	Address string `id:"address" short:"a" default:"0.0.0.0" desc:"监听地址"`
	Port    string `id:"port" short:"p" default:"2017" desc:"监听端口"`
	RedisServer string `id:"redis" default:":6379" desc:"DEPRESSED!! redis server socket"`
	Config  string `id:"config" default:"config.json" desc:"V2RayA配置文件路径"`
}

var param Param

func init() {
	err := gonfig.Load(&param, gonfig.Conf{
		FileDisable: true,
		//FlagIgnoreUnknown: true,
	})
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log.Fatal(err)
		}
	}
}

func GetServiceConfig() *Param {
	return &param
}
