package httpClient

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

type timeoutTransport func(*http.Request) (*http.Response, error)

func (f timeoutTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFallbackPreservesClientTimeout(t *testing.T) {
	old := http.DefaultClient
	defer func() { http.DefaultClient = old }()
	fallbackCalled := false
	http.DefaultClient = &http.Client{Transport: timeoutTransport(func(r *http.Request) (*http.Response, error) {
		fallbackCalled = true
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("fallback request has no deadline")
			return nil, errors.New("unbounded fallback")
		}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	client := &http.Client{Timeout: 10 * time.Millisecond, Transport: timeoutTransport(func(*http.Request) (*http.Response, error) { return nil, errors.New("proxy unavailable") })}
	_, err := HttpGetUsingSpecificClient(client, "http://example.invalid/")
	if err == nil || !fallbackCalled {
		t.Fatalf("fallback: called=%v, err=%v", fallbackCalled, err)
	}
}
