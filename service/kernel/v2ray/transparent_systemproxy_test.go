//go:build !windows && !darwin

package v2ray

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

// TestDeleteTransparentProxyRulesWithoutSnapshotDisablesProxy is the
// regression test for the mutual-exclusion fix.
//
// deleteTransparentProxyRules used to bail out before touching the system proxy
// whenever no snapshot was persisted. The proxy therefore survived the switch
// still pointing at 52345 — a port only the system_proxy mode listens on — so
// after switching back to tun every proxy-using application sat in SYN_SENT
// with no way out, and had to be restarted to recover.
//
// This drives the real entry point rather than GetCleanCommands directly, so
// it fails against the old code and passes against the new.
func TestDeleteTransparentProxyRulesWithoutSnapshotDisablesProxy(t *testing.T) {
	env := conf.GetEnvironmentConfig()
	previous := *env
	env.Config, env.Lite = t.TempDir(), true
	t.Cleanup(func() { *env = previous })

	// Point PATH at a directory holding a gsettings stub that records calls,
	// so the cleanup drives the stub instead of the real desktop.
	dir := t.TempDir()
	output := filepath.Join(dir, "calls")
	t.Setenv("PATH", dir)
	if err := os.Symlink("/bin/sh", filepath.Join(dir, "sh")); err != nil {
		t.Fatal(err)
	}
	stub := "#!/bin/sh\nif [ \"$1\" = get ]; then echo \"'auto'\"; else echo \"$*\" >> \"" +
		output + "\"; fi\n"
	if err := os.WriteFile(filepath.Join(dir, "gsettings"), []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}

	// No snapshot was ever captured — the exact state that used to skip cleanup.
	if err := configure.SetSystemProxySnapshot(nil); err != nil {
		t.Fatal(err)
	}
	var snapshot interface{}
	if found, err := configure.GetSystemProxySnapshot(&snapshot); err != nil || found {
		t.Fatalf("expected no snapshot, found=%v err=%v", found, err)
	}

	// The call under test.
	deleteTransparentProxyRules()

	b, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("the system proxy was never touched: %v", err)
	}
	// gsettings takes the value unquoted, so the recorded line reads
	// "set org.gnome.system.proxy mode none".
	if !strings.Contains(string(b), "mode none") {
		t.Errorf("proxy was not disabled without a snapshot: %s", b)
	}
}