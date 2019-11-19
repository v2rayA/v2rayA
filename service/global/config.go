package global

import (
	"github.com/stevenroose/gonfig"
	"log"
)

type Param struct {
	Address string `id:"v2raya_address" short:"a" default:"0.0.0.0" desc:"监听地址"`
	Port    string `id:"v2raya_port" short:"p" default:"2017" desc:"监听端口"`
	Config  string `id:"v2raya_config" short:"c" default:"/etc/v2ray/v2raya.json" desc:"V2RayA配置文件路径"`
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
