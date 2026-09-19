package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

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
