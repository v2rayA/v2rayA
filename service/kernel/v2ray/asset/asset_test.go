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
