package v2ray

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Helpers shared by the TUN transparent proxy: the process exclusion list
// and the user's route scripts. They predate the in-core TUN and keep the
// TinyTun-era settings semantics.

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func parseCustomExcludeProcesses(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '\n', '\r', ';', '\t':
			return true
		default:
			return false
		}
	})
	custom := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if base := filepath.Base(p); base != "" && base != "." {
			p = base
		}
		custom = append(custom, p)
	}
	return dedupeStrings(custom)
}

// collectExcludeProcesses returns the process basenames that should be excluded
// from TUN proxying to prevent traffic loops.  It dynamically resolves
// v2rayA's own executable and the v2ray/xray core binary; names are only added
// when the path can actually be resolved, so no hardcoded strings are written.
func collectExcludeProcesses(customRaw string) []string {
	var processes []string

	// Exclude v2rayA itself so its own outgoing connections are not captured.
	if exe, err := os.Executable(); err == nil {
		if name := filepath.Base(exe); name != "" && name != "." {
			processes = append(processes, name)
		}
	} else {
		log.Warn("tun: failed to resolve own executable for process exclusion: %v", err)
	}

	// Exclude the v2ray/xray core so its proxy-server connections bypass TUN.
	if corePath, err := where.GetV2rayBinPath(); err == nil {
		if name := filepath.Base(corePath); name != "" && name != "." {
			processes = append(processes, name)
		}
	} else {
		log.Warn("tun: failed to resolve core binary for process exclusion: %v", err)
	}

	processes = append(processes, parseCustomExcludeProcesses(customRaw)...)

	return dedupeStrings(processes)
}

// shellInfo holds the binary path and argument format for a given shell type.
type shellInfo struct {
	// bin is the path/name of the shell binary.
	bin string
	// scriptFlag is the flag used to pass a script file to the shell.
	// For shells that accept a file path directly (bash, zsh, sh, fish), this is empty and the file is appended.
	// For PowerShell, this is "-File".
	// For cmd, this is "/C".
	scriptFlag string
	// ext is the file extension for the temporary script file.
	ext string
}

// resolveShellInfo maps a shell type string and optional custom path to a shellInfo.
func resolveShellInfo(shellType, shellPath string) (shellInfo, error) {
	// Custom shell: use user-provided path directly; assume POSIX-style (-c)
	if shellType == "custom" || (shellType == "" && shellPath != "") {
		if shellPath == "" {
			return shellInfo{}, fmt.Errorf("custom shell path is empty")
		}
		return shellInfo{bin: shellPath, ext: ".sh"}, nil
	}

	switch shellType {
	case "bash", "":
		bin := "/bin/bash"
		if runtime.GOOS == "windows" {
			bin = "bash.exe"
		}
		return shellInfo{bin: bin, ext: ".sh"}, nil
	case "zsh":
		return shellInfo{bin: "/bin/zsh", ext: ".sh"}, nil
	case "sh":
		return shellInfo{bin: "/bin/sh", ext: ".sh"}, nil
	case "fish":
		return shellInfo{bin: "fish", ext: ".fish"}, nil
	case "windows_powershell":
		return shellInfo{bin: "powershell.exe", scriptFlag: "-File", ext: ".ps1"}, nil
	case "pwsh":
		return shellInfo{bin: "pwsh.exe", scriptFlag: "-File", ext: ".ps1"}, nil
	case "cmd":
		return shellInfo{bin: "cmd.exe", scriptFlag: "/C", ext: ".bat"}, nil
	case "git_bash":
		return shellInfo{bin: "bash.exe", ext: ".sh"}, nil
	default:
		return shellInfo{}, fmt.Errorf("unknown shell type: %v", shellType)
	}
}

