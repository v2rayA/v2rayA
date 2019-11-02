package global

import (
	"github.com/stevenroose/gonfig"
	"log"
)

type Param struct {
	Address    string `id:"address" short:"a" default:"0.0.0.0" desc:"监听地址"`
	Port       string `id:"port" short:"p" default:"2017" desc:"监听端口"`
	ConfigPath string `id:"conf" short:"c" default:"./.tr" desc:"配置文件所在路径，默认为当前目录下的.tr文件"`
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
