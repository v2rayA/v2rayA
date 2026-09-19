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

var recoverHostState = v2ray.RecoverHostState

// recoverPendingHostState tears down what a start that never committed
// left on the host. A teardown that fails is logged, not fatal: refusing
// to start would leave the same state behind with no service to fix it.
func recoverPendingHostState() {
	state, err := configure.GetHostState()
	if err != nil {
		log.Warn("read pending host state: %v", err)
		return
	}
	if state == nil {
		var snapshot interface{}
		found, err := configure.GetSystemProxySnapshot(&snapshot)
		if err != nil || !found {
			return
		}
		state = &configure.HostState{TransparentType: configure.TransparentSystemProxy}
	}
	log.Warn("recovering host state left by an interrupted start (transparent %v)", state.TransparentType)
	if err := recoverHostState(state); err != nil {
		// The marker stays so the next start tries again.
		log.Warn("recover host state: %v", err)
		return
	}
	if err := configure.SetHostState(nil); err != nil {
		log.Warn("clear pending host state: %v", err)
	}
}

func run() error {
	recoverPendingHostState()
	cleanup := func() {
		fmt.Println("Quitting...")
		v2ray.ProcessManager.CheckAndStopTransparentProxy(nil)
		v2ray.ProcessManager.Stop(false)
		_ = db.Close()
	}

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

	// Repair what a crashed or killed previous run left in the operating
	// system before anything else, whether or not the core is started.
	v2ray.CleanupTunResidual()

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
		v2ray.ProcessManager.MarkShuttingDown()
		errch <- nil
	}()
	return waitForShutdown(errch, cleanup)
}

func waitForShutdown(errch <-chan error, cleanup func()) (err error) {
	defer cleanup()
	if err = <-errch; err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}
