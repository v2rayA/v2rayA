package resolv

import (
	"os"
	"runtime"
	"strings"

	"github.com/v2rayA/v2rayA/conf"
)

const resolvConf = "/etc/resolv.conf"

func WriteResolvConf(servers []string) {
	if runtime.GOOS != "linux" {
		return
	}
	var sb strings.Builder
	for _, server := range servers {
		sb.WriteString("nameserver " + server + "\n")
	}
	os.WriteFile(resolvConf, []byte(sb.String()), 0644)
}

func CheckResolvConf() {
	if runtime.GOOS != "linux" {
		return
	}
	if conf.GetEnvironmentConfig().Lite {
		return
	}
	if _, err := os.Stat(resolvConf); os.IsNotExist(err) {
		WriteResolvConf([]string{"223.6.6.6"})
	}
}
