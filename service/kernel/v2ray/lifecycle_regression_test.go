package v2ray

import (
	"net"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestRunningConcurrentLifecycle(t *testing.T) {
	var m CoreProcessManager
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 10000 {
			m.mu.Lock()
			m.p = &Process{}
			m.mu.Unlock()
			m.mu.Lock()
			m.p = nil
			m.mu.Unlock()
		}
	}()
	for range 10000 {
		m.Running()
	}
	wg.Wait()
}

func TestProcessCloseAfterReaped(t *testing.T) {
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	p := &Process{proc: cmd.Process, template: &Template{}, procCancel: func() {}}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestNewProcessEarlyExit(t *testing.T) {
	bin, err := exec.LookPath("false")
	if err != nil {
		t.Fatal(err)
	}
	env := conf.GetEnvironmentConfig()
	previous := *env
	t.Cleanup(func() { *env = previous })
	env.Config = t.TempDir()
	env.V2rayAssetsDirectory = env.Config
	env.V2rayBin = bin
	env.CoreStartupTimeout = 2
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	exited := make(chan struct{})
	_, err = NewProcess(&Template{API: &coreObj.APIObject{}, ApiPort: port}, func() error { return nil }, func() error { return nil }, func(*Process) { close(exited) })
	<-exited
	if err == nil || !strings.Contains(err.Error(), "exited right after starting") {
		t.Fatalf("startup error: %v", err)
	}
}
