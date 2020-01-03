package download

import (
	"errors"
	"fmt"
	"github.com/Code-Hex/pget"
	"path/filepath"
	"runtime"
)

func Pget(u string, p string) (err error) {
	//TODO: support proxy
	pg := pget.New()
	pg.URLs = []string{u}
	pg.TargetDir = filepath.Dir(p)
	pg.Utils.SetFileName(filepath.Base(p))
	pg.Procs = runtime.NumCPU()
	if err := pg.Checking(); err != nil {
		return errors.New(fmt.Sprintf("failed to check header: %s", err))
	}
	if err := pg.Download(); err != nil {
		return errors.New(fmt.Sprintf("failed to download: %s", err))
	}
	if err := pg.Utils.BindwithFiles(pg.Procs); err != nil {
		return errors.New(fmt.Sprintf("failed to download: %s", err))
	}
	return
}
