package httpClient

import (
	"errors"
	"net/http"
	"testing"
)

type timeoutTransport func(*http.Request) (*http.Response, error)

func (f timeoutTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSelectedClientFailureDoesNotFallBackToDirect(t *testing.T) {
	old := http.DefaultClient
	defer func() { http.DefaultClient = old }()
	fallbackCalled := false
	http.DefaultClient = &http.Client{Transport: timeoutTransport(func(r *http.Request) (*http.Response, error) {
		fallbackCalled = true
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})}
	proxyErr := errors.New("proxy unavailable")
	client := &http.Client{Transport: timeoutTransport(func(*http.Request) (*http.Response, error) { return nil, proxyErr })}
	_, err := HttpGetUsingSpecificClient(client, "http://example.invalid/")
	if err == nil || fallbackCalled || !errors.Is(err, proxyErr) {
		t.Fatalf("fallback: called=%v, err=%v", fallbackCalled, err)
	}
}
