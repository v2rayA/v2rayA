package cmds

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecCommandWithInput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "stdin")
	err := ExecCommandWithInput("sh", []string{"-c", `IFS= read -r value; printf %s "$value" > "$1"`, "sh", output}, "payload\n")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Fatalf("stdin payload = %q, want payload", got)
	}
}
