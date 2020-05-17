package config

import (
	"github.com/stevenroose/gonfig"
	"log"
	"sync"
)

type Params struct {
	File     string `id:"file" short:"f" desc:"input file where share-links are split by lines"`
	Link     string `id:"link" short:"l" desc:"subscription link or share-link"`
	Timeout  int    `id:"timeout" short:"t" default:"10000" desc:"test timeout(ms)"`
	Parallel int    `id:"parallel" short:"p" default:"5" desc:"the max number of parallel tests"`
}

var params Params

func initFunc() {
	err := gonfig.Load(&params, gonfig.Conf{
		FileDisable:       true,
		EnvDisable:        true,
		FlagIgnoreUnknown: true,
	})
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log.Fatal(err)
		}
	}
}

var once sync.Once

func GetConfig() *Params {
	once.Do(initFunc)
	return &params
}