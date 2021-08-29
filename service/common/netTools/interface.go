package netTools

import (
	"fmt"
	"os/exec"
	"strings"
)

var NoDefaultInterface = fmt.Errorf("default interfaces not found")

// only for linux
func GetDefaultInterfaceName() ([]string, error) {
	b, err := exec.Command("sh", "-c", "awk '$2 == 00000000 { print $1 }' /proc/net/route").Output()
	if err != nil {
		return nil, err
	}
	ifnames := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(ifnames) == 1 && ifnames[0] == "" {
		return nil, NoDefaultInterface
	}
	return ifnames, nil
}
