//go:build darwin
// +build darwin

package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/db/configure"
)

type systemProxy struct{}

var SystemProxy systemProxy

// MacOSServiceSnapshot holds the saved proxy state for a single network service
type MacOSServiceSnapshot struct {
	WebEnabled       bool
	WebServer        string
	WebPort          int
	SecureWebEnabled bool
	SecureWebServer  string
	SecureWebPort    int
	SocksEnabled     bool
	SocksServer      string
	SocksPort        int
	AutoProxyEnabled bool
	AutoProxyURL     string
}

// macosProxyState stores the saved proxy state across all network services
type macosProxyState struct {
	mu       sync.Mutex
	saved    bool
	services map[string]*MacOSServiceSnapshot
}

var savedMacOSProxy macosProxyState

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

type MacOSProxySnapshot struct {
	Services map[string]*MacOSServiceSnapshot `json:"services"`
}

// GetNetworkServices 用于获取MacOS设备的 networkservices
func GetNetworkServices() ([]string, error) {
	cmd := exec.Command("/usr/sbin/networksetup", "-listallnetworkservices")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot get network services: %v", err)
	}
	lines := strings.Split(string(stdoutStderr), "\n")
	var services []string
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
		if len(lines[i]) == 0 || strings.Contains(lines[i], "*") {
			continue
		}
		services = append(services, lines[i])
	}
	return services, nil
}

func (p *systemProxy) AddIPWhitelist(cidr string) {}

func (p *systemProxy) RemoveIPWhitelist(cidr string) {}

// parseProxyOutput parses the output of networksetup -getwebproxy / -getsecurewebproxy / -getsocksfirewallproxy
func parseProxyOutput(output string) (enabled bool, server string, port int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Enabled:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Enabled:"))
			enabled = val == "Yes"
		} else if strings.HasPrefix(line, "Server:") {
			server = strings.TrimSpace(strings.TrimPrefix(line, "Server:"))
		} else if strings.HasPrefix(line, "Port:") {
			portStr := strings.TrimSpace(strings.TrimPrefix(line, "Port:"))
			port, _ = strconv.Atoi(portStr)
		}
	}
	return
}

// parseAutoProxyOutput parses the output of networksetup -getautoproxyurl
func parseAutoProxyOutput(output string) (enabled bool, url string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Enabled:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Enabled:"))
			enabled = val == "Yes"
		} else if strings.HasPrefix(line, "URL:") {
			url = strings.TrimSpace(strings.TrimPrefix(line, "URL:"))
		}
	}
	return
}

// readServiceState reads and returns the current proxy state for a given network service
func readServiceState(service string) (*MacOSServiceSnapshot, error) {
	state := &MacOSServiceSnapshot{}

	out, err := exec.Command("/usr/sbin/networksetup", "-getwebproxy", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getwebproxy: %v", err)
	}
	state.WebEnabled, state.WebServer, state.WebPort = parseProxyOutput(string(out))

	out, err = exec.Command("/usr/sbin/networksetup", "-getsecurewebproxy", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getsecurewebproxy: %v", err)
	}
	state.SecureWebEnabled, state.SecureWebServer, state.SecureWebPort = parseProxyOutput(string(out))

	out, err = exec.Command("/usr/sbin/networksetup", "-getsocksfirewallproxy", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getsocksfirewallproxy: %v", err)
	}
	state.SocksEnabled, state.SocksServer, state.SocksPort = parseProxyOutput(string(out))

	out, err = exec.Command("/usr/sbin/networksetup", "-getautoproxyurl", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getautoproxyurl: %v", err)
	}
	state.AutoProxyEnabled, state.AutoProxyURL = parseAutoProxyOutput(string(out))

	return state, nil
}

