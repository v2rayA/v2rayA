//go:build !windows && !darwin
// +build !windows,!darwin

package iptables

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type systemProxy struct{}

var SystemProxy systemProxy

// linuxGsettingsState holds saved GNOME/gsettings proxy values
type linuxGsettingsState struct {
	mode      string
	httpHost  string
	httpPort  string
	httpsHost string
	httpsPort string
	socksHost string
	socksPort string
}

// linuxKDEState holds saved KDE proxy values
type linuxKDEState struct {
	proxyType  string
	httpProxy  string
	httpsProxy string
	socksProxy string
}

// linuxProxySavedState stores original proxy state before v2rayA modifies it
type linuxProxySavedState struct {
	mu        sync.Mutex
	saved     bool
	gsettings linuxGsettingsState
	kde       linuxKDEState
	snapshot  *LinuxProxySnapshot
}

type LinuxProxySnapshot struct {
	GNOMECaptured bool              `json:"gnomeCaptured"`
	KDECaptured   bool              `json:"kdeCaptured"`
	GSettings     map[string]string `json:"gsettings"`
	KDE           map[string]string `json:"kde"`
}

var savedLinuxProxy linuxProxySavedState

func (p *systemProxy) AddIPWhitelist(cidr string) {}

func (p *systemProxy) RemoveIPWhitelist(cidr string) {}

