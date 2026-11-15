package dat

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type geoassetBytes struct{}

func (geoassetBytes) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

// Exercise the complete download/checksum/install path without GitHub or a
// file-sized fixture in memory. No core is running, so checksum requests use
// the same default transport as the downloads.
func serveGeoasset(t testing.TB, size int64, checksum string, broken bool) {
	t.Helper()
	old := http.DefaultTransport
	http.DefaultTransport = versionTransport(func(r *http.Request) (*http.Response, error) {
		var body io.Reader = io.LimitReader(geoassetBytes{}, size)
		length := size
		if strings.HasSuffix(r.URL.Path, ".sha256sum") {
			body = strings.NewReader(checksum + "  asset.dat\n")
			length = -1
		} else if broken {
			body = io.MultiReader(strings.NewReader("partial"), failingGeoasset{})
		}
		return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(body), ContentLength: length, Header: make(http.Header)}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = old })
}

type failingGeoasset struct{}

func (failingGeoasset) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestGeoassetUpdatePreservesOldFile(t *testing.T) {
	hash := sha256.Sum256([]byte("xxx"))
	for _, update := range []struct {
		name string
		run  func() error
	}{
		{"geoip.dat", UpdateLocalGeoIP},
		{"geosite.dat", UpdateLocalGeoSite},
		{"LoyalsoldierSite.dat", func() error { _, err := UpdateLocalGFWList(); return err }},
	} {
		for _, mode := range []string{"success", "checksum mismatch", "interrupted"} {
			t.Run(update.name+"/"+mode, func(t *testing.T) {
				old, fetched := g, gFetchedAt
				t.Cleanup(func() { g, gFetchedAt = old, fetched })
				g, gFetchedAt = GFWList{Tag: "test", UpdateTime: time.Now()}, time.Now()
				dir := t.TempDir()
				t.Setenv("XRAY_LOCATION_ASSET", dir)
				target := filepath.Join(dir, update.name)
				if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
					t.Fatal(err)
				}
				checksum := fmt.Sprintf("%x", hash)
				if mode == "checksum mismatch" {
					checksum = strings.Repeat("0", 64)
				}
				serveGeoasset(t, 3, checksum, mode == "interrupted")
				err := update.run()
				want := "old"
				if mode == "success" {
					want = "xxx"
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil {
					t.Fatal("expected failure")
				}
				if got, err := os.ReadFile(target); err != nil || string(got) != want {
					t.Fatalf("target = %q, %v; want %q", got, err, want)
				}
			})
		}
	}
}

func TestCheckSha256ReadError(t *testing.T) {
	if ok, actual := checkSha256(t.TempDir(), ""); ok || actual != "" {
		t.Fatalf("directory accepted: %v %q", ok, actual)
	}
	if ok, actual := checkSha256(filepath.Join(t.TempDir(), "missing"), ""); ok || actual != "" {
		t.Fatalf("missing file accepted: %v %q", ok, actual)
	}
}

func BenchmarkGeoassetUpdate(b *testing.B) {
	const size = 64 << 20
	hash := sha256.New()
	if _, err := io.Copy(hash, io.LimitReader(geoassetBytes{}, size)); err != nil {
		b.Fatal(err)
	}
	serveGeoasset(b, size, fmt.Sprintf("%x", hash.Sum(nil)), false)
	dir := b.TempDir()
	b.Setenv("XRAY_LOCATION_ASSET", dir)
	if err := os.WriteFile(filepath.Join(dir, "geoip.dat"), []byte("old"), 0644); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(size)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := UpdateLocalGeoIP(); err != nil {
			b.Fatal(err)
		}
	}
}