// runTunRouteScript writes the given script to a temp file and executes it
// with the configured shell.  stage is only used for logging.
func runTunRouteScript(stage, script string) error {
	if strings.TrimSpace(script) == "" {
		return nil
	}
	setting := configure.GetSettingNotNil()
	si, err := resolveShellInfo(setting.TunRouteShellType, setting.TunRouteShellPath)
	if err != nil {
		return fmt.Errorf("tun route script (%s): %w", stage, err)
	}

	// Resolve the shell binary via PATH if it is not an absolute path.
	binPath := si.bin
	if !filepath.IsAbs(binPath) {
		if resolved, err := exec.LookPath(binPath); err == nil {
			binPath = resolved
		}
	}

	// Write script to a temp file with restricted permissions.
	tmpFile, err := os.CreateTemp("", "tun_route_*"+si.ext)
	if err != nil {
		return fmt.Errorf("tun route script (%s): failed to create temp file: %w", stage, err)
	}
	defer os.Remove(tmpFile.Name())

	// Restrict permissions immediately after creation (before writing content).
	if runtime.GOOS != "windows" {
		if err = os.Chmod(tmpFile.Name(), 0600); err != nil {
			tmpFile.Close()
			return fmt.Errorf("tun route script (%s): failed to chmod temp file: %w", stage, err)
		}
	}

	if _, err = tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		return fmt.Errorf("tun route script (%s): failed to write temp file: %w", stage, err)
	}
	tmpFile.Close()

	// Make the file executable on non-Windows systems.
	if runtime.GOOS != "windows" {
		if err = os.Chmod(tmpFile.Name(), 0700); err != nil {
			return fmt.Errorf("tun route script (%s): failed to chmod temp file executable: %w", stage, err)
		}
	}

	// Build command arguments.
	var args []string
	if si.scriptFlag != "" {
		args = []string{binPath, si.scriptFlag, tmpFile.Name()}
	} else {
		args = []string{binPath, tmpFile.Name()}
	}

	log.Info("tun: running %s script via %s", stage, binPath)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("tun route script (%s) exited with error: %w", stage, err)
	}
	return nil
}

// The proxy nodes' own addresses must stay reachable outside the TUN for
// every process, not only the core: an ssh session to one's own VPS or a
// panel hosted on it would otherwise be sent through that very node. The
// TinyTun configuration carried them as skip_ips; the route installers
// take them from here.

func resolveHostToIPs(hostname string) ([]string, error) {
	if ip := net.ParseIP(hostname); ip != nil {
		return []string{ip.String()}, nil
	}
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %v: %w", hostname, err)
	}
	return addrs, nil
}
func isResolvableHost(host string) bool {
	if host == "" {
		return false
	}
	if net.ParseIP(host) != nil {
		return true
	}
	if strings.Contains(host, "/") {
		return false
	}
	for _, prefix := range []string{"geoip:", "ext:", "regexp:", "domain:", "keyword:", "full:", "geosite:"} {
		if strings.HasPrefix(host, prefix) {
			return false
		}
	}
	return true
}
func collectNodeIPs(tmpl *Template) []string {
	// First pass: gather all unique, resolvable hostnames.
	seenHost := make(map[string]struct{})
	var hostnames []string
	addHostname := func(h string) {
		if !isResolvableHost(h) {
			return
		}
		if _, ok := seenHost[h]; ok {
			return
		}
		seenHost[h] = struct{}{}
		hostnames = append(hostnames, h)
	}

	// Include live connections as well as the template snapshot.
	if css := configure.GetConnectedServers(); css != nil {
		for _, cs := range css.Get() {
			sr, err := cs.LocateServerRaw()
			if err != nil {
				log.Warn("tun: failed to locate server raw for the node bypass: %v", err)
				continue
			}
			addHostname(sr.ServerObj.GetHostname())
		}
	}

	// Source 2: serverInfoMap in the template (covers cases where the template
	// was constructed from a snapshot that may differ from the live DB state).
	for _, info := range tmpl.serverInfoMap {
		addHostname(info.Info.GetHostname())
	}

	if len(hostnames) == 0 {
		return nil
	}

	// Second pass: resolve all hostnames concurrently.
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		seenIP = make(map[string]struct{})
		result []string
	)
	for _, hostname := range hostnames {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			ips, err := resolveHostToIPs(host)
			if err != nil {
				log.Warn("tun: failed to resolve node hostname %v: %v", host, err)
				return
			}
			mu.Lock()
			for _, ip := range ips {
				if _, ok := seenIP[ip]; ok {
					continue
				}
				seenIP[ip] = struct{}{}
				result = append(result, ip)
			}
			mu.Unlock()
		}(hostname)
	}
	wg.Wait()
	return result
}
