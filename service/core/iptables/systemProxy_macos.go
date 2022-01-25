//go:build darwin
// +build darwin

package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
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
		if strings.Contains(lines[i], "*") {
			continue
		}
		services = append(services, lines[i])
	}
	return services, nil
}

func (p *systemProxy) AddIPWhitelist(cidr string) {}

func (p *systemProxy) RemoveIPWhitelist(cidr string) {}

func (p *systemProxy) GetSetupCommands() SetupCommands {
	networkServices, err := GetNetworkServices()
	if err != nil {
		log.Error("%v", err)
		return ""
	}
	var commands []string
	for _, service := range networkServices {
		commands = append(commands, fmt.Sprintf("/usr/sbin/networksetup -setwebproxystate %v on", service))
		commands = append(commands, fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxystate %v on", service))
		commands = append(commands, fmt.Sprintf("/usr/sbin/networksetup -setwebproxy %v 127.0.0.1 32345", service))
		commands = append(commands, fmt.Sprintf("/usr/sbin/networksetup -setsecurewebproxy %v 127.0.0.1 32345", service))
	}
	return SetupCommands(strings.Join(commands, "\n"))
}

func (p *systemProxy) GetCleanCommands() CleanCommands {
	networkServices, err := GetNetworkServices()
	if err != nil {
		log.Error("%v", err)
		return ""
	}

	commands := ""
	for _, service := range networkServices {
		commands += "/usr/sbin/networksetup" + "-setautoproxystate" + service + "off\n"
		commands += "/usr/sbin/networksetup" + "-setwebproxystate" + service + "off\n"
		commands += "/usr/sbin/networksetup" + "-setsecurewebproxystate" + service + "off\n"
		commands += "/usr/sbin/networksetup" + "-setsocksfirewallproxystate" + service + "off\n"
	}
	return CleanCommands(commands)
}
