package netstat

import (
	"testing"
)

func Test(t *testing.T) {
	t.Log(Print([]string{"tcp", "tcp6", "udp", "udp6"}))
}
