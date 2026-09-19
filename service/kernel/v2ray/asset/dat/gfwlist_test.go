package dat

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInstallGeoSiteFileRejectsInvalidDataWithoutReplacingExistingFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "LoyalsoldierSite.dat")
	downloaded := target + ".new"
	if err := os.WriteFile(target, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(downloaded, []byte("not a dat file"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := installGeoSiteFile(downloaded, target); err == nil {
		t.Fatal("invalid geosite database was installed")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "existing" {
		t.Fatalf("existing file changed to %q", got)
	}
}

type versionTransport func(*http.Request) (*http.Response, error)

func (f versionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRemoteGFWListVersionCacheExpires(t *testing.T) {
	old, fetched := g, gFetchedAt
	defer func() { g, gFetchedAt = old, fetched }()
	g = GFWList{}
	gFetchedAt = time.Time{}
	calls := 0
	client := &http.Client{Transport: versionTransport(func(*http.Request) (*http.Response, error) {
		calls++
		tag := "202609190001"
		if calls > 1 {
			tag = "202609190002"
		}
		return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(strings.NewReader(`[{"name":"` + tag + `"}]`)), Header: make(http.Header)}, nil
	})}
	first, err := GetRemoteGFWListUpdateTime(client)
	if err != nil {
		t.Fatal(err)
	}
	cached, err := GetRemoteGFWListUpdateTime(client)
	if err != nil || cached != first || calls != 1 {
		t.Fatalf("fresh cache: %+v, %v, requests=%d", cached, err, calls)
	}
	gFetchedAt = time.Now().Add(-time.Hour - time.Second)
	next, err := GetRemoteGFWListUpdateTime(client)
	if err != nil {
		t.Fatal(err)
	}
	if next.Tag == first.Tag || calls != 2 {
		t.Fatalf("expired cache returned %+v; requests=%d", next, calls)
	}
}