// checkCommand checks if a command exists
func checkCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// readGsettingsValue runs "gsettings get <schema> <key>" and returns the trimmed output
func readGsettingsValue(schema, key string) string {
	out, err := exec.Command("gsettings", "get", schema, key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// readKDEConfigValue reads a KDE config value using kreadconfig5/6 or grep fallback
func readKDEConfigValue(configFile, group, key string) (string, error) {
	// Try kreadconfig6 first, then kreadconfig5
	for _, cmd := range []string{"kreadconfig6", "kreadconfig5"} {
		if checkCommand(cmd) {
			out, err := exec.Command(cmd, "--file", configFile, "--group", group, "--key", key).Output()
			if err == nil {
				return strings.TrimSpace(string(out)), nil
			}
			return "", fmt.Errorf("capture KDE proxy %s: %w", key, err)
		}
	}
	return "", fmt.Errorf("cannot capture KDE proxy without kreadconfig")
}

// saveGsettingsState saves current GNOME proxy settings
func saveGsettingsState(state *linuxGsettingsState) {
	if !checkCommand("gsettings") {
		return
	}
	state.mode = readGsettingsValue("org.gnome.system.proxy", "mode")
	state.httpHost = readGsettingsValue("org.gnome.system.proxy.http", "host")
	state.httpPort = readGsettingsValue("org.gnome.system.proxy.http", "port")
	state.httpsHost = readGsettingsValue("org.gnome.system.proxy.https", "host")
	state.httpsPort = readGsettingsValue("org.gnome.system.proxy.https", "port")
	state.socksHost = readGsettingsValue("org.gnome.system.proxy.socks", "host")
	state.socksPort = readGsettingsValue("org.gnome.system.proxy.socks", "port")
}

// saveKDEState saves current KDE proxy settings
func saveKDEState(state *linuxKDEState) error {
	configFile := "kioslaverc"
	group := "Proxy Settings"
	for _, field := range []struct {
		key   string
		value *string
	}{
		{"ProxyType", &state.proxyType}, {"httpProxy", &state.httpProxy},
		{"httpsProxy", &state.httpsProxy}, {"socksProxy", &state.socksProxy},
	} {
		value, err := readKDEConfigValue(configFile, group, field.key)
		if err != nil {
			return err
		}
		*field.value = value
	}
	return nil
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
		PreFunc: func() error {
			savedLinuxProxy.mu.Lock()
			defer savedLinuxProxy.mu.Unlock()
			// A snapshot still pending from an earlier run is the user's
			// original: it is kept as the state to go back to, and the
			// capture is not repeated (the registry already holds our proxy).
			var previous LinuxProxySnapshot
			if found, err := configure.GetSystemProxySnapshot(&previous); err != nil {
				return err
			} else if found {
				log.Warn("system proxy: reusing the original state saved by an earlier run")
				savedLinuxProxy.snapshot = &previous
				savedLinuxProxy.saved = true
				return nil
			} else if savedLinuxProxy.saved {
				return nil
			}

			if hasGsettings {
				saveGsettingsState(&savedLinuxProxy.gsettings)
			}
			if hasKDE {
				if err := saveKDEState(&savedLinuxProxy.kde); err != nil {
					return err
				}
			}
			gs, kde := savedLinuxProxy.gsettings, savedLinuxProxy.kde
			snapshot := &LinuxProxySnapshot{
				GNOMECaptured: hasGsettings, KDECaptured: hasKDE,
				GSettings: map[string]string{"mode": gs.mode, "http.host": gs.httpHost, "http.port": gs.httpPort, "https.host": gs.httpsHost, "https.port": gs.httpsPort, "socks.host": gs.socksHost, "socks.port": gs.socksPort},
				KDE:       map[string]string{"ProxyType": kde.proxyType, "httpProxy": kde.httpProxy, "httpsProxy": kde.httpsProxy, "socksProxy": kde.socksProxy},
			}
			if hasGsettings {
				for key, value := range snapshot.GSettings {
					if value == "" {
						return fmt.Errorf("could not capture GNOME proxy %s", key)
					}
				}
			}
			if err := configure.SetSystemProxySnapshot(snapshot); err != nil {
				return err
			}
			savedLinuxProxy.snapshot = snapshot
			savedLinuxProxy.saved = true
			return nil
		},
		Cmds: commands.String(),
	}
}

func (p *systemProxy) GetCleanCommands() Setter {
	// Only support Linux in lite mode
	if runtime.GOOS != "linux" {
		return Setter{}
	}

	savedLinuxProxy.mu.Lock()
	snapshot := savedLinuxProxy.snapshot
	savedLinuxProxy.mu.Unlock()
	if snapshot == nil {
		var stored LinuxProxySnapshot
		found, err := configure.GetSystemProxySnapshot(&stored)
		if err != nil {
			return NewErrorSetter(err)
		}
		if found {
			snapshot = &stored
		}
	}
	if snapshot == nil {
		if !conf.GetEnvironmentConfig().Lite {
			return Setter{}
		}
		// No saved state: fall back to original behavior (disable proxy)
		var commands strings.Builder

		if checkCommand("gsettings") {
			commands.WriteString("gsettings set org.gnome.system.proxy mode 'none'\n")
			log.Info("Disabling system proxy via gsettings")
		}

		if checkCommand("kwriteconfig6") {
			commands.WriteString("kwriteconfig6 --file kioslaverc --group 'Proxy Settings' --key ProxyType 0\n")
			commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
			log.Info("Disabling system proxy via kwriteconfig6")
		}

		if checkCommand("kwriteconfig5") {
			commands.WriteString("kwriteconfig5 --file kioslaverc --group 'Proxy Settings' --key ProxyType 0\n")
			commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
			log.Info("Disabling system proxy via kwriteconfig5")
		}

		return Setter{
			Cmds: commands.String(),
		}
	}

	var commands strings.Builder
	if snapshot.GNOMECaptured {
		if !checkCommand("gsettings") {
			return NewErrorSetter(fmt.Errorf("gsettings is required to restore captured GNOME proxy"))
		}
		for _, key := range []string{"mode", "http.host", "http.port", "https.host", "https.port", "socks.host", "socks.port"} {
			schema, field := "org.gnome.system.proxy", key
			if backend, suffix, ok := strings.Cut(key, "."); ok {
				schema += "." + backend
				field = suffix
			}
			value := strings.ReplaceAll(snapshot.GSettings[key], "'", "'\\''")
			fmt.Fprintf(&commands, "gsettings set %s %s '%s'\n", schema, field, value)
		}
	}
	if snapshot.KDECaptured {
		cmd := "kwriteconfig6"
		if !checkCommand(cmd) {
			cmd = "kwriteconfig5"
		}
		if !checkCommand(cmd) {
			return NewErrorSetter(fmt.Errorf("kwriteconfig is required to restore captured KDE proxy"))
		}
		for _, key := range []string{"ProxyType", "httpProxy", "httpsProxy", "socksProxy"} {
			value := strings.ReplaceAll(snapshot.KDE[key], "'", "'\\''")
			fmt.Fprintf(&commands, "%s --file kioslaverc --group 'Proxy Settings' --key %s '%s'\n", cmd, key, value)
		}
		commands.WriteString("dbus-send --type=signal /KIO/Scheduler org.kde.KIO.Scheduler.reparseSlaveConfiguration string:''\n")
	}

	return Setter{
		Cmds: commands.String(),
		AfterFunc: func() error {
			savedLinuxProxy.mu.Lock()
			defer savedLinuxProxy.mu.Unlock()
			if err := configure.SetSystemProxySnapshot(nil); err != nil {
				return err
			}
			savedLinuxProxy.saved, savedLinuxProxy.snapshot = false, nil
			savedLinuxProxy.gsettings, savedLinuxProxy.kde = linuxGsettingsState{}, linuxKDEState{}
			return nil
		},
	}
}
