package cmds

import (
	"bytes"
	"os/exec"
	"strings"
)

func IsCommandValid(command string) bool {
	out, err := exec.Command("sh", "-c", "type '"+command+"'").CombinedOutput()
	if err != nil {
		out, err = exec.Command("sh", "-c", "command -v '"+command+"'").CombinedOutput()
	}
	return err == nil && len(bytes.TrimSpace(out)) > 0
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
