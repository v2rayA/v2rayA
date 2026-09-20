//go:build darwin

package iptables

import (
	"strings"
	"testing"
)

func TestMacOSRestoreCommandsShellQuoteServiceName(t *testing.T) {
	const service = "Wi-Fi $(touch /tmp/x) 'guest'"
	savedMacOSProxy.mu.Lock()
	previousSaved, previousServices := savedMacOSProxy.saved, savedMacOSProxy.services
	savedMacOSProxy.saved = true
	savedMacOSProxy.services = map[string]*MacOSServiceSnapshot{service: {}}
	savedMacOSProxy.mu.Unlock()
	t.Cleanup(func() {
		savedMacOSProxy.mu.Lock()
		savedMacOSProxy.saved, savedMacOSProxy.services = previousSaved, previousServices
		savedMacOSProxy.mu.Unlock()
	})

	commands := SystemProxy.GetCleanCommands().Cmds
	want := shellQuote(service)
	if !strings.Contains(commands, want) {
		t.Fatalf("commands do not contain shell-quoted service %q: %s", want, commands)
	}
	if strings.Contains(commands, `"Wi-Fi $(touch /tmp/x)`) {
		t.Fatalf("commands leave substitution active: %s", commands)
	}
}
