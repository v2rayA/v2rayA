package iptables

import (
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/cmds"
)

// http://briteming.hatenablog.com/entry/2019/06/18/175518

const watcherInterval = 3 * time.Second

var (
	watcher                 *LocalIPWatcher
	mutex                   sync.Mutex
	commandAvailable        = cmds.IsCommandValid
	executeCommands         = cmds.ExecCommands
	executeCommandWithInput = cmds.ExecCommandWithInput
)

type Setter struct {
	Cmds      string
	AfterFunc func() error
	PreFunc   func() error
	steps     []setterStep
}

func NewErrorSetter(err error) Setter {
	return Setter{PreFunc: func() error {
		return err
	}}
}

type proxySetter interface {
	GetSetupCommands() Setter
	GetCleanCommands() Setter
	AddIPWhitelist(cidr string)
	RemoveIPWhitelist(cidr string)
}

// watch interface changes and add specific IPs to whitelist on iptables
func SetWatcher(setter proxySetter) {
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

func (c Setter) Run(stopAtError bool) error {
	mutex.Lock()
	defer mutex.Unlock()
	return c.run(stopAtError, rewriteIptablesCommands)
}

func rewriteIptablesCommands(commands string) string {
	if common.IsDocker() {
		return rewriteIptablesBinaries(commands, "legacy")
	}
	if (!commandAvailable("iptables") || IsNftablesSupported()) &&
		commandAvailable("iptables-nft") {
		return rewriteIptablesBinaries(commands, "nft")
	}
	return commands
}

func rewriteIptablesBinaries(commands, variant string) string {
	commands = strings.ReplaceAll(commands, "iptables", "iptables-"+variant)
	return strings.ReplaceAll(commands, "ip6tables", "ip6tables-"+variant)
}

func (c Setter) run(stopAtError bool, rewrite func(string) string) error {
	var errs []error
	if c.PreFunc != nil {
		e := c.PreFunc()
		if e != nil {
			errs = append(errs, e)
			if stopAtError && len(errs) > 0 {
				return errs[0]
			}
		}
	}
	if len(c.steps) > 0 {
		for _, step := range c.steps {
			var e error
			if step.whitelist != nil {
				e = step.whitelist.run(stopAtError, rewrite)
			} else if len(step.commands) > 0 {
				e = executeCommands(rewrite(step.commands), stopAtError)
			}
			if e != nil {
				errs = append(errs, e)
				if stopAtError {
					return errs[0]
				}
			}
		}
	} else if len(c.Cmds) > 0 {
		e := executeCommands(rewrite(c.Cmds), stopAtError)
		if e != nil {
			errs = append(errs, e)
			if stopAtError {
				return errs[0]
			}
		}
	}
	if c.AfterFunc != nil {
		e := c.AfterFunc()
		if e != nil {
			errs = append(errs, e)
			if stopAtError && len(errs) > 0 {
				return errs[0]
			}
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