func (p *systemProxy) GetSetupCommands() Setter {
	networkServices, err := GetNetworkServices()
	if err != nil {
		return NewErrorSetter(err)
	}
	var commands string
	for _, service := range networkServices {
		commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %s on\n", shellQuote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %s on\n", shellQuote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %s on\n", shellQuote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxy %s 127.0.0.1 52345\n", shellQuote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxy %s 127.0.0.1 52345\n", shellQuote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxy %s 127.0.0.1 52306\n", shellQuote(service))
	}
	return Setter{
		PreFunc: func() error {
			savedMacOSProxy.mu.Lock()
			defer savedMacOSProxy.mu.Unlock()
			// a snapshot still pending from an earlier run is the user's
			// original; keep it and do not capture our own proxy as the state
			var previous MacOSProxySnapshot
			if found, err := configure.GetSystemProxySnapshot(&previous); err != nil {
				return err
			} else if found {
				log.Warn("system proxy: reusing the original state saved by an earlier run")
				savedMacOSProxy.services = previous.Services
				savedMacOSProxy.saved = true
				return nil
			} else if savedMacOSProxy.saved {
				return nil
			}

			services := make(map[string]*MacOSServiceSnapshot, len(networkServices))
			for _, service := range networkServices {
				state, err := readServiceState(service)
				if err != nil {
					return fmt.Errorf("failed to save proxy state for %v: %v", service, err)
				}
				services[service] = state
			}
			if err := configure.SetSystemProxySnapshot(&MacOSProxySnapshot{Services: services}); err != nil {
				return err
			}
			savedMacOSProxy.services = services
			savedMacOSProxy.saved = true
			return nil
		},
		Cmds: commands,
	}
}

func (p *systemProxy) GetCleanCommands() Setter {
	savedMacOSProxy.mu.Lock()
	saved := savedMacOSProxy.saved
	services := savedMacOSProxy.services
	savedMacOSProxy.mu.Unlock()
	if !saved {
		var snapshot MacOSProxySnapshot
		found, err := configure.GetSystemProxySnapshot(&snapshot)
		if err != nil {
			return NewErrorSetter(err)
		}
		if found {
			saved, services = true, snapshot.Services
		}
	}

	if !saved || services == nil {
		networkServices, err := GetNetworkServices()
		if err != nil {
			return NewErrorSetter(err)
		}
		// No saved state: fall back to simply turning everything off
		commands := ""
		for _, service := range networkServices {
			commands += fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %s off\n", shellQuote(service))
			commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %s off\n", shellQuote(service))
			commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %s off\n", shellQuote(service))
			commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %s off\n", shellQuote(service))
		}
		return Setter{Cmds: commands}
	}

	var commands strings.Builder
	for service, state := range services {

		// Restore auto proxy URL
		if state.AutoProxyEnabled && state.AutoProxyURL != "" {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxyurl %s %s\n", shellQuote(service), shellQuote(state.AutoProxyURL)))
		}
		if state.AutoProxyEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %s on\n", shellQuote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %s off\n", shellQuote(service)))
		}

		// Restore web proxy
		if state.WebEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxy %s %s %d\n", shellQuote(service), shellQuote(state.WebServer), state.WebPort))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %s on\n", shellQuote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %s off\n", shellQuote(service)))
		}

		// Restore secure web proxy
		if state.SecureWebEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxy %s %s %d\n", shellQuote(service), shellQuote(state.SecureWebServer), state.SecureWebPort))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %s on\n", shellQuote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %s off\n", shellQuote(service)))
		}

		// Restore SOCKS proxy
		if state.SocksEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxy %s %s %d\n", shellQuote(service), shellQuote(state.SocksServer), state.SocksPort))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %s on\n", shellQuote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %s off\n", shellQuote(service)))
		}
	}

	return Setter{Cmds: commands.String(), AfterFunc: func() error {
		savedMacOSProxy.mu.Lock()
		defer savedMacOSProxy.mu.Unlock()
		if err := configure.SetSystemProxySnapshot(nil); err != nil {
			return err
		}
		savedMacOSProxy.saved, savedMacOSProxy.services = false, nil
		return nil
	}}
}
