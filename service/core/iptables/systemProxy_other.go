//go:build !windows && !darwin
// +build !windows,!darwin

package iptables

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type systemProxy struct{}

var SystemProxy systemProxy

func (p *systemProxy) AddIPWhitelist(cidr string) {}

func (p *systemProxy) RemoveIPWhitelist(cidr string) {}

// checkCommand checks if a command exists
func checkCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (p *systemProxy) GetSetupCommands() Setter {
	// Only support Linux in lite mode
	if runtime.GOOS != "linux" {
		return NewErrorSetter(fmt.Errorf("does not support to configure system proxy on your OS"))
	}

	if !conf.GetEnvironmentConfig().Lite {
		return NewErrorSetter(fmt.Errorf("system proxy is only supported in lite mode on Linux"))
	}

	var commands strings.Builder
	hasGsettings := false
	hasKDE := false

	// Try gsettings for GNOME-based desktops
	if checkCommand("gsettings") {
		hasGsettings = true
		commands.WriteString("gsettings set org.gnome.system.proxy mode 'manual'\n")
		commands.WriteString("gsettings set org.gnome.system.proxy.http host '127.0.0.1'\n")
		commands.WriteString("gsettings set org.gnome.system.proxy.http port 52345\n")
		commands.WriteString("gsettings set org.gnome.system.proxy.https host '127.0.0.1'\n")
		commands.WriteString("gsettings set org.gnome.system.proxy.https port 52345\n")
		commands.WriteString("gsettings set org.gnome.system.proxy.socks host '127.0.0.1'\n")
		commands.WriteString("gsettings set org.gnome.system.proxy.socks port 52306\n")
		log.Info("Using gsettings to configure system proxy (HTTP: 52345, SOCKS: 52306)")
	} else {
		log.Warn("gsettings command not found. GNOME-based applications may not use the system proxy. Please install gsettings if you are using GNOME desktop.")
	}

	// Try kwriteconfig6 for KDE Plasma 6
	if checkCommand("kwriteconfig6") {
		hasKDE = true
		commands.WriteString("kwriteconfig6 --file kioslaverc --group 'Proxy Settings' --key ProxyType 1\n")
		commands.WriteString("kwriteconfig6 --file kioslaverc --group 'Proxy Settings' --key httpProxy 'http://127.0.0.1:52345'\n")
		commands.WriteString("kwriteconfig6 --file kioslaverc --group 'Proxy Settings' --key httpsProxy 'http://127.0.0.1:52345'\n")
		commands.WriteString("kwriteconfig6 --file kioslaverc --group 'Proxy Settings' --key socksProxy 'socks://127.0.0.1:52306'\n")
		// Notify KDE about the change
		commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
		log.Info("Using kwriteconfig6 to configure system proxy (HTTP: 52345, SOCKS: 52306)")
	}

	// Try kwriteconfig5 for KDE Plasma 5
	if checkCommand("kwriteconfig5") {
		hasKDE = true
		commands.WriteString("kwriteconfig5 --file kioslaverc --group 'Proxy Settings' --key ProxyType 1\n")
		commands.WriteString("kwriteconfig5 --file kioslaverc --group 'Proxy Settings' --key httpProxy 'http://127.0.0.1:52345'\n")
		commands.WriteString("kwriteconfig5 --file kioslaverc --group 'Proxy Settings' --key httpsProxy 'http://127.0.0.1:52345'\n")
		commands.WriteString("kwriteconfig5 --file kioslaverc --group 'Proxy Settings' --key socksProxy 'socks://127.0.0.1:52306'\n")
		// Notify KDE about the change
		commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
		log.Info("Using kwriteconfig5 to configure system proxy (HTTP: 52345, SOCKS: 52306)")
	}

	if !hasKDE {
		log.Warn("kwriteconfig5/kwriteconfig6 commands not found. KDE applications may not use the system proxy. Please install kwriteconfig if you are using KDE Plasma desktop.")
	}

	if !hasGsettings && !hasKDE {
		return NewErrorSetter(fmt.Errorf("no supported desktop environment found. Please install gsettings (GNOME), kwriteconfig6 (KDE 6), or kwriteconfig5 (KDE 5)"))
	}

	return Setter{
		Cmds: commands.String(),
	}
}

func (p *systemProxy) GetCleanCommands() Setter {
	// Only support Linux in lite mode
	if runtime.GOOS != "linux" {
		return Setter{}
	}

	if !conf.GetEnvironmentConfig().Lite {
		return Setter{}
	}

	var commands strings.Builder

	// Try gsettings for GNOME-based desktops
	if checkCommand("gsettings") {
		commands.WriteString("gsettings set org.gnome.system.proxy mode 'none'\n")
		log.Info("Disabling system proxy via gsettings")
	}

	// Try kwriteconfig6 for KDE Plasma 6
	if checkCommand("kwriteconfig6") {
		commands.WriteString("kwriteconfig6 --file kioslaverc --group 'Proxy Settings' --key ProxyType 0\n")
		commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
		log.Info("Disabling system proxy via kwriteconfig6")
	}

	// Try kwriteconfig5 for KDE Plasma 5
	if checkCommand("kwriteconfig5") {
		commands.WriteString("kwriteconfig5 --file kioslaverc --group 'Proxy Settings' --key ProxyType 0\n")
		commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
		log.Info("Disabling system proxy via kwriteconfig5")
	}

	return Setter{
		Cmds: commands.String(),
	}
}
