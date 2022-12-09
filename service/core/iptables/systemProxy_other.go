//go:build !windows && !darwin
// +build !windows,!darwin

package iptables

import "fmt"

type systemProxy struct{}

var SystemProxy systemProxy

func (p *systemProxy) AddIPWhitelist(cidr string) {}

func (p *systemProxy) RemoveIPWhitelist(cidr string) {}

func (p *systemProxy) GetSetupCommands() Setter {
	return NewErrorSetter(fmt.Errorf("does not support to configure system proxy on your OS"))
}

func (p *systemProxy) GetCleanCommands() Setter {
	return Setter{}
}
