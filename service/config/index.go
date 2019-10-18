package config

import (
	"github.com/stevenroose/gonfig"
	"log"
)

type Config struct {
	Address     string `id:"address" short:"a" default:"0.0.0.0" desc:"listening address"`
	Port        string `id:"port" short:"p" default:"8080" desc:"listening port"`
	//ApiRoot     string `id:"api_root" short:"a" default:"http://129.204.71.113:9999/api/v2" desc:"the backend api root"`
	//Username    string `id:"username" short:"u" desc:"your username"`
	//Password    string `id:"password" short:"p" desc:"your password"`
	//LoginType   string `id:"login_type" short:"t" default:"school" desc:"login type=enum{school, phone, normal}"`
	//Socks5Proxy string `id:"socks5_proxy" desc:"socks5 proxy socket"`
	//Timeout     int    `id:"timeout" default:"30000" desc:"the timeout limit of waiting http response headers. unit is ms"`
	//RetryLimit  int    `id:"retry_limit" default:"5" desc:"the number of retry when crawl fail"`
}

var config Config

func init() {
	err := gonfig.Load(&config, gonfig.Conf{
		FileDisable: true,
		//FlagIgnoreUnknown: true,
	})
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log.Fatal(err)
		}
	}
}

func Get() *Config {
	return &config
}
