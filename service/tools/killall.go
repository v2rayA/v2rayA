package tools

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func KillAll(process string, immediately bool) (e error) {
	out, e := exec.Command("sh", "-c", "ps -e|grep v2ray|awk '{print $1}'").Output()
	if e != nil {
		return
	}
	pids := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, pid := range pids {
		cmd := fmt.Sprintf("kill %v", pid)
		if immediately {
			cmd += " -9"
		}
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			e = errors.New(err.Error() + string(out))
		}
	}
	return
}
