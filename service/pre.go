package main

import (
	"os"
	"runtime"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func hello() {
	v2rayPath, _ := where.GetV2rayBinPath()
	log.Alert("V2Ray binary is %v", v2rayPath)
	log.Alert("V2Ray asset directory is %v", asset.GetV2rayLocationAssetOverride())
	wd, _ := os.Getwd()
	log.Alert("v2rayA working directory is %v", wd)
	log.Alert("v2rayA configuration directory is %v", conf.GetEnvironmentConfig().Config)
	log.Alert("Golang: %v", runtime.Version())
	log.Alert("OS: %v", runtime.GOOS)
	log.Alert("Arch: %v", runtime.GOARCH)
	log.Alert("Lite: %v", conf.GetEnvironmentConfig().Lite)
	log.Alert("Version: %v", conf.Version)
	log.Alert("Starting...")
}
