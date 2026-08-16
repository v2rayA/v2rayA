package dns

import (
	"testing"
	"time"
)

func TestNewMarkedDnsClient(t *testing.T) {
	client := newMarkedDnsClient("udp")
	if client.Net != "udp" {
		t.Errorf("Net = %q, want %q", client.Net, "udp")
	}
	if client.UDPSize != 4096 {
		t.Errorf("UDPSize = %d, want 4096", client.UDPSize)
	}
	if client.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", client.Timeout)
	}
	if client.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", client.ReadTimeout)
	}
	if client.WriteTimeout != 5*time.Second {
		t.Errorf("WriteTimeout = %v, want 5s", client.WriteTimeout)
	}
	if client.Dialer == nil {
		t.Fatal("Dialer is nil")
	}
}

func TestNewMarkedDnsClientTCP(t *testing.T) {
	client := newMarkedDnsClient("tcp")
	if client.Net != "tcp" {
		t.Errorf("Net = %q, want %q", client.Net, "tcp")
	}
	if client.Dialer == nil {
		t.Fatal("Dialer is nil")
	}
}
