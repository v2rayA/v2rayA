//go:build tinytun

package v2ray

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// tinytunLogConf represents the log settings in TinyTun config.
type tinytunLogConf struct {
	Loglevel string `json:"loglevel"`
}

// tinytunTunConf represents the TUN interface settings in TinyTun config.
// Note: auto_route is intentionally omitted here; it is passed as --auto-route CLI flag instead.
type tinytunTunConf struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Netmask  string `json:"netmask"`
	Ipv6Mode string `json:"ipv6_mode,omitempty"`
	MTU      int    `json:"mtu,omitempty"`
}

// tinytunSocks5Conf represents the SOCKS5 proxy settings in TinyTun config.
type tinytunSocks5Conf struct {
	Address      string  `json:"address"`
	Username     *string `json:"username"`
	Password     *string `json:"password"`
	DnsOverSocks bool    `json:"dns_over_socks5"`
}

// tinytunDnsServerConf represents a single DNS server entry in TinyTun config.
type tinytunDnsServerConf struct {
	Address string `json:"address"`
	Route   string `json:"route"`
}

// tinytunDnsConf represents the DNS settings in TinyTun config.
type tinytunDnsConf struct {
	Servers    []tinytunDnsServerConf `json:"servers"`
	ListenPort int                    `json:"listen_port"`
	TimeoutMs  int                    `json:"timeout_ms"`
}

// tinytunFilteringConf represents the filtering settings in TinyTun config.
type tinytunFilteringConf struct {
	SkipIPs          []string `json:"skip_ips"`
	SkipNetworks     []string `json:"skip_networks"`
	BlockPorts       []int    `json:"block_ports"`
	AllowPorts       []int    `json:"allow_ports"`
	ExcludeProcesses []string `json:"exclude_processes"`
}

// tinytunRouteConf represents the route settings in TinyTun config.
type tinytunRouteConf struct {
	AutoDetectInterface bool    `json:"auto_detect_interface"`
	DefaultInterface    *string `json:"default_interface"`
}

// tinytunConfig is the top-level TinyTun JSON configuration.
type tinytunConfig struct {
	Log       tinytunLogConf       `json:"log"`
	Tun       tinytunTunConf       `json:"tun"`
	Socks5    tinytunSocks5Conf    `json:"socks5"`
	DNS       tinytunDnsConf       `json:"dns"`
	Filtering tinytunFilteringConf `json:"filtering"`
	Route     tinytunRouteConf     `json:"route"`
}

const (
	tinytunBinName        = "tinytun"
	tinytunConfigFileName = "tinytun.json"
	// tinytunSocksPort is the SOCKS5 port in v2ray dedicated for TinyTun traffic.
	// This matches the "transparent" inbound added in setInbound for TransparentTun.
	tinytunSocksPort = 52345
)

// tinyTunState tracks the running TinyTun process.
var tinyTunState struct {
	cancel context.CancelFunc
	mu     sync.Mutex
}

// GetTinyTunBinPath returns the path to the TinyTun binary.
// It first checks the --tinytun-bin / V2RAYA_TINYTUN_BIN configuration,
// then searches PATH and the current working directory.
func GetTinyTunBinPath() (string, error) {
	if binPath := conf.GetEnvironmentConfig().TinyTunBin; binPath != "" {
		return binPath, nil
	}
	return getTinyTunBinPathAuto()
}

func getTinyTunBinPathAuto() (string, error) {
	target := tinytunBinName
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(target), ".exe") {
		target += ".exe"
	}
	// Search in PATH
	if path, err := exec.LookPath(target); err == nil {
		return path, nil
	}
	// Search in current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("tinytun binary not found: please install tinytun or use --tinytun-bin")
	}
	path := filepath.Join(pwd, target)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("tinytun binary not found: please install tinytun or use --tinytun-bin to specify its path")
}

// resolveHostToIPs resolves a hostname to a list of IP strings.
// If the input is already an IP address it is returned as-is.
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

// collectNodeIPs returns the deduplicated list of IP addresses for all proxy
// nodes referenced by tmpl. Domain-based node addresses are resolved to IPs.
func collectNodeIPs(tmpl *Template) []string {
	seen := make(map[string]struct{})
	var result []string
	appendIPs := func(hostname string) {
		if !isResolvableHost(hostname) {
			return
		}
		ips, err := resolveHostToIPs(hostname)
		if err != nil {
			log.Warn("tinytun: failed to resolve node hostname %v: %v", hostname, err)
			return
		}
		for _, ip := range ips {
			if _, ok := seen[ip]; ok {
				continue
			}
			seen[ip] = struct{}{}
			result = append(result, ip)
		}
	}

	// Source 1: read directly from the connected-server database.
	// This is the most reliable source because it does not depend on whether
	// serverInfoMap was populated (e.g. balancer paths skip serverInfoMap).
	if css := configure.GetConnectedServers(); css != nil {
		for _, cs := range css.Get() {
			sr, err := cs.LocateServerRaw()
			if err != nil {
				log.Warn("tinytun: failed to locate server raw for skip_ips: %v", err)
				continue
			}
			appendIPs(sr.ServerObj.GetHostname())
		}
	}

	// Source 2: serverInfoMap in the template (covers cases where the template
	// was constructed from a snapshot that may differ from the live DB state).
	for _, info := range tmpl.serverInfoMap {
		appendIPs(info.Info.GetHostname())
	}

	return result
}

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

