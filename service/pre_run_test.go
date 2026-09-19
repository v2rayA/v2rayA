package main

import (
	"errors"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestWaitForShutdownCleansUpBeforeReturningRouterError(t *testing.T) {
	routerErr := errors.New("listener failed")
	errch := make(chan error, 1)
	errch <- routerErr
	cleaned := false

	err := waitForShutdown(errch, func() { cleaned = true })
	if !cleaned {
		t.Fatal("cleanup had not run before waitForShutdown returned")
	}
	if !errors.Is(err, routerErr) {
		t.Fatalf("error %v does not wrap router error %v", err, routerErr)
	}
}

func TestStartupRecoversMarkedHostState(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	oldConfig, oldRecover := env.Config, recoverHostState
	env.Config = t.TempDir()
	t.Cleanup(func() { env.Config, recoverHostState = oldConfig, oldRecover })
	for _, tc := range []struct {
		name         string
		marked, fail bool
	}{
		{"marked-success", true, false},
		{"marked-failure", true, true},
		{"snapshot-success", false, false},
		{"snapshot-failure", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := configure.SetHostState(nil); err != nil {
				t.Fatal(err)
			}
			if err := configure.SetSystemProxySnapshot(nil); err != nil {
				t.Fatal(err)
			}
			want := configure.HostState{TransparentType: configure.TransparentTun, APIPort: 23456, TunTeardownScript: "restore-routes"}
			if tc.marked {
				if err := configure.SetHostState(&want); err != nil {
					t.Fatal(err)
				}
			} else {
				want = configure.HostState{TransparentType: configure.TransparentSystemProxy}
				if err := configure.SetSystemProxySnapshot(map[string]string{"original": "proxy"}); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			recoverHostState = func(got *configure.HostState) error {
				called = true
				if *got != want {
					t.Fatalf("recovered %+v, want %+v", got, want)
				}
				if tc.fail {
					return errors.New("recovery failed")
				}
				return configure.SetSystemProxySnapshot(nil)
			}
			// Recovery never blocks the start; a failed one keeps its marker.
			recoverPendingHostState()
			if !called {
				t.Error("pending host state was not recovered")
			}
			got, err := configure.GetHostState()
			if err != nil {
				t.Fatal(err)
			}
			if tc.marked && tc.fail && (got == nil || *got != want) {
				t.Error("failed recovery lost its marker")
			}
			if !tc.fail && got != nil {
				t.Error("successful recovery retained its marker")
			}
			var snapshot map[string]string
			found, err := configure.GetSystemProxySnapshot(&snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if found != (!tc.marked && tc.fail) {
				t.Errorf("snapshot present=%v after recovery", found)
			}
		})
	}
}
