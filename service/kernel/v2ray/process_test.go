package v2ray

import "testing"

func TestProcessCloseMarksExpectedStopBeforeCancel(t *testing.T) {
	process := &Process{template: &Template{}}
	process.procCancel = func() {
		if !process.expectedStop.Load() {
			t.Fatal("expected stop must be recorded before canceling the core")
		}
	}

	if err := process.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !process.expectedStop.Load() {
		t.Fatal("Close() did not record the expected stop")
	}
}

func TestLogInfoWriterEmptyWrite(t *testing.T) {
	n, err := (logInfoWriter{}).Write(nil)
	if n != 0 || err != nil {
		t.Fatalf("empty write: n=%d, err=%v", n, err)
	}
}
