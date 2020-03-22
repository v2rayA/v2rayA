package process

import (
	"errors"
	"os/exec"
	"strings"
)

func KillAll(process string, immediately bool) (e error) {
	out, err := exec.Command("sh", "-c", "ps -e -o pid -o comm|grep "+process+"$|awk '{print $1}'").CombinedOutput()
	if err != nil || strings.Contains(string(out), "invalid option") {
		if strings.Contains(strings.ToLower(string(out)), "busybox") {
			out, e = exec.Command("sh", "-c", "ps|awk '{print $1,$5}'|grep "+process+"$|awk '{print $1}'").Output()
		} else {
			return
		}
	}
	pids := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, pid := range pids {
		if len(pid) <= 0 {
			continue
		}
		cmd := "kill "
		if immediately {
			cmd += "-9 "
		}
		cmd += pid
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			e = errors.New(err.Error() + string(out))
		}
	}
	return
}
