package main

import (
	"bytes"
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	_ "github.com/v2rayA/v2rayA/plugin/pingtunnel"
	_ "github.com/v2rayA/v2rayA/plugin/shadowsocksr"
	_ "github.com/v2rayA/v2rayA/plugin/trojan"
	"github.com/v2rayA/v2rayA/service"
	"io/ioutil"
	"moona/config"
	"os"
	"time"
)

const configFilePath = "/tmp/testLatency.json"

func ImportServers(c *config.Params) {
	if c.Link == "" && c.File == "" {
		fmt.Printf("Run '%v --help' for usage.\n", os.Args[0])
		os.Exit(0)
	}
	if c.Link != "" {
		fmt.Printf("Importing %v\n", c.Link)
		err := service.Import(c.Link, nil)
		if err != nil {
			fmt.Printf("Failed to import %v: %v\n", c.Link, err)
			os.Exit(1)
		}
	}
	if c.File != "" {
		b, err := ioutil.ReadFile(c.File)
		if err != nil {
			fmt.Printf("Failed to import %v: %v\n", c.File, err)
			os.Exit(1)
		}
		lines := bytes.Split(b, []byte("\n"))
		for _, line := range lines {
			l := string(bytes.TrimSpace(line))
			if l == "" {
				continue
			}
			err := service.Import(l, nil)
			if err != nil {
				fmt.Printf("Skip %v: %v: %v\n", c.File, l, err)
				continue
			}
		}
	}
}

func ConfigureV2rayA() {
	global.Version = "moona"
	global.DontLoadConfig()
	global.SetConfig(global.Params{
		Address:          "127.0.0.1:20177",
		Config:           configFilePath,
		Mode:             "universal",
		PluginListenPort: 30177,
		PassCheckRoot:    true,
	})
	_ = os.Remove(configFilePath)
}

func GenerateTestList() configure.Whiches {
	t := touch.GenerateTouch()
	var whiches configure.Whiches
	for _, s := range t.Servers {
		whiches.Add(configure.Which{
			TYPE: s.TYPE,
			ID:   s.ID,
			Sub:  0,
		})

	}
	for _, ss := range t.Subscriptions {
		for _, s := range ss.Servers {
			whiches.Add(configure.Which{
				TYPE: s.TYPE,
				ID:   s.ID,
				Sub:  ss.ID - 1,
			})
		}
	}
	err := whiches.FillLinks()
	if err != nil {
		fmt.Printf("Failed to generate links: %v\n", err)
		os.Exit(1)
	}
	return whiches
}

func main() {
	if !common.IsInDocker() {
		fmt.Printf("moona must run with docker")
		os.Exit(1)
	}
	c := config.GetConfig()
	ConfigureV2rayA()
	ImportServers(c)
	testList := GenerateTestList()
	_, err := service.TestHttpLatency(testList.Get(), time.Duration(c.Timeout)*time.Millisecond, c.Parallel, true)
	if err != nil {
		fmt.Printf("Failed to test latencies: %v\n", err)
		os.Exit(1)
	}
}
