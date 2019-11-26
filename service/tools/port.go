package tools

import (
	"bytes"
	"fmt"
	"os/exec"
)

func IsPortOccupied(port string) (occupied bool, which string) {
	out, err := exec.Command("sh", "-c", fmt.Sprintf("netstat -tunlp|awk '{print $7,$4}'|grep %v$", port)).Output()
	occupied = err == nil && len(bytes.TrimSpace(out)) > 0
	which = string(bytes.SplitN(out, []byte(" "), 2)[0])
	return
}
