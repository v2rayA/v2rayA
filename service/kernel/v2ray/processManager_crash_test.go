package v2ray

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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

func TestStartRetainsFailedRecovery(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	env.Lite = true
	t.Cleanup(func() { *env = previous; _ = configure.SetHostState(nil) })
	marker := &configure.HostState{TransparentType: configure.TransparentTun, APIPort: 23456, TunTeardownScript: "exit 23"}
	if err := configure.SetHostState(marker); err != nil {
		t.Fatal(err)
	}
	var m CoreProcessManager
	if err := m.Start(&Template{Setting: configure.NewSetting()}); err == nil {
		t.Fatal("Start ignored failed pending recovery")
	}
	got, err := configure.GetHostState()
	if err != nil || got == nil || *got != *marker {
		t.Fatalf("pending recovery was replaced: marker=%+v err=%v", got, err)
	}
	if m.p != nil || m.generation != 0 {
		t.Fatal("Start progressed after failed recovery")
	}
}
