//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/router"
	"golang.org/x/sys/windows/svc"
	svcdbg "golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog svcdbg.Log

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
		defer func() {
			if r := recover(); r != nil {
				elog.Error(1, fmt.Sprintf("panic in runService: %v\n%s", r, debug.Stack()))
				errChan <- fmt.Errorf("panic: %v", r)
			}
		}()
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
				changes <- svc.Status{State: svc.StopPending}
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
		elog = svcdbg.New("v2rayA")
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
		err = svcdbg.Run("v2rayA", &v2rayAService{})
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

// cleanupResources cleans up resources with an overall 30-second safety timeout.
// Individual cleanup steps are logged so Windows Event Log shows exactly where
// a hang occurs if the SCM deadline approaches.
func cleanupResources() {
	const overallTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), overallTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				elog.Error(1, fmt.Sprintf("panic during cleanup: %v\n%s", r, debug.Stack()))
			}
			close(done)
		}()
		doCleanup()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		elog.Error(1, fmt.Sprintf("cleanup timed out after %v, forcing exit", overallTimeout))
	}
}

// doCleanup performs the actual resource teardown.
func doCleanup() {
	elog.Info(1, "Cleaning up resources...")

	// Step 1: stop accepting new HTTP requests so the frontend receives no
	// further responses after this point.
	elog.Info(1, "Shutting down HTTP server...")
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()
	if err := router.Shutdown(httpCtx); err != nil {
		elog.Error(1, fmt.Sprintf("HTTP server shutdown error: %v", err))
	} else {
		elog.Info(1, "HTTP server stopped")
	}

	// Step 2: capture the running core before Stop() clears it.
	runningCore := v2ray.ProcessManager.Process()

	// Step 3: stop v2ray/xray and transparent proxy (beforeStop calls
	// CheckAndStopTransparentProxy internally, so no explicit call needed here).
	elog.Info(1, "Stopping v2ray/xray process...")
	v2ray.ProcessManager.Stop(false)
	elog.Info(1, "v2ray/xray process stopped")

	if runningCore != nil {
		elog.Info(1, "Waiting for v2ray/xray core to exit...")
		waitCtx, waitCancel := context.WithTimeout(context.Background(), coreShutdownTimeout)
		defer waitCancel()
		if err := runningCore.WaitUntilExit(waitCtx); err != nil {
			elog.Error(1, fmt.Sprintf("Timeout while waiting for v2ray/xray core to exit: %v", err))
		} else {
			elog.Info(1, "v2ray/xray core exit confirmed")
		}
	}

	// Step 4: close the database.
	elog.Info(1, "Closing database...")
	if err := db.Close(); err != nil {
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
