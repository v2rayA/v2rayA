//go:build linux

package httpClient

import (
	"testing"
	"time"
)

func TestDirectDialerMarksOnlyRequestedConnections(t *testing.T) {
	plain := DirectDialer(time.Second, false)
	if plain.Control != nil || plain.Resolver != nil {
		t.Fatal("plain dialer unexpectedly changes the socket or resolver")
	}
	marked := DirectDialer(2*time.Second, true)
	if marked.Control == nil || marked.Resolver == nil {
		t.Fatal("marked dialer must mark both destination and DNS sockets")
	}
	if marked.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", marked.Timeout)
	}
}
