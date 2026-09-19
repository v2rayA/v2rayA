package v2ray

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestCoreExitFixture(t *testing.T) {
	addr := os.Getenv("V2RAYA_EXIT_FIXTURE_ADDR")
	if addr == "" {
		return
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	c, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	waitCrashCondition(t, func() bool { _, err := os.Stat(os.Getenv("V2RAYA_EXIT_FIXTURE_HOOK")); return err == nil })
}

func waitCrashCondition(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !ready() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for lifecycle fixture")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestStartRejectsExitDuringPostStart(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	t.Cleanup(func() { *env = previous })
	dir := t.TempDir()
	env.Config, env.V2rayAssetsDirectory, env.CoreStartupTimeout = dir, dir, 3
	env.Lite = true
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	core, hook := filepath.Join(dir, "core"), filepath.Join(dir, "hook")
	entered, release := filepath.Join(dir, "entered"), filepath.Join(dir, "release")
	if err := os.WriteFile(core, []byte(fmt.Sprintf("#!/bin/sh\nexec %q -test.run=^TestCoreExitFixture$\n", bin)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hook, []byte(fmt.Sprintf("#!/bin/sh\ncase \"$1\" in --stage=post-start) : > %q; while [ ! -f %q ]; do /bin/sleep 0.01; done;; esac\n", entered, release)), 0700); err != nil {
		t.Fatal(err)
	}
	env.V2rayBin, env.CoreHook = core, hook
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	t.Setenv("V2RAYA_EXIT_FIXTURE_ADDR", fmt.Sprintf("127.0.0.1:%d", port))
	t.Setenv("V2RAYA_EXIT_FIXTURE_HOOK", entered)
	setting := configure.NewSetting()
	setting.Transparent = configure.TransparentClose
	var m CoreProcessManager
	result := make(chan error, 1)
	t.Cleanup(func() { _ = os.WriteFile(release, nil, 0600); m.Stop(true) })
	go func() { result <- m.Start(&Template{API: &coreObj.APIObject{}, ApiPort: port, Setting: setting}) }()
	waitCrashCondition(t, func() bool { _, err := os.Stat(entered); return err == nil })
	waitCrashCondition(t, func() bool { return !m.Running() })
	if err := os.WriteFile(release, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err == nil {
		t.Error("Start returned success after its process exited during the post-start hook")
	}
	if configure.GetRunning() {
		t.Error("Start persisted running=true after its process exited")
	}
}

func TestStaleGenerationCannotMutateCurrentProcess(t *testing.T) {
	p := &Process{template: &Template{Setting: tunSetting()}}
	m := CoreProcessManager{p: p, generation: 2}
	for _, stale := range []struct {
		p          *Process
		generation uint64
	}{{p, 1}, {&Process{template: p.template}, 2}} {
		mutated := false
		if err := m.mutateHost(stale.p, stale.generation, func() error { mutated = true; return nil }); err == nil || mutated {
			t.Errorf("stale owner mutated host: err=%v mutated=%v", err, mutated)
		}
		if stop, _ := m.syncConnectivityState(stale.p, stale.generation); !stop {
			t.Error("stale connectivity monitor did not stop")
		}
		m.startConnectivityMonitor(stale.p, stale.generation)
		if m.connectivityStop != nil || m.networkPaused || m.p != p {
			t.Fatal("stale monitor changed current process state")
		}
	}
}

func TestStartContinuesPastFailedRecovery(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	env.Lite = true
	env.Config = t.TempDir()
	t.Cleanup(func() { *env = previous; _ = configure.SetHostState(nil) })
	marker := &configure.HostState{TransparentType: configure.TransparentTun, APIPort: 23456, TunTeardownScript: "exit 23"}
	if err := configure.SetHostState(marker); err != nil {
		t.Fatal(err)
	}
	// A recovery that fails is logged, the marker cleared, and the start
	// goes on: with no core binary it fails later, at the process.
	var m CoreProcessManager
	err := m.Start(&Template{Setting: configure.NewSetting(), API: &coreObj.APIObject{}})
	if err == nil || strings.Contains(err.Error(), "recover pending host state") {
		t.Fatalf("Start error = %v, want a failure after recovery", err)
	}
	if got, err := configure.GetHostState(); err != nil || got != nil {
		t.Fatalf("marker after start: %+v err=%v", got, err)
	}
	if m.generation != 1 {
		t.Fatalf("generation = %d, want 1: Start did not progress past recovery", m.generation)
	}
}

func TestCoreExitDuringShutdownIsNotACrash(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	env.Config = t.TempDir()
	t.Cleanup(func() { *env = previous })
	if err := configure.SetRunning(true); err != nil {
		t.Fatal(err)
	}
	if err := configure.SetLastKernelExitStatus(configure.LastKernelExitRunning); err != nil {
		t.Fatal(err)
	}
	var m CoreProcessManager
	p := &Process{template: &Template{Setting: configure.NewSetting()}, done: make(chan struct{})}
	close(p.done)
	m.p = p
	// without the mark the same exit is recorded as a crash
	m.handleUnexpectedStop(p)
	if configure.GetLastKernelExitStatus() != configure.LastKernelExitCrashed {
		t.Fatalf("status %v after an unexpected core exit", configure.GetLastKernelExitStatus())
	}
	if err := configure.SetRunning(true); err != nil {
		t.Fatal(err)
	}
	if err := configure.SetLastKernelExitStatus(configure.LastKernelExitRunning); err != nil {
		t.Fatal(err)
	}
	m.p = p
	m.MarkShuttingDown()
	m.handleUnexpectedStop(p)
	if !configure.GetRunning() || configure.GetLastKernelExitStatus() != configure.LastKernelExitRunning {
		t.Fatalf("running=%v status=%v after a core exit during shutdown", configure.GetRunning(), configure.GetLastKernelExitStatus())
	}
}
