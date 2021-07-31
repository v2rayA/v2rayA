package resolv

import (
	"os"
	"strings"
)

const resolvConf = "/etc/resolv.conf"

func WriteResolvConf(servers []string) {
	var sb strings.Builder
	for _, server := range servers {
		sb.WriteString("nameserver " + server + "\n")
	}
	os.WriteFile(resolvConf, []byte(sb.String()), 0644)
}

func CheckResolvConf() {
	if _, err := os.Stat(resolvConf); os.IsNotExist(err) {
		WriteResolvConf([]string{"223.6.6.6"})
	}
}
