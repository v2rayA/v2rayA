package iptables

import (
	"strings"
	"sync"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/common/cmds"
)

// http://briteming.hatenablog.com/entry/2019/06/18/175518

var mutex sync.Mutex

type SetupCommands string
type CleanCommands string

type iptablesSetter interface {
	GetSetupCommands() SetupCommands
	GetCleanCommands() CleanCommands
}

func (c SetupCommands) Setup(preprocess *func(c *SetupCommands)) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if preprocess != nil {
		(*preprocess)(&c)
	}
	commands := string(c)
	if common.IsInDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	return cmds.ExecCommands(commands, true)
}

func (c CleanCommands) Clean() {
	mutex.Lock()
	defer mutex.Unlock()
	commands := string(c)
	if common.IsInDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	_ = cmds.ExecCommands(commands, false)
}
