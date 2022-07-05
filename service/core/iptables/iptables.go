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

type Setter struct {
	Cmds      string
	AfterFunc func() error
	PreFunc   func() error
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
	commands := c.Cmds
	if common.IsDocker() {
		commands = strings.ReplaceAll(commands, "iptables", "iptables-legacy")
		commands = strings.ReplaceAll(commands, "ip6tables", "ip6tables-legacy")
	}
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
	if len(commands) > 0 {
		e := cmds.ExecCommands(commands, stopAtError)
		if e != nil {
			errs = append(errs, e)
			if stopAtError && len(errs) > 0 {
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
