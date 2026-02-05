//go:build windows
// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// loadEnvFile 从指定路径加载环境变量配置文件
func loadEnvFile(envFilePath string) error {
	if envFilePath == "" {
		return nil
	}

	// 展开相对路径为绝对路径
	absPath, err := filepath.Abs(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("Windows env file not found: %s (skipping)", absPath)
			return nil
		}
		return fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	log.Info("Loading Windows env file: %s", absPath)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	loadedCount := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 KEY=VALUE 格式
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Warn("Invalid line %d in env file: %s", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 移除可能的引号
		if len(value) >= 2 {
			if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
				value = value[1 : len(value)-1]
			}
		}

		// 设置环境变量
		if err := os.Setenv(key, value); err != nil {
			log.Warn("Failed to set env var %s: %v", key, err)
		} else {
			log.Debug("Loaded env: %s=%s", key, value)
			loadedCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	log.Info("Loaded %d environment variables from %s", loadedCount, absPath)
	return nil
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

	// 读取环境变量配置文件路径
	// 优先使用环境变量 V2RAYA_WIN_ENVFILE
	envFilePath := os.Getenv("V2RAYA_WIN_ENVFILE")
	if envFilePath == "" {
		// 如果环境变量未设置，尝试从命令行参数读取（仅在服务配置中可能有用）
		for i, arg := range os.Args {
			if arg == "--win-envfile" && i+1 < len(os.Args) {
				envFilePath = os.Args[i+1]
				break
			}
		}
	}

	// 加载环境变量文件
	if err := loadEnvFile(envFilePath); err != nil {
		elog.Error(1, fmt.Sprintf("Failed to load env file: %v", err))
		// 不是致命错误，继续运行
		log.Warn("Failed to load env file: %v", err)
	}

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
