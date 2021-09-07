package iptables

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/cmds"
	"strings"
	"sync"
	"time"
)

// http://briteming.hatenablog.com/entry/2019/06/18/175518

const watcherInterval = 3 * time.Second

var watcher *LocalIPWatcher
var mutex sync.Mutex

type SetupCommands string
type CleanCommands string

type iptablesSetter interface {
	GetSetupCommands() SetupCommands
	GetCleanCommands() CleanCommands
	AddIPWhitelist(cidr string)
	RemoveIPWhitelist(cidr string)
}

// watch interface changes and add specific IPs to whitelist on iptables
func SetWatcher(setter iptablesSetter) {
	if watcher != nil {
		watcher.Close()
	}
	watcher = NewLocalIPWatcher(watcherInterval, setter.AddIPWhitelist, setter.RemoveIPWhitelist)
}

func CloseWatcher() {
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
}

func (c SetupCommands) Setup(preprocess func(c *SetupCommands)) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if preprocess != nil {
		preprocess(&c)
	}
	commands := string(c)
	if common.IsDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	err = cmds.ExecCommands(commands, true)
	if err != nil {
		return
	}
	return
}

func (c CleanCommands) Clean() {
	mutex.Lock()
	defer mutex.Unlock()
	commands := string(c)
	if common.IsDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
	_ = cmds.ExecCommands(commands, false)
}
