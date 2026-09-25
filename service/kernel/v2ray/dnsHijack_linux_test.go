package v2ray

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupAndRestoreResolvSymlink(t *testing.T) {
	dir := t.TempDir()
	resolv := filepath.Join(dir, "resolv.conf")
	backup := filepath.Join(dir, "resolv.conf.backup")
	withPaths(t, resolv, backup)

	stub := filepath.Join(dir, "stub-resolv.conf")
	if err := os.WriteFile(stub, []byte("nameserver 127.0.0.53\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(stub, resolv); err != nil {
		t.Fatal(err)
	}
	hij := &ResolvHijacker{}
	if err := hij.HijackResolv(); err != nil {
		t.Fatalf("hijack: %v", err)
	}
	// The hijack must replace the link, not write through it into the
	// resolver's own file.
	if b, err := os.ReadFile(stub); err != nil || string(b) != "nameserver 127.0.0.53\n" {
		t.Fatalf("the link target was overwritten: %q (err %v)", b, err)
	}
	saved, err := os.ReadFile(backup)
	if err != nil || !strings.HasPrefix(string(saved), symlinkMarker) {
		t.Fatalf("the link target was not saved: %q (err %v)", saved, err)
	}
	if !restoreResolv() {
		t.Fatal("restore reported failure")
	}
	fi, err := os.Lstat(resolv)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the symlink is not back: %v (err %v)", fi, err)
	}
	target, err := os.Readlink(resolv)
	if err != nil || target != stub {
		t.Fatalf("link points at %q, want %q (err %v)", target, stub, err)
	}
}
