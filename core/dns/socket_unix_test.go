//go:build !windows

package dns

import (
	"testing"
)

// On Unix, markedDialer() sets Dialer.Control so every socket created by the
// dialer carries SO_MARK=0x80 (the fwmark that makes iptables skip the DNS
// redirect rules). Without Control the upstream queries would be hijacked
// back into this module — an infinite loop and unbounded memory growth.
// Windows has no SO_MARK (markFd is a no-op there), so these assertions are
// platform-specific and live in a !windows file.
func TestMarkedDnsClientHasControl(t *testing.T) {
	udpClient := newMarkedDnsClient("udp")
	if udpClient.Dialer == nil {
		t.Fatal("Dialer is nil")
	}
	if udpClient.Dialer.Control == nil {
		t.Error("Dialer.Control is nil — UDP sockets would not carry SO_MARK=0x80 and queries could loop back into the DNS module")
	}

	tcpClient := newMarkedDnsClient("tcp")
	if tcpClient.Dialer == nil {
		t.Fatal("Dialer is nil")
	}
	if tcpClient.Dialer.Control == nil {
		t.Error("Dialer.Control is nil — TCP sockets would not carry SO_MARK=0x80")
	}
}
