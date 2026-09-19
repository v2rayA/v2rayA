package main

import (
	"errors"
	"testing"
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
