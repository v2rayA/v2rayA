package log

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/v2rayA/beego/v2/logs"
)

func newMsg(text string) *logs.LogMsg {
	return &logs.LogMsg{Level: logs.LevelInfo, Msg: text, When: time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)}
}

func TestMemoryLogKeepsTailAndOffsets(t *testing.T) {
	w := &memoryWriter{max: 64}
	for i := 0; i < 5; i++ {
		_ = w.WriteMsg(newMsg(strings.Repeat("x", 10)))
	}
	if len(w.buf) > w.max {
		t.Fatalf("buffer holds %d bytes, cap is %d", len(w.buf), w.max)
	}
	if w.base == 0 {
		t.Fatal("nothing was evicted although the cap was exceeded")
	}
	if !bytes.HasSuffix(w.buf, []byte("\n")) || bytes.IndexByte(w.buf, '2') != 0 {
		t.Fatalf("eviction must cut on line boundaries, got %q", w.buf)
	}
	// A reader that is behind the oldest line gets the oldest line onward.
	old := memory
	memory = w
	defer func() { memory = old }()
	data, start := ReadMemory(0)
	if start != w.base || !bytes.Equal(data, w.buf) {
		t.Fatalf("read from 0: start=%d base=%d", start, w.base)
	}
	// A reader at the end gets nothing new.
	data, start = ReadMemory(w.base + int64(len(w.buf)))
	if len(data) != 0 {
		t.Fatalf("read at end returned %d bytes", len(data))
	}
	// A reader in the middle gets exactly the remainder.
	mid := w.base + 5
	data, start = ReadMemory(mid)
	if start != mid || !bytes.Equal(data, w.buf[5:]) {
		t.Fatalf("read from the middle returned start=%d len=%d", start, len(data))
	}
}

func TestMemoryLogLineShapeMatchesFile(t *testing.T) {
	w := &memoryWriter{max: 1 << 10}
	_ = w.WriteMsg(newMsg("hello"))
	line := string(w.buf)
	if !strings.HasPrefix(line, "2026/09/18 12:00:00.000 ") || !strings.Contains(line, "[I]") || !strings.HasSuffix(line, "hello\n") {
		t.Fatalf("unexpected line %q", line)
	}
}
