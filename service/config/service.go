package config

import (
	"github.com/stevenroose/gonfig"
	"log"
)

type Param struct {
	Address string `id:"address" short:"a" default:"0.0.0.0" desc:"listening address"`
	Port    string `id:"port" short:"p" default:"2017" desc:"listening port"`
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