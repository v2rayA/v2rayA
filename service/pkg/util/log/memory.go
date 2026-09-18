package log

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/v2rayA/beego/v2/logs"
)

// memoryWriter keeps the most recent log output in memory so the web log
// viewer has something to show when v2rayA runs without --log-file — the
// Windows service and every "log to console" setup. It used to answer
// "log printed to console, please see log in console", which a service user
// cannot do.
type memoryWriter struct {
	mu   sync.Mutex
	buf  []byte
	max  int
	base int64 // absolute offset of buf[0] in the stream of everything written
}

const memoryLogCapacity = 4 << 20 // 4 MiB, tens of thousands of lines

var memory = &memoryWriter{max: memoryLogCapacity}

func init() {
	logs.Register("memory", func() logs.Logger { return memory })
}

func (w *memoryWriter) Init(string) error              { return nil }
func (w *memoryWriter) Destroy()                       {}
func (w *memoryWriter) Flush()                         {}
func (w *memoryWriter) SetFormatter(logs.LogFormatter) {}

// WriteMsg renders the line the way the file adapter does, so the viewer
// sees the same text whichever source it reads.
func (w *memoryWriter) WriteMsg(lm *logs.LogMsg) error {
	line := fmt.Sprintf("%s %s\n", lm.When.Format("2006/01/02 15:04:05.000"), lm.OldStyleFormat())
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, line...)
	if len(w.buf) > w.max {
		// Drop whole lines from the front until the buffer fits again.
		cut := len(w.buf) - w.max
		if nl := bytes.IndexByte(w.buf[cut:], '\n'); nl >= 0 {
			cut += nl + 1
		}
		w.base += int64(cut)
		w.buf = append([]byte(nil), w.buf[cut:]...)
	}
	return nil
}

// ReadMemory returns the buffered log from absolute offset skip, or from the
// oldest line still held when skip is older than that, together with the
// offset the returned data starts at.
func ReadMemory(skip int64) (data []byte, start int64) {
	memory.mu.Lock()
	defer memory.mu.Unlock()
	end := memory.base + int64(len(memory.buf))
	switch {
	case skip < memory.base:
		start = memory.base
	case skip > end:
		start = end
	default:
		start = skip
	}
	return append([]byte(nil), memory.buf[start-memory.base:]...), start
}
