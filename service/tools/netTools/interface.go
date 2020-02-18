package netTools

import (
	"os/exec"
	"strings"
)

func GetDefaultInterface() ([]string, error) {
	b, err := exec.Command("sh", "-c", "awk '$2 == 00000000 { print $1 }' /proc/net/route").Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n"), nil
}
