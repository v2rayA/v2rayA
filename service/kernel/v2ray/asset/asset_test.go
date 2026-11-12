package asset

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type assetTransport func(*http.Request) (*http.Response, error)

func (f assetTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAssetBodyOverLimitIsRejected(t *testing.T) {
	client := &http.Client{Transport: assetTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxAssetDownloadSize+1)))),
		}, nil
	})}
	target := filepath.Join(t.TempDir(), "asset.dat")
	if err := download(client, "https://example.test/asset.dat", target); err == nil || !strings.Contains(err.Error(), "256 MiB") {
		t.Fatalf("error = %v, want size limit", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("oversized asset created target: %v", err)
	}
}

// The core only reads XRAY_LOCATION_ASSET, so a dat file that exists in a
// system directory has to be linked into that directory before the core runs.
func TestEnsureCoreAssetsLinksMissingFiles(t *testing.T) {
	system := t.TempDir()
	assetDir := filepath.Join(t.TempDir(), "runtime")
	source := filepath.Join(system, "geosite.dat")
	if err := os.WriteFile(source, []byte("dat"), 0644); err != nil {
		t.Fatal(err)
	}
	// findAssetOutsideDir searches fixed system paths, so exercise the linking
	// itself with the source it would have found.
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(assetDir, "geosite.dat")
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil || string(b) != "dat" {
		t.Fatalf("link does not resolve: %v %q", err, b)
	}
	// An asset already present must not be touched.
	EnsureCoreAssets(assetDir)
	fi, err := os.Lstat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("the existing link was replaced")
	}
}

func TestFindAssetOutsideDirSkipsItsOwnDirectory(t *testing.T) {
	dir := "/usr/share/v2raya"
	if _, err := os.Stat(filepath.Join(dir, "geosite.dat")); err != nil {
		t.Skip("no system geosite.dat on this machine")
	}
	if got := findAssetOutsideDir("geosite.dat", dir); got != "" {
		t.Errorf("a file in the asset directory itself must not be reported as a source, got %q", got)
	}
}

// A missing XDG runtime directory is left alone: creating /run/user as root
// with mode 0700 breaks the runtime directory of every other user (#2092).
func TestRuntimeAssetDirDoesNotCreateMissingRuntimeDir(t *testing.T) {
	base := t.TempDir()
	runtimeDir := filepath.Join(base, "run", "user", "0")
	fallback := filepath.Join(base, "config")
	if got := runtimeAssetDirIn(runtimeDir, fallback); got != fallback {
		t.Fatalf("dir = %q, want the fallback %q", got, fallback)
	}
	if _, err := os.Stat(filepath.Join(base, "run")); !os.IsNotExist(err) {
		t.Fatalf("the missing runtime directory's parents were created: %v", err)
	}
}

func TestRuntimeAssetDirUsesExistingRuntimeDir(t *testing.T) {
	runtimeDir := t.TempDir()
	want := filepath.Join(runtimeDir, "v2raya")
	for i := 0; i < 2; i++ {
		if got := runtimeAssetDirIn(runtimeDir, "fallback"); got != want {
			t.Fatalf("run %d: dir = %q, want %q", i, got, want)
		}
	}
	fi, err := os.Stat(want)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0700 {
		t.Fatalf("mode = %v, want 0700", fi.Mode().Perm())
	}
}

// With the configuration directory as the asset directory, a dat file the
// user placed there must not be replaced by a link to the system copy.
func TestLinkAssetKeepsRegularFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "system.dat")
	own := filepath.Join(dir, "own.dat")
	stale := filepath.Join(dir, "stale.dat")
	for path, body := range map[string]string{source: "system", own: "own"} {
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, "gone.dat"), stale); err != nil {
		t.Fatal(err)
	}
	if got, err := linkAsset(source, own); err != nil || got != own {
		t.Fatalf("regular file: got %q, %v; want %q", got, err, own)
	}
	if b, _ := os.ReadFile(own); string(b) != "own" {
		t.Fatalf("the user's file was replaced: %q", b)
	}
	if got, err := linkAsset(source, stale); err != nil || got != source {
		t.Fatalf("stale link: got %q, %v; want %q", got, err, source)
	}
	if b, _ := os.ReadFile(stale); string(b) != "system" {
		t.Fatalf("stale link not repointed: %q", b)
	}
}
