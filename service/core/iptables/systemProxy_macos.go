//go:build darwin
// +build darwin

package iptables

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type systemProxy struct{}

var SystemProxy systemProxy

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
		commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxy %v 127.0.0.1 32345\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxy %v 127.0.0.1 32345\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxy %v 127.0.0.1 32346\n", strconv.Quote(service))
	}
	return Setter{
		Cmds: commands,
	}
}

func (p *systemProxy) GetCleanCommands() Setter {
	networkServices, err := GetNetworkServices()
	if err != nil {
		return NewErrorSetter(err)
	}

	commands := ""
	for _, service := range networkServices {
		commands += fmt.Sprintf("/usr/sbin/networksetup -setautoproxystate %v off\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v off\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v off\n", strconv.Quote(service))
		commands += fmt.Sprintf("/usr/sbin/networksetup -setsocksfirewallproxystate %v off\n", strconv.Quote(service))
	}
	return Setter{Cmds: commands}
}
