package main

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	service2 "github.com/v2rayA/v2rayA/kernel/v2ray/service"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/pkg/util/privilege"
)

func checkEnvironment() {
	config := conf.GetEnvironmentConfig()
	if len(config.PrintReport) > 0 {
		db.SetReadOnly()
		config.Report()
		os.Exit(0)
	}
	if !config.PassCheckRoot {
		switch runtime.GOOS {
		case "linux":
			if !privilege.IsRootOrAdmin() && !config.Lite {
				log.Fatal("Please execute this program with sudo or as a root user for the best experience.\n" +
					"If you are sure you are root user, use the --passcheckroot parameter to skip the check.\n" +
					"If you don't want to run as root or you are a non-linux user, use --lite please.\n" +
					"For example:\n" +
					"$ v2raya --lite",
				)
			}
		case "windows":
			if !privilege.IsRootOrAdmin() && !config.Lite {
				log.Fatal("Please run v2rayA as Administrator (or SYSTEM) with elevation, or start with --lite to skip privilege checks.")
			}
		}
	}
	if config.ResetPassword {
		fmt.Println("Config directory:", config.Config)
		fmt.Println("Resetting password...\nIf no response for a long time, please stop other v2rayA instances and try again.")
		err := configure.ResetAccounts()
		if err != nil {
			log.Fatal("checkEnvironment: %v", err)
		}
		fmt.Println("Succeed. It will work after you restart v2rayA.")
		os.Exit(0)
	}
	_, v2rayAListeningPort, err := net.SplitHostPort(config.Address)
	if err != nil {
		log.Fatal("checkEnvironment: %v", err)
	}
	if occupied, sockets, err := ports.IsPortOccupied([]string{v2rayAListeningPort + ":tcp"}); occupied {
		if err != nil {
			log.Fatal("netstat: %v", err)
		}
		for _, socket := range sockets {
			process, err := socket.Process()
			if err == nil {
				log.Fatal("Port %v is occupied by %v/%v", v2rayAListeningPort, process.Name, process.PID)
			}
		}
	}
}

func checkTProxySupportability() {
	if conf.GetEnvironmentConfig().Lite {
		return
	}
	// check if tproxy can be enabled
	if err := service2.CheckAndProbeTProxy(); err != nil {
		log.Info("Cannot load TPROXY module: %v", err)
	}
}
