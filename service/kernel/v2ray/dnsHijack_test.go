package v2ray

import (
	"github.com/v2rayA/v2rayA/db/configure"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// The hijack overwrites /etc/resolv.conf, so what was there has to come back
// when the proxy stops: a plain file by content, a symlink by target.
func TestBackupAndRestoreResolv(t *testing.T) {
	dir := t.TempDir()
	resolv := filepath.Join(dir, "resolv.conf")
	backup := filepath.Join(dir, "resolv.conf.backup")
	withPaths(t, resolv, backup)

	original := "nameserver 192.168.1.1\nsearch lan\n"
	if err := os.WriteFile(resolv, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := backupResolv(); err != nil {
		t.Fatalf("backup: %v", err)
	}
	if err := os.WriteFile(resolv, []byte(HijackFlag+"\nnameserver 127.2.0.17\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// A second hijack must not overwrite the backup with the hijacked file.
	if err := backupResolv(); err != nil {
		t.Fatalf("second backup: %v", err)
	}
	if !restoreResolv() {
		t.Fatal("restore reported failure")
	}
	b, err := os.ReadFile(resolv)
	if err != nil || string(b) != original {
		t.Fatalf("got %q, want %q (err %v)", b, original, err)
	}
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Error("the backup should be gone after a successful restore")
	}
}

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

func TestRestoreResolvWithoutBackup(t *testing.T) {
	dir := t.TempDir()
	withPaths(t, filepath.Join(dir, "resolv.conf"), filepath.Join(dir, "resolv.conf.backup"))
	if restoreResolv() {
		t.Error("restore must report failure when nothing was saved")
	}
}

func TestRestoreResolvIgnoresEmptyBackup(t *testing.T) {
	dir := t.TempDir()
	resolv := filepath.Join(dir, "resolv.conf")
	backup := filepath.Join(dir, "resolv.conf.backup")
	withPaths(t, resolv, backup)

	current := []byte("nameserver 192.0.2.53\n")
	if err := os.WriteFile(resolv, current, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if restoreResolv() {
		t.Fatal("restore must report failure for an empty backup")
	}
	got, err := os.ReadFile(resolv)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(current) {
		t.Fatalf("resolver changed to %q, want %q", got, current)
	}
}

func withPaths(t *testing.T, resolv, backup string) {
	t.Helper()
	oldResolv, oldBackup := resolvPath, resolvBackupPath
	resolvPath, resolvBackupPath = resolv, backup
	t.Cleanup(func() { resolvPath, resolvBackupPath = oldResolv, oldBackup })
}

func TestResolvHijackerConcurrentResetRemove(t *testing.T) {
	dir := t.TempDir()
	resolv := filepath.Join(dir, "resolv.conf")
	withPaths(t, resolv, filepath.Join(dir, "backup"))
	original := "nameserver 192.0.2.53\n"
	if err := os.WriteFile(resolv, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	previous := configure.GetSettingNotNil()
	setting := *previous
	setting.Transparent = configure.TransparentFollowRule
	if err := configure.SetSetting(&setting); err != nil {
		t.Fatal(err)
	}
	defer configure.SetSetting(previous)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				resetResolvHijacker()
				removeResolvHijacker()
			}
		}()
	}
	wg.Wait()
	removeResolvHijacker()
	got, err := os.ReadFile(resolv)
	if err != nil || string(got) != original {
		t.Fatalf("resolver not restored: %q, %v", got, err)
	}
}

func TestBackupAndRestoreEmptyResolv(t *testing.T) {
	dir := t.TempDir()
	resolv := filepath.Join(dir, "resolv.conf")
	backup := filepath.Join(dir, "resolv.conf.backup")
	withPaths(t, resolv, backup)
	if err := os.WriteFile(resolv, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := backupResolv(); err != nil {
		t.Fatalf("backup: %v", err)
	}
	if err := os.WriteFile(resolv, []byte(HijackFlag+"\nnameserver 127.2.0.17\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if !restoreResolv() {
		t.Fatal("an empty original was not restored")
	}
	if b, err := os.ReadFile(resolv); err != nil || len(b) != 0 {
		t.Fatalf("got %q (err %v), want an empty file", b, err)
	}
}

func TestReadOriginalResolv(t *testing.T) {
	dir := t.TempDir()
	path, backup := filepath.Join(dir, "resolv.conf"), filepath.Join(dir, "backup")
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	original := "nameserver 127.0.0.53\nnameserver 2001:db8::53\nnameserver invalid\n# nameserver 8.8.8.8\n"
	want := []string{"127.0.0.53:53", "[2001:db8::53]:53"}
	check := func(want []string, fails bool) {
		t.Helper()
		found, err := readOriginalResolv(path, backup)
		if (err != nil) != fails || !reflect.DeepEqual(found, want) {
			t.Fatalf("got %v %v; want %v error=%v", found, err, want, fails)
		}
	}
	write(path, original)
	check(want, false)
	write(backup, original)
	write(path, HijackFlag+"\nnameserver 127.2.0.17\nnameserver 223.5.5.5\n")
	check(want, false)
	target := filepath.Join(dir, "real.conf")
	write(target, original)
	write(backup, symlinkMarker+"real.conf\n")
	check(want, false)
	write(backup, symlinkMarker+target+"\n")
	check(want, false)
	write(backup, missingMarker+"\n")
	check(nil, false)
	write(backup, emptyMarker+"\n")
	check(nil, false)
	write(backup, "")
	check(nil, true)
	write(backup, symlinkMarker+"missing\n")
	check(nil, true)
	write(backup, HijackFlag+"\nnameserver 127.2.0.17\n")
	check(nil, true)
}
