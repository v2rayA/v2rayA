package v2ray

import (
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
	"strings"
)

// TODO: 由于实在不清楚如何调用用户配置，所以暂时把端口写死，http 端口(带分流规则) 20172，sock5 端口 20170
// TODO: 适配自动配置 Windows 代理
// TODO: 在界面中自定义是否开启自动配置代理

// SetProxyOnMacOS 是应用于MacOS平台的自动配置代理
func SetProxyOnMacOS(flag bool) []string {
	NetworkServices := GetNetworkServices()
	var commands []string
	switch {
	case flag == true:
		for _, service := range NetworkServices {
			commands = append(commands, "/usr/sbin/networksetup"+"-setwebproxystate"+service+"on")
			commands = append(commands, "/usr/sbin/networksetup"+"-setsocksfirewallproxystate"+service+"on")
			commands = append(commands, "/usr/sbin/networksetup"+"-setwebproxy"+service+"127.0.0.1 20172")
			commands = append(commands, "/usr/sbin/networksetup"+"-setsocksfirewallproxy"+service+"127.0.0.1 20170")
		}
	case flag == false:
		for _, service := range NetworkServices {
			commands = append(commands, "/usr/sbin/networksetup"+"-setwebproxystate"+service+"off")
			commands = append(commands, "/usr/sbin/networksetup"+"-setsocksfirewallproxystate"+service+"off")
		}
	}
	return commands
}

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

// AutoSetProxy 用于配置代理模块的运行，其中 platform 预留了 Windows 的位置
func AutoSetProxy(flag bool, platform string) {
	switch {
	case platform == "MacOS":
		commands := SetProxyOnMacOS(flag)
		for _, command := range commands {
			cmd := exec.Command(command)
			err := cmd.Run()
			if err != nil {
				log.Error("Failed to call set proxy: %v", err)
			} else {
				log.Info("Proxy have already set")
			}
		}
	case platform == "Windows":
		//自动配置 Windows 代理
	}
}