// collectBypassInterfaceNetworks returns the IP network CIDRs of all active
// interfaces whose names match any of the comma-separated glob patterns in
// the patterns string.  It is used to populate skip_networks in the TinyTun
// config so that traffic to these subnets bypasses the TUN proxy.
func collectBypassInterfaceNetworks(patterns string) []string {
	if patterns == "" {
		return nil
	}
	var patternList []string
	for _, p := range strings.Split(patterns, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			patternList = append(patternList, p)
		}
	}
	if len(patternList) == 0 {
		return nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Warn("tinytun: failed to enumerate local interfaces: %v", err)
		return nil
	}

	var networks []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		matched := false
		for _, pattern := range patternList {
			if ok, _ := filepath.Match(pattern, iface.Name); ok {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Warn("tinytun: failed to list addresses for interface %v: %v", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP == nil || ipnet.Mask == nil {
				continue
			}
			networkIP := ipnet.IP.Mask(ipnet.Mask)
			if networkIP == nil {
				continue
			}
			networks = append(networks, (&net.IPNet{IP: networkIP, Mask: ipnet.Mask}).String())
		}
	}
	return dedupeStrings(networks)
}

// collectExcludeProcesses returns the process basenames that should be excluded
// from TinyTun proxying to prevent traffic loops.  It dynamically resolves
// v2rayA's own executable and the v2ray/xray core binary; names are only added
// when the path can actually be resolved, so no hardcoded strings are written.
func collectExcludeProcesses() []string {
	var processes []string

	// Exclude v2rayA itself so its own outgoing connections are not captured.
	if exe, err := os.Executable(); err == nil {
		if name := filepath.Base(exe); name != "" && name != "." {
			processes = append(processes, name)
		}
	} else {
		log.Warn("tinytun: failed to resolve own executable for process exclusion: %v", err)
	}

	// Exclude the v2ray/xray core so its proxy-server connections bypass TUN.
	if corePath, err := where.GetV2rayBinPath(); err == nil {
		if name := filepath.Base(corePath); name != "" && name != "." {
			processes = append(processes, name)
		}
	} else {
		log.Warn("tinytun: failed to resolve core binary for process exclusion: %v", err)
	}

	return dedupeStrings(processes)
}

// generateTinyTunConfig generates a TinyTun JSON config file and returns its path.
func generateTinyTunConfig(tmpl *Template) (string, error) {
	setting := configure.GetSettingNotNil()
	skipIPs := dedupeStrings(append([]string{"127.0.0.1", "198.18.0.1"}, collectNodeIPs(tmpl)...))
	dnsServers := []tinytunDnsServerConf{
		// Forward all DNS from TinyTun to the v2fly dns-in-tun dokodemo-door (127.0.0.1:6053).
		// v2fly will apply its own DNS routing rules (direct / proxy) and return answers.
		// Route is "direct" so TinyTun sends these packets straight to the local loopback
		// without going through the SOCKS5 proxy.
		{Address: "127.0.0.1:6053", Route: "direct"},
	}
	skipNetworks := dedupeStrings(collectBypassInterfaceNetworks(setting.TunBypassInterfaces))

	cfg := tinytunConfig{
		Log: tinytunLogConf{
			Loglevel: setting.LogLevel,
		},
		Tun: tinytunTunConf{
			Name:    "tun0",
			IP:      "198.18.0.1",
			Netmask: "255.255.255.255",
			MTU:     1500,
		},
		Socks5: tinytunSocks5Conf{
			Address:      fmt.Sprintf("127.0.0.1:%d", tinytunSocksPort),
			DnsOverSocks: true,
		},
		DNS: tinytunDnsConf{
			Servers:    dnsServers,
			ListenPort: 53,
			TimeoutMs:  5000,
		},
		Filtering: tinytunFilteringConf{
			SkipIPs:          skipIPs,
			SkipNetworks:     skipNetworks,
			BlockPorts:       []int{22, 23, 25, 110, 143},
			AllowPorts:       []int{80, 443, 53},
			ExcludeProcesses: collectExcludeProcesses(),
		},
		Route: tinytunRouteConf{
			AutoDetectInterface: true,
		},
	}

	data, err := jsoniter.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal tinytun config: %w", err)
	}

	configPath := filepath.Join(conf.GetEnvironmentConfig().Config, tinytunConfigFileName)
	if err = os.WriteFile(configPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write tinytun config to %v: %w", configPath, err)
	}
	return configPath, nil
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

