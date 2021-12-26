package iptables

import (
	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
	"strings"
)

// GetNetworkServices 用于获取MacOS设备的 networkservices
func GetNetworkServices() []string {
	cmd := exec.Command("/usr/sbin/networksetup", "-listallnetworkservices")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("%v", err)
	}
	NetworkServices := strings.Split(string(stdoutStderr), "\n")
	return NetworkServices[1:]
}

// SetProxyOnMacOS 是应用于MacOS平台的自动配置代理
func SetProxyOnMacOS(flag bool, portHttpWithPac int32, portSocks5 int32) string {
	NetworkServices := GetNetworkServices()
	commands := ""
	switch {
	case flag == true:
		for _, service := range NetworkServices {
			commands += "/usr/sbin/networksetup" + "-setwebproxystate" + service + "on\n"
			commands += "/usr/sbin/networksetup" + "-setsocksfirewallproxystate" + service + "on\n"
			commands += "/usr/sbin/networksetup" + "-setwebproxy" + service + "127.0.0.1 " + string(portHttpWithPac) + "\n"
			commands += "/usr/sbin/networksetup" + "-setsocksfirewallproxy" + service + "127.0.0.1 " + string(portSocks5) + "\n"
		}
	case flag == false:
		for _, service := range NetworkServices {
			commands += "/usr/sbin/networksetup" + "-setwebproxystate" + service + "off\n"
			commands += "/usr/sbin/networksetup" + "-setsocksfirewallproxystate" + service + "off\n"
		}
	}
	return commands
}

// AutoSetProxyOnMac 用于配置代理模块的运行
func AutoSetProxyOnMac(flag bool, portHttpWithPac int32, portSocks5 int32) {
	commands := SetProxyOnMacOS(flag, portHttpWithPac, portSocks5)
	cmds.ExecCommands(commands, false)
}
