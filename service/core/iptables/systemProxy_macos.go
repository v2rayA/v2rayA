//go:build darwin
// +build darwin

package iptables

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type systemProxy struct{}

var SystemProxy systemProxy

// macosServiceState holds the saved proxy state for a single network service
type macosServiceState struct {
	webEnabled       bool
	webServer        string
	webPort          int
	secureWebEnabled bool
	secureWebServer  string
	secureWebPort    int
	socksEnabled     bool
	socksServer      string
	socksPort        int
	autoProxyEnabled bool
	autoProxyURL     string
}

// macosProxyState stores the saved proxy state across all network services
type macosProxyState struct {
	mu       sync.Mutex
	saved    bool
	services map[string]*macosServiceState
}

var savedMacOSProxy macosProxyState

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
func readServiceState(service string) (*macosServiceState, error) {
	state := &macosServiceState{}

	out, err := exec.Command("/usr/sbin/networksetup", "-getwebproxy", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getwebproxy: %v", err)
	}
	state.webEnabled, state.webServer, state.webPort = parseProxyOutput(string(out))

	out, err = exec.Command("/usr/sbin/networksetup", "-getsecurewebproxy", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getsecurewebproxy: %v", err)
	}
	state.secureWebEnabled, state.secureWebServer, state.secureWebPort = parseProxyOutput(string(out))

	out, err = exec.Command("/usr/sbin/networksetup", "-getsocksfirewallproxy", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getsocksfirewallproxy: %v", err)
	}
	state.socksEnabled, state.socksServer, state.socksPort = parseProxyOutput(string(out))

	out, err = exec.Command("/usr/sbin/networksetup", "-getautoproxyurl", service).Output()
	if err != nil {
		return nil, fmt.Errorf("getautoproxyurl: %v", err)
	}
	state.autoProxyEnabled, state.autoProxyURL = parseAutoProxyOutput(string(out))

	return state, nil
}

func (p *systemProxy) GetSetupCommands() Setter {
	networkServices, err := GetNetworkServices()
	if err != nil {
		return NewErrorSetter(err)
	}
	var commands string
	for _, service := range networkServices {
		commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v on\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v on\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %v on\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxy %v 127.0.0.1 52345\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxy %v 127.0.0.1 52345\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxy %v 127.0.0.1 52306\n", strconv.Quote(service))
	}
	return Setter{
		PreFunc: func() error {
			savedMacOSProxy.mu.Lock()
			defer savedMacOSProxy.mu.Unlock()

			services := make(map[string]*macosServiceState, len(networkServices))
			for _, service := range networkServices {
				state, err := readServiceState(service)
				if err != nil {
					return fmt.Errorf("failed to save proxy state for %v: %v", service, err)
				}
				services[service] = state
			}
			savedMacOSProxy.services = services
			savedMacOSProxy.saved = true
			return nil
		},
		Cmds: commands,
	}
}

func (p *systemProxy) GetCleanCommands() Setter {
	networkServices, err := GetNetworkServices()
	if err != nil {
		return NewErrorSetter(err)
	}

	savedMacOSProxy.mu.Lock()
	saved := savedMacOSProxy.saved
	services := savedMacOSProxy.services
	savedMacOSProxy.mu.Unlock()

	if !saved || services == nil {
		// No saved state: fall back to simply turning everything off
		commands := ""
		for _, service := range networkServices {
			commands += fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %v off\n", strconv.Quote(service))
			commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v off\n", strconv.Quote(service))
			commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v off\n", strconv.Quote(service))
			commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %v off\n", strconv.Quote(service))
		}
		return Setter{Cmds: commands}
	}

	var commands strings.Builder
	for _, service := range networkServices {
		state, ok := services[service]
		if !ok {
			// Service not found in saved state: turn off everything
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %v off\n", strconv.Quote(service)))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v off\n", strconv.Quote(service)))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v off\n", strconv.Quote(service)))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %v off\n", strconv.Quote(service)))
			continue
		}

		// Restore auto proxy URL
		if state.autoProxyEnabled && state.autoProxyURL != "" {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxyurl %v %v\n", strconv.Quote(service), strconv.Quote(state.autoProxyURL)))
		}
		if state.autoProxyEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %v on\n", strconv.Quote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %v off\n", strconv.Quote(service)))
		}

		// Restore web proxy
		if state.webEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxy %v %v %d\n", strconv.Quote(service), strconv.Quote(state.webServer), state.webPort))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v on\n", strconv.Quote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v off\n", strconv.Quote(service)))
		}

		// Restore secure web proxy
		if state.secureWebEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxy %v %v %d\n", strconv.Quote(service), strconv.Quote(state.secureWebServer), state.secureWebPort))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v on\n", strconv.Quote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v off\n", strconv.Quote(service)))
		}

		// Restore SOCKS proxy
		if state.socksEnabled {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxy %v %v %d\n", strconv.Quote(service), strconv.Quote(state.socksServer), state.socksPort))
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %v on\n", strconv.Quote(service)))
		} else {
			commands.WriteString(fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %v off\n", strconv.Quote(service)))
		}
	}

	savedMacOSProxy.mu.Lock()
	savedMacOSProxy.saved = false
	savedMacOSProxy.mu.Unlock()

	return Setter{Cmds: commands.String()}
}
