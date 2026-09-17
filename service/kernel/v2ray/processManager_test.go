package v2ray

import (
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

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
