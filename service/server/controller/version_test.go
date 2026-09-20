package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
)

func TestGetSettingReportsGFWListAndGeoSiteVersionsSeparately(t *testing.T) {
	assetDir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", assetDir)
	date := time.Date(2026, time.September, 15, 12, 0, 0, 0, time.Local)
	for _, name := range []string{"LoyalsoldierSite.dat", "geosite.dat"} {
		path := filepath.Join(assetDir, name)
		if err := os.WriteFile(path, []byte("dat"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, date, date); err != nil {
			t.Fatal(err)
		}
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/setting", nil)
	GetSetting(ctx)
	var response struct {
		Data struct {
			LocalGFWListVersion string `json:"localGFWListVersion"`
			LocalGeositeVersion string `json:"localGeositeVersion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.LocalGFWListVersion != "2026-09-15" {
		t.Errorf("localGFWListVersion = %q, want 2026-09-15", response.Data.LocalGFWListVersion)
	}
	if response.Data.LocalGeositeVersion != "2026-09-15" {
		t.Errorf("localGeositeVersion = %q, want 2026-09-15", response.Data.LocalGeositeVersion)
	}
}

func TestGetVersionIncludesDockerFlag(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/version", nil)
	GetVersion(ctx)
	var response struct {
		Data struct {
			Docker *bool `json:"docker"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Docker == nil || *response.Data.Docker != common.IsDocker() {
		t.Fatalf("docker = %v, want %v", response.Data.Docker, common.IsDocker())
	}
}

type versionTransport func(*http.Request) (*http.Response, error)

func (f versionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRemoteVersionRequestHasDeadline(t *testing.T) {
	old := http.DefaultTransport
	defer func() { http.DefaultTransport = old }()
	called := false
	http.DefaultTransport = versionTransport(func(r *http.Request) (*http.Response, error) {
		called = true
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("remote version request has no deadline")
		}
		return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(strings.NewReader(`[]`)), Header: make(http.Header)}, nil
	})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	GetRemoteGFWListVersion(ctx)
	if !called {
		t.Fatal("remote version was not queried")
	}
}
