package cmds

import (
	"bytes"
	"os/exec"
)

func IsCommandValid(command string) bool {
	out, err := exec.Command("sh", "-c", "type '"+command+"'").CombinedOutput()
	if err != nil {
		out, err = exec.Command("sh", "-c", "which '"+command+"'").CombinedOutput()
	}
	return err == nil && len(bytes.TrimSpace(out)) > 0
}
