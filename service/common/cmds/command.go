package cmds

import (
	"os/exec"
	"strings"
)

func IsCommandValid(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func ExecCommands(commands string, stopWhenError bool) error {
	lines := strings.Split(commands, "\n")
	var e error
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) <= 0 || strings.HasPrefix(line, "#") {
			continue
		}
		out, err := exec.Command("sh", "-c", line).CombinedOutput()
		if err != nil {
			e = newError(line, " ", string(out)).Base(err)
			if stopWhenError {
				return e
			}
		}
	}
	return e
}
