//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

const coreShutdownTimeout = 30 * time.Second

// tryRunAsService attempts to run as a Windows service, returns true if successful
func tryRunAsService() bool {
	isWindowsService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal("Failed to check if running as Windows service: %v", err)
	}

	if isWindowsService {
		// Running as Windows service
		if err := runAsService(false); err != nil {
			log.Fatal("Failed to run as Windows service: %v", err)
		}
		return true
	}

	// Check if running in debug mode
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		if err := runAsService(true); err != nil {
			log.Fatal("Failed to run in debug mode: %v", err)
		}
		return true
	}

	return false
}

type v2rayAService struct{}

// Execute implements svc.Handler interface
func (m *v2rayAService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	// Notify service manager that service is starting
	changes <- svc.Status{State: svc.StartPending}

	// Start main program
	errChan := make(chan error, 1)
	go func() {
		// Execute official v2rayA main function
		if err := runService(); err != nil {
			errChan <- err
		}
	}()

	// Notify service manager that service has started
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	elog.Info(1, "v2rayA service started")

	// Wait for service control commands
loop:
	for {
		select {
		case err := <-errChan:
			elog.Error(1, fmt.Sprintf("v2rayA service error: %v", err))
			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				elog.Info(1, "v2rayA service stopping")
				// Graceful shutdown: stop v2ray/xray process first
				cleanupResources()
				break loop
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}

	// Notify service manager that service is stopping
	changes <- svc.Status{State: svc.StopPending}
	return
}

// runAsService runs as a Windows service
func runAsService(isDebug bool) error {
	var err error

	if isDebug {
		elog = debug.New("v2rayA")
	} else {
		elog, err = eventlog.Open("v2rayA")
		if err != nil {
			return fmt.Errorf("failed to create event log: %w", err)
		}
	}
	defer elog.Close()

	elog.Info(1, "v2rayA service starting")

	// Run service
	if isDebug {
		err = debug.Run("v2rayA", &v2rayAService{})
	} else {
		err = svc.Run("v2rayA", &v2rayAService{})
	}

	if err != nil {
		elog.Error(1, fmt.Sprintf("v2rayA service failed: %v", err))
		return err
	}

	elog.Info(1, "v2rayA service stopped")
	return nil
}

// cleanupResources cleans up resources, stops v2ray/xray processes
func cleanupResources() {
	elog.Info(1, "Cleaning up resources...")
	runningCore := v2ray.ProcessManager.Process()

	// Stop transparent proxy
	v2ray.ProcessManager.CheckAndStopTransparentProxy(nil)
	elog.Info(1, "Transparent proxy stopped")

	// Stop v2ray/xray process
	v2ray.ProcessManager.Stop(false)
	elog.Info(1, "v2ray/xray process stopped")
	if runningCore != nil {
		elog.Info(1, "Waiting for v2ray/xray core to exit...")
		ctx, cancel := context.WithTimeout(context.Background(), coreShutdownTimeout)
		defer cancel()
		if err := runningCore.WaitUntilExit(ctx); err != nil {
			elog.Error(1, fmt.Sprintf("Timeout while waiting for v2ray/xray core to exit: %v", err))
		} else {
			elog.Info(1, "v2ray/xray core exit confirmed")
		}
	}

	// Close database connection
	if err := db.DB().Close(); err != nil {
		elog.Error(1, fmt.Sprintf("Failed to close database: %v", err))
	} else {
		elog.Info(1, "Database connection closed")
	}

	elog.Info(1, "Resource cleanup completed")
}

// runService runs the main program in service context
func runService() error {
	// Add short delay to ensure service status updated
	time.Sleep(100 * time.Millisecond)

	// Execute original main function logic
	checkEnvironment()
	if err := checkPlatformSpecific(); err != nil {
		return err
	}
	initConfigure()
	checkUpdate()
	hello()
	return run()
}

// checkPlatformSpecific Windows specific checks
func checkPlatformSpecific() error {
	// Windows doesn't need TProxy checks
	return nil
}
