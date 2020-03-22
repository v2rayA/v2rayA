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
	ifnames := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(ifnames) == 1 && ifnames[0] == "" {
		ifnames = nil
	}
	return ifnames, nil
}
