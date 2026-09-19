package service

import (
	"errors"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

type updateTransport func(*http.Request) (*http.Response, error)

func (f updateTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type failedUpdateBody struct{ closed bool }

func (b *failedUpdateBody) Read(p []byte) (int, error) {
	return copy(p, "Package: v2raya\n"), errors.New("interrupted")
}
func (b *failedUpdateBody) Close() error { b.closed = true; return nil }

func TestCheckUpdateClosesFailedBody(t *testing.T) {
	old := http.DefaultClient
	defer func() { http.DefaultClient = old }()
	body := &failedUpdateBody{}
	http.DefaultClient = &http.Client{Transport: updateTransport(func(r *http.Request) (*http.Response, error) {
		arch := runtime.GOARCH
		switch arch {
		case "386":
			arch = "i386"
		case "arm":
			arch = "armhf"
		case "mipsle":
			arch = "mips32le"
		}
		if !strings.Contains(r.URL.Path, "/binary-"+arch+"/") {
			t.Errorf("wrong architecture index: %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
	})}
	_, _, err := CheckUpdate()
	if err == nil {
		t.Error("partial response must fail")
	}
	if !body.closed {
		t.Error("partial response body leaked")
	}
}
