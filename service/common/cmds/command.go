package cmds

import (
	"fmt"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
	"strings"
)

func IsCommandValid(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func ExecCommandWithInput(command string, args []string, input string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ExecCommandWithInput: %s %s: %s: %w", command, strings.Join(args, " "), out, err)
	}
	return nil
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
			e = fmt.Errorf("ExecCommands: %v %v: %w", line, string(out), err)
			if stopWhenError {
				log.Trace("%v", e)
				return e
			}
		}
	}
	return e
}
