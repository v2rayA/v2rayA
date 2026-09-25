package v2ray

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestProcessCloseReapsCoreBeforeReturning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses the POSIX sleep executable")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd, err := RunWithLog(ctx, "sleep", []string{"sleep", "30"}, "", os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	p := &Process{proc: cmd.Process, procCancel: cancel, procDone: make(chan struct{}), template: &Template{}}
	go func() {
		_ = cmd.Wait()
		close(p.procDone)
	}()
	finished := make(chan error, 1)
	go func() { finished <- p.Close() }()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("core shutdown did not finish")
	}
	select {
	case <-p.procDone:
	default:
		t.Fatal("Close returned before the core was reaped")
	}
	if cmd.ProcessState == nil {
		t.Fatal("command wait did not complete")
	}
	if err := p.Close(); err != nil {
		t.Fatalf("repeated Close: %v", err)
	}
}
