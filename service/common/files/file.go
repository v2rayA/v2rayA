package files

import (
	"os"
	"time"
)

func GetFileModTime(path string) (t time.Time, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	t = fi.ModTime()
	return
}
