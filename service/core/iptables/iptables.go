package iptables

import (
	"V2RayA/tools"
	"V2RayA/tools/cmds"
	"strings"
)

// http://briteming.hatenablog.com/entry/2019/06/18/175518

type SetupCommands string
type CleanCommands string

type iptablesSetter interface {
	GetSetupCommands() SetupCommands
	GetCleanCommands() CleanCommands
}

func (c SetupCommands) Setup(preprocess *func(c *SetupCommands)) (err error) {
	if preprocess != nil {
		(*preprocess)(&c)
	}
	commands := string(c)
	if tools.IsInDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	return cmds.ExecCommands(commands, true)
}

func (c CleanCommands) Clean() {
	commands := string(c)
	if tools.IsInDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	_ = cmds.ExecCommands(commands, false)
}
