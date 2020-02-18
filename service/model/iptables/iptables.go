package iptables

import (
	"V2RayA/global"
	"V2RayA/persistence/configure"
	"V2RayA/tools/cmds"
	"errors"
	"log"
	"strings"
)

// http://briteming.hatenablog.com/entry/2019/06/18/175518

var UseTproxy = false

type SetupCommands string
type CleanCommands string

type iptablesSetter interface {
	GetSetupCommands() SetupCommands
	GetCleanCommands() CleanCommands
}

func (c SetupCommands) Setup() (err error) {
	defer func() {
		if err != nil {
			DeleteRules()
		}
	}()
	commands := string(c)
	//先看要不要把自己的端口加进去
	selfPort := strings.Split(global.GetEnvironmentConfig().Address, ":")[1]
	wl := configure.GetPortWhiteListNotNil()
	if !wl.Has(selfPort, "tcp") {
		wl.TCP = append(wl.TCP, selfPort)
	}
	commands = strings.ReplaceAll(commands, "{{TCP_PORTS}}", strings.Join(wl.TCP, ","))
	if len(wl.UDP) > 0 {
		commands = strings.ReplaceAll(commands, "{{UDP_PORTS}}", strings.Join(wl.UDP, ","))
	} else { //没有UDP端口就把这行删了
		lines := strings.Split(commands, "\n")
		for i, line := range lines {
			if strings.Contains(line, "{{UDP_PORTS}}") {
				lines = append(lines[:i], lines[i+1:]...)
				break
			}
		}
		commands = strings.Join(lines, "\n")
	}
	if global.ServiceControlMode == global.DockerMode {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	return cmds.ExecCommands(commands, true)
}

func (c CleanCommands) Clean() {
	commands := string(c)
	if global.ServiceControlMode == global.DockerMode {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	_ = cmds.ExecCommands(commands, false)
}

func DeleteRules() {
	Tproxy.GetCleanCommands().Clean()
	Redirect.GetCleanCommands().Clean()
}

func WriteRules() error {
	if UseTproxy {
		if err := Tproxy.GetSetupCommands().Setup(); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = errors.New("内核未编译xt_TPROXY")
			}
			UseTproxy = false
		}
	}
	if !UseTproxy {
		if err := Redirect.GetSetupCommands().Setup(); err != nil {
			log.Println(err)
			return errors.New("机器不支持透明代理: " + err.Error())
		}
	}
	return nil
}
