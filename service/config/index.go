package config

import (
	"github.com/stevenroose/gonfig"
	"log"
)

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
