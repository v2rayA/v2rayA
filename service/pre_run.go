package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/ipforward"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/router"
	"github.com/v2rayA/v2rayA/server/service"
)

func run() (err error) {
	// Check the last kernel exit status to decide startup behavior.
	lastExit := configure.GetLastKernelExitStatus()
	shouldStart := configure.GetRunning()

	switch lastExit {
	case configure.LastKernelExitCrashed:
		log.Warn("v2ray-core exited abnormally the last time; check the logs for details")
		// Even if the kernel crashed, the running flag was set to false by
		// handleUnexpectedStop, so shouldStart will be false. We do NOT attempt
		// to auto-start after a crash to give the user a chance to inspect.
		if shouldStart {
			// This shouldn't normally happen if handleUnexpectedStop correctly
			// cleared the flag, but be defensive.
			log.Warn("the running flag was left set after a crash; clearing it")
			_ = configure.SetRunning(false)
			shouldStart = false
		}
	case configure.LastKernelExitRunning:
		// The kernel was running when v2rayA exited — auto-start.
		log.Info("v2ray-core was running when v2rayA last exited; attempting to restore")
	default:
		// LastKernelExitStopped or empty (fresh install / legacy data).
		log.Info("v2ray-core was not running when v2rayA last exited")
	}

	if shouldStart {
		//configure the ip forward
		setting := service.GetSetting()
		if setting.IpForward != ipforward.IsIpForwardOn() {
			e := ipforward.WriteIpForward(setting.IpForward)
			if e != nil {
				log.Warn("Connect: %v", e)
			}
		}
		// Start.
		err := v2ray.UpdateV2RayConfig()
		if err != nil {
			log.Error("failed to start v2ray-core: %v", err)
		}
	}
	//w := configure.GetConnectedServers()
	//log.Println(err, ", which:", w)
	//_ = configure.ClearConnected()
	errch := make(chan error)
	// start server
	go func() {
		errch <- router.Run()
	}()
	// listen for signals to handle transparent proxy shutdown
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGILL)
		<-sigs
		errch <- nil
	}()
	if err = <-errch; err != nil {
		log.Fatal("run: %v", err)
	}
	fmt.Println("Quitting...")
	v2ray.ProcessManager.CheckAndStopTransparentProxy(nil)
	v2ray.ProcessManager.Stop(false)
	_ = db.Close()
	return nil
}
