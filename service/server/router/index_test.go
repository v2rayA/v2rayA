package router

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPServerTimeouts(t *testing.T) {
	server := newHTTPServer(http.NotFoundHandler())
	if server.ReadHeaderTimeout != 10*time.Second || server.IdleTimeout != 120*time.Second {
		t.Fatalf("timeouts = read header %v, idle %v", server.ReadHeaderTimeout, server.IdleTimeout)
	}
	if server.ReadTimeout != 0 {
		t.Fatalf("read timeout = %v, want zero for streaming endpoints", server.ReadTimeout)
	}
}