// runTinyTunScript writes the given script to a temp file and executes it
// with the configured shell.  stage is only used for logging.
func runTinyTunScript(stage, script string) error {
	if strings.TrimSpace(script) == "" {
		return nil
	}
	setting := configure.GetSettingNotNil()
	si, err := resolveShellInfo(setting.TunRouteShellType, setting.TunRouteShellPath)
	if err != nil {
		return fmt.Errorf("tinytun route script (%s): %w", stage, err)
	}

	// Resolve the shell binary via PATH if it is not an absolute path.
	binPath := si.bin
	if !filepath.IsAbs(binPath) {
		if resolved, err := exec.LookPath(binPath); err == nil {
			binPath = resolved
		}
	}

	// Write script to a temp file with restricted permissions.
	tmpFile, err := os.CreateTemp("", "tinytun_*"+si.ext)
	if err != nil {
		return fmt.Errorf("tinytun route script (%s): failed to create temp file: %w", stage, err)
	}
	defer os.Remove(tmpFile.Name())

	// Restrict permissions immediately after creation (before writing content).
	if runtime.GOOS != "windows" {
		if err = os.Chmod(tmpFile.Name(), 0600); err != nil {
			tmpFile.Close()
			return fmt.Errorf("tinytun route script (%s): failed to chmod temp file: %w", stage, err)
		}
	}

	if _, err = tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		return fmt.Errorf("tinytun route script (%s): failed to write temp file: %w", stage, err)
	}
	tmpFile.Close()

	// Make the file executable on non-Windows systems.
	if runtime.GOOS != "windows" {
		if err = os.Chmod(tmpFile.Name(), 0700); err != nil {
			return fmt.Errorf("tinytun route script (%s): failed to chmod temp file executable: %w", stage, err)
		}
	}

	// Build command arguments.
	var args []string
	if si.scriptFlag != "" {
		args = []string{binPath, si.scriptFlag, tmpFile.Name()}
	} else {
		args = []string{binPath, tmpFile.Name()}
	}

	log.Info("tinytun: running %s script via %s", stage, binPath)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("tinytun route script (%s) exited with error: %w", stage, err)
	}
	return nil
}

// startTinyTun generates the TinyTun config and starts the TinyTun process.
func startTinyTun(tmpl *Template) error {
	binPath, err := GetTinyTunBinPath()
	if err != nil {
		return err
	}

	configPath, err := generateTinyTunConfig(tmpl)
	if err != nil {
		return err
	}

	log.Info("Starting TinyTun from %v with config %v", binPath, configPath)

	ctx, cancel := context.WithCancel(context.Background())

	setting := configure.GetSettingNotNil()
	args := []string{binPath, "run", "--config", configPath}
	if setting.TunAutoRoute {
		args = append(args, "--auto-route")
	}
	_, err = RunWithLog(ctx, binPath, args, "", os.Environ())
	if err != nil {
		cancel()
		return fmt.Errorf("failed to start tinytun: %w", err)
	}

	tinyTunState.mu.Lock()
	if tinyTunState.cancel != nil {
		tinyTunState.cancel()
	}
	tinyTunState.cancel = cancel
	tinyTunState.mu.Unlock()

	// Run user-defined setup script when auto_route is disabled.
	if !setting.TunAutoRoute {
		if err = runTinyTunScript("setup", setting.TunSetupScript); err != nil {
			log.Warn("tinytun setup script error: %v", err)
		}
	}

	return nil
}

// stopTinyTun stops the running TinyTun process if one is active.
func stopTinyTun() {
	// Run user-defined teardown script when auto_route is disabled.
	setting := configure.GetSettingNotNil()
	if !setting.TunAutoRoute {
		if err := runTinyTunScript("teardown", setting.TunTeardownScript); err != nil {
			log.Warn("tinytun teardown script error: %v", err)
		}
	}

	tinyTunState.mu.Lock()
	cancel := tinyTunState.cancel
	tinyTunState.cancel = nil
	tinyTunState.mu.Unlock()

	if cancel != nil {
		log.Info("Stopping TinyTun")
		cancel()
	}
}

// IsTinyTunEnabled reports whether TinyTun support was compiled into this binary.
func IsTinyTunEnabled() bool { return true }
