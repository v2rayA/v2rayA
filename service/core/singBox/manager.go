package singBox

import (
	"github.com/v2rayA/v2rayA/db/configure"
)

var singleProcess *Process

func SetupTunnel(setting *configure.Setting) (err error) {
	CleanTunnel()
	t := NewTunTemplate(setting)
	singleProcess, err = NewProcess(t, func() error { return nil }, func() error { return nil }, func(p *Process) {})
	return
}

func CleanTunnel() (err error) {
	if singleProcess != nil {
		err = singleProcess.Close()
		singleProcess = nil
	}
	return
}
