package v2ray

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestRunPreStartHookTimesOut(t *testing.T) {
	dir := t.TempDir()
	hook := filepath.Join(dir, "slow-hook")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nsleep 3\n"), 0700); err != nil {
		t.Fatal(err)
	}

	previous := *conf.GetEnvironmentConfig()
	config := previous
	config.CoreHook = hook
	config.Config = dir
	config.CoreStartupTimeout = 1
	conf.SetConfig(config)
	t.Cleanup(func() { conf.SetConfig(previous) })

	started := time.Now()
	err := (&CoreProcessManager{}).runPreStartHook()
	elapsed := time.Since(started)
	if elapsed >= 2*time.Second {
		t.Fatalf("hook returned after %v, want less than 2s", elapsed)
	}
	if err == nil || !strings.Contains(err.Error(), hook) || !strings.Contains(err.Error(), "1s") {
		t.Fatalf("error %q must name hook %q and deadline 1s", err, hook)
	}
}

// Shutdown asks for the transparent proxy teardown twice: pre_run calls it and
// Stop calls it again through beforeStop. Only the first call may run the
// user's hooks and delete the rules.
func TestCheckAndStopTransparentProxyRunsOnce(t *testing.T) {
	var m CoreProcessManager
	if m.transparentOn.Load() {
		t.Fatal("a fresh manager must not claim installed rules")
	}
	// Without a setup, the teardown is a no-op.
	m.CheckAndStopTransparentProxy(&configure.Setting{Transparent: configure.TransparentClose})
	if m.transparentOn.Load() {
		t.Error("the mark must stay off")
	}

	m.transparentOn.Store(true)
	m.CheckAndStopTransparentProxy(&configure.Setting{Transparent: configure.TransparentClose})
	if m.transparentOn.Load() {
		t.Error("the first teardown must clear the mark")
	}

	// A teardown with no setting and no running template cannot remove
	// anything, so the mark has to survive for the next attempt.
	m.transparentOn.Store(true)
	m.CheckAndStopTransparentProxy(nil)
	if !m.transparentOn.Load() {
		t.Error("a teardown that did nothing must keep the mark")
	}
}
