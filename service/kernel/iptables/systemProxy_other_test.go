//go:build !windows && !darwin

package iptables

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestSystemProxyRestoresPersistedSnapshot(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	env.Config, env.Lite = t.TempDir(), true
	t.Cleanup(func() { *env = previous })
	for _, outcome := range []string{"success", "command-failure", "missing-tool"} {
		t.Run(outcome, func(t *testing.T) {
			dir := t.TempDir()
			tool, output := filepath.Join(dir, "gsettings"), filepath.Join(dir, "calls")
			t.Setenv("PATH", dir)
			t.Setenv("PROXY_CALLS", output)
			if err := os.Symlink("/bin/sh", filepath.Join(dir, "sh")); err != nil {
				t.Fatal(err)
			}
			fake := "#!/bin/sh\nif [ \"$1\" = get ]; then case \"$3\" in mode) echo \"'auto'\";; host) echo \"'original.example'\";; port) echo 8123;; esac; else echo \"$*\" >> \"$PROXY_CALLS\"; [ \"$PROXY_FAIL\" != yes ]; fi\n"
			if err := os.WriteFile(tool, []byte(fake), 0700); err != nil {
				t.Fatal(err)
			}
			if err := configure.SetSystemProxySnapshot(nil); err != nil {
				t.Fatal(err)
			}
			savedLinuxProxy.saved = false
			if err := SystemProxy.GetSetupCommands().Run(true); err != nil {
				t.Fatal(err)
			}
			var snapshot map[string]interface{}
			found, err := configure.GetSystemProxySnapshot(&snapshot)
			if err != nil || !found {
				t.Errorf("capture was not persisted: found=%v err=%v", found, err)
			}
			// a second setup keeps the pending snapshot as the original
			// instead of refusing, so a start after a crash still works
			if err := SystemProxy.GetSetupCommands().Run(true); err != nil {
				t.Errorf("setup with a pending snapshot failed: %v", err)
			}
			var again map[string]interface{}
			if _, err := configure.GetSystemProxySnapshot(&again); err != nil || fmt.Sprint(again) != fmt.Sprint(snapshot) {
				t.Errorf("setup overwrote the original proxy state: %v", again)
			}
			savedLinuxProxy = linuxProxySavedState{}
			if err := os.WriteFile(output, nil, 0600); err != nil {
				t.Fatal(err)
			}
			if outcome == "command-failure" {
				t.Setenv("PROXY_FAIL", "yes")
			}
			if outcome == "missing-tool" {
				if err := os.Remove(tool); err != nil {
					t.Fatal(err)
				}
			}
			err = SystemProxy.GetCleanCommands().Run(true)
			if outcome == "success" {
				if err != nil {
					t.Fatal(err)
				}
				b, err := os.ReadFile(output)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(b), "mode 'auto'") || !strings.Contains(string(b), "host 'original.example'") || !strings.Contains(string(b), "port 8123") {
					t.Errorf("original settings were not restored: %s", b)
				}
			} else if err == nil {
				t.Error("failed restoration returned success")
			}
			found, err = configure.GetSystemProxySnapshot(&snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if found != (outcome != "success") {
				t.Errorf("persisted snapshot present=%v after %s", found, outcome)
			}
		})
	}
	if err := configure.SetSystemProxySnapshot(nil); err != nil {
		t.Fatal(err)
	}
}
