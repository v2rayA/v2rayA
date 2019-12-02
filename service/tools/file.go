package tools

import (
	"os"
	"time"
)

func GetFileModTime(path string) (t time.Time, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	fi, err := f.Stat()
	if err != nil {
		return
	}
	t = fi.ModTime()
	return
}
