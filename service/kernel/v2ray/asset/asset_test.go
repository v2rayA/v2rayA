package asset

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adrg/xdg"
)

func TestGetV2rayLocationAssetKeepsAssetWhenDataAndRuntimeDirsMatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG runtime symlinks are not used on Windows")
	}

	baseDir := t.TempDir()
	setXDGAssetDirs(t, baseDir, baseDir)

	assetPath := filepath.Join(baseDir, "v2raya", "geoip.dat")
	if err := os.MkdirAll(filepath.Dir(assetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const content = "geoip data"
	if err := os.WriteFile(assetPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	gotPath, err := GetV2rayLocationAsset("geoip.dat")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != assetPath {
		t.Fatalf("asset path = %q, want %q", gotPath, assetPath)
	}

	info, err := os.Lstat(assetPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("asset was replaced by a self-referential symlink")
	}
	gotContent, err := os.ReadFile(assetPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotContent) != content {
		t.Fatalf("asset content = %q, want %q", gotContent, content)
	}
}

func TestGetV2rayLocationAssetDoesNotCreateSelfSymlinkForMissingAsset(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG runtime symlinks are not used on Windows")
	}

	baseDir := t.TempDir()
	setXDGAssetDirs(t, baseDir, baseDir)

	assetPath := filepath.Join(baseDir, "v2raya", "geosite.dat")
	gotPath, err := GetV2rayLocationAsset("geosite.dat")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != assetPath {
		t.Fatalf("asset path = %q, want %q", gotPath, assetPath)
	}
	if _, err := os.Lstat(assetPath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing asset should not be created as a symlink, got error %v", err)
	}
}

func setXDGAssetDirs(t *testing.T, dataHome, runtimeDir string) {
	t.Helper()

	// Register this before t.Setenv so the environment is restored before
	// xdg reloads its package-level directory state during cleanup.
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("XDG_DATA_DIRS", "")
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("XRAY_LOCATION_ASSET", "")
	xdg.Reload()
}
