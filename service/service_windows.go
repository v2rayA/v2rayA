//go:build windows
// +build windows

package main

import (
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

// tryRunAsService 尝试作为 Windows 服务运行，如果成功返回 true
func tryRunAsService() bool {
	isWindowsService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal("Failed to check if running as Windows service: %v", err)
	}

	if isWindowsService {
		// 作为 Windows 服务运行
		if err := runAsService(false); err != nil {
			log.Fatal("Failed to run as Windows service: %v", err)
		}
		return true
	}

	// 检查是否以 debug 模式运行服务
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		if err := runAsService(true); err != nil {
			log.Fatal("Failed to run in debug mode: %v", err)
		}
		return true
	}

	return false
}

type v2rayAService struct{}

// Execute 实现 svc.Handler 接口
func (m *v2rayAService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	// 通知服务管理器服务正在启动
	changes <- svc.Status{State: svc.StartPending}

	// 启动主程序
	errChan := make(chan error, 1)
	go func() {
		// 执行实际的 v2rayA 主函数
		if err := runService(); err != nil {
			errChan <- err
		}
	}()

	// 通知服务管理器服务已启动
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	elog.Info(1, "v2rayA service started")

	// 等待服务控制命令
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
				// 优雅关闭：先停止 v2ray/xray 进程
				cleanupResources()
				break loop
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}

	// 通知服务管理器服务正在停止
	changes <- svc.Status{State: svc.StopPending}
	return
}

// runAsService 作为 Windows 服务运行
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

	// 运行服务
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

// cleanupResources 清理资源，关闭 v2ray/xray 进程
func cleanupResources() {
	elog.Info(1, "Cleaning up resources...")

	// 停止透明代理
	v2ray.ProcessManager.CheckAndStopTransparentProxy(nil)
	elog.Info(1, "Transparent proxy stopped")

	// 停止 v2ray/xray 进程
	v2ray.ProcessManager.Stop(false)
	elog.Info(1, "v2ray/xray process stopped")

	// 关闭数据库连接
	if err := db.DB().Close(); err != nil {
		elog.Error(1, fmt.Sprintf("Failed to close database: %v", err))
	} else {
		elog.Info(1, "Database connection closed")
	}

	elog.Info(1, "Resource cleanup completed")
}

// runService 在服务环境中运行主程序
func runService() error {
	// 添加短暂延迟确保服务状态已更新
	time.Sleep(100 * time.Millisecond)

	// 执行原来的 main 函数逻辑
	checkEnvironment()
	if err := checkPlatformSpecific(); err != nil {
		return err
	}
	initConfigure()
	checkUpdate()
	hello()
	return run()
}

// checkPlatformSpecific Windows 平台特定检查
func checkPlatformSpecific() error {
	// Windows 不需要 TProxy 检查
	return nil
}
