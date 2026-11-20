package v2ray

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCoreParentDeathHelper(t *testing.T) {
	role := os.Getenv("V2RAYA_PARENT_DEATH_TEST")
	if role == "" {
		return
	}
	ready := os.Getenv("V2RAYA_PARENT_DEATH_READY")
	if role == "core" {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()
		if err := os.WriteFile(ready, []byte(fmt.Sprintf("%d %s", os.Getpid(), listener.Addr())), 0600); err != nil {
			t.Fatal(err)
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Fatal(err)
			}
			conn.Close()
		}
	}
	proc, err := RunWithLog(context.Background(), os.Args[0],
		[]string{os.Args[0], "-test.run=^TestCoreParentDeathHelper$"}, "",
		append(os.Environ(), "V2RAYA_PARENT_DEATH_TEST=core"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = proc.Wait()
}

func TestCoreDiesWithParent(t *testing.T) {
	start := func() (*exec.Cmd, string) {
		t.Helper()
		ready := filepath.Join(t.TempDir(), "ready")
		cmd := exec.Command(os.Args[0], "-test.run=^TestCoreParentDeathHelper$")
		cmd.Env = append(os.Environ(), "V2RAYA_PARENT_DEATH_TEST=parent", "V2RAYA_PARENT_DEATH_READY="+ready)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			data, err := os.ReadFile(ready)
			if err == nil {
				fields := strings.Fields(string(data))
				if len(fields) == 2 {
					pid, err := strconv.Atoi(fields[0])
					if err != nil {
						t.Fatal(err)
					}
					// Also clean up the deliberately orphaned child when testing without the fix.
					child, err := os.FindProcess(pid)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = child.Kill(); _ = child.Release() })
					return cmd, fields[1]
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("core did not open its listener")
		return nil, ""
	}
	first, firstAddr := start()
	_, secondAddr := start()
	if err := first.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	var listener net.Listener
	var err error
	for time.Now().Before(deadline) {
		listener, err = net.Listen("tcp", firstAddr)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("killed parent's core still holds %s: %v", firstAddr, err)
	}
	listener.Close()
	conn, err := net.DialTimeout("tcp", secondAddr, time.Second)
	if err != nil {
		t.Fatalf("second core was affected: %v", err)
	}
	conn.Close()
}
