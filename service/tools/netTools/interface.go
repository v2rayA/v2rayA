package netTools

import (
	"os/exec"
	"strings"
)

func GetDefaultInterface() (string, error) {
	b, err := exec.Command("sh", "-c", "awk '$2 == 00000000 { print $1 }' /proc/net/route").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
