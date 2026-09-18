package v2ray

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	tunSupported = true
	// tunBindsEgress says the core binds its sockets to the physical
	// interface (no socket mark here).
	tunBindsEgress = true
	// The kernel assigns the unit; the service finds the device by address.
	tunDeviceName = "utun"
)

var (
	tunHalfRoutes4 = []string{"0.0.0.0/1", "128.0.0.0/1"}
	tunHalfRoutes6 = []string{"::/1", "8000::/1"}
)

// tunState remembers what was installed so stop can undo exactly that.
var tunState struct {
	device string
	routes [][]string
	// dns maps network service name to the resolvers it had before.
	dns map[string][]string
}

// tunIPv6Enabled follows the same ipv6-support setting the Linux rules do.
func tunIPv6Enabled() bool { return iptables.IsIPv6Supported() }

func tunInstalled() bool { return tunState.device != "" }

func tunDevicePresent() bool {
	dev, err := tunFindDeviceOnce()
	return err == nil && dev == tunState.device
}

// tunDNSBackupPath is where the resolvers replaced by the TUN are kept
// while it runs. A crash leaves the system pointed at 127.0.0.1 with
// nothing listening; the next start restores from this file first.
func tunDNSBackupPath() string {
	return filepath.Join(conf.GetEnvironmentConfig().Config, "tun-dns-backup.json")
}

func restoreDNS(saved map[string][]string) {
	for svc, before := range saved {
		args := append([]string{"-setdnsservers", svc}, before...)
		if len(before) == 0 {
			args = []string{"-setdnsservers", svc, "Empty"}
		}
		if _, err := run("networksetup", args...); err != nil {
			log.Warn("tun: restore DNS: %v", err)
		}
	}
}

// routeAdd installs a route on the physical side. Those routes outlive a
// killed process (the utun's own routes go away with the utun), so a
// second start finds them in place and `route add` reports File exists;
// such a route is a leftover of ours, deleted and added again.
func routeAdd(args []string) (string, error) {
	out, err := run("route", args...)
	if err == nil || !strings.Contains(out, "File exists") {
		return out, err
	}
	del := append([]string{}, args...)
	del[1] = "delete"
	if _, derr := run("route", del...); derr != nil {
		return out, err
	}
	return run("route", args...)
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// defaultRoute reports the gateway and interface of the default route for
// one family, as `route -n get` prints them. The /1 routes never change
// the answer.
func defaultRoute(family string) (gateway, iface string) {
	out, err := run("route", "-n", "get", family, "default")
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		switch f[0] {
		case "gateway:":
			gateway = f[1]
		case "interface:":
			iface = f[1]
		}
	}
	return gateway, iface
}

// tunEgressInterface is the interface the IPv4 default route uses.
func tunEgressInterface() string {
	_, iface := defaultRoute("-inet")
	return iface
}

// tunFindDeviceOnce is one pass of tunFindDevice.
func tunFindDeviceOnce() (string, error) {
	addr := strings.SplitN(tunAddress4, "/", 2)[0]
	out, _ := run("ifconfig")
	current := ""
	for _, line := range strings.Split(out, "\n") {
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			current = strings.TrimSuffix(strings.Fields(line)[0], ":")
			continue
		}
		f := strings.Fields(line)
		if len(f) >= 2 && f[0] == "inet" && f[1] == addr && strings.HasPrefix(current, "utun") {
			return current, nil
		}
	}
	return "", fmt.Errorf("no utun device holds %s", addr)
}

// tunFindDevice returns the utun holding our IPv4 address.
func tunFindDevice() (string, error) {
	addr := strings.SplitN(tunAddress4, "/", 2)[0]
	deadline := time.Now().Add(5 * time.Second)
	for {
		out, _ := run("ifconfig")
		current := ""
		for _, line := range strings.Split(out, "\n") {
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				current = strings.TrimSuffix(strings.Fields(line)[0], ":")
				continue
			}
			f := strings.Fields(line)
			if len(f) >= 2 && f[0] == "inet" && f[1] == addr && strings.HasPrefix(current, "utun") {
				return current, nil
			}
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("no utun device holds %s; the core did not start its TUN inbound", addr)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func tunRoutesUp(_ *Template, nodeIPs []string) error {
	dev, err := tunFindDevice()
	if err != nil {
		return err
	}
	// Without a physical default route the core's sockets were not bound
	// to an interface, and the /1 routes would send the core's own
	// connections back into its device, each one spawning the next.
	// Refusing here leaves the device up and unrouted, which is harmless.
	if gw, iface := defaultRoute("-inet"); gw == "" || iface == "" || strings.HasPrefix(iface, "utun") {
		return fmt.Errorf("tun: no IPv4 default route; connect to a network and start the transparent proxy again")
	}
	tunState.device = dev
	tunState.routes = nil
	tunState.dns = nil
	// Host routes for the proxy nodes through the physical gateway, so
	// every process reaches the nodes directly.
	for _, ip := range nodeIPs {
		fam := "-inet"
		if strings.Contains(ip, ":") {
			fam = "-inet6"
		}
		gw, _ := defaultRoute(fam)
		if gw == "" {
			continue
		}
		args := []string{"-n", "add", fam, "-host", ip, gw}
		if _, err := routeAdd(args); err != nil {
			log.Warn("tun: node route %s: %v", ip, err)
			continue
		}
		tunState.routes = append(tunState.routes, args)
	}
	// A socket bound to the physical interface with IP_BOUND_IF looks
	// routes up in that interface's scope; once the /1 routes point at the
	// utun, the only match is theirs and the lookup fails with
	// ENETUNREACH. A scoped copy of the physical default route gives the
	// bound sockets — the core's outbounds and the DNS module's upstreams —
	// a way out again.
	families := []string{"-inet"}
	if tunIPv6Enabled() {
		families = append(families, "-inet6")
	}
	for _, fam := range families {
		gw, iface := defaultRoute(fam)
		if gw == "" || iface == "" || strings.HasPrefix(iface, "utun") {
			continue
		}
		args := []string{"-n", "add", "-ifscope", iface, fam, "default", gw}
		if _, err := routeAdd(args); err != nil {
			if fam == "-inet" {
				tunRoutesDown()
				return err
			}
			log.Warn("tun: %v", err)
			continue
		}
		tunState.routes = append(tunState.routes, args)
	}
	// Gateway form, not -interface: a socket bound to the physical
	// interface with IP_BOUND_IF gets ENETUNREACH behind an interface-form
	// route here, and works behind a gateway-form one (which is also what
	// xray's own tun installs).
	for _, p := range tunHalfRoutes4 {
		args := []string{"-n", "add", "-inet", p, tunGateway4}
		if _, err := run("route", args...); err != nil {
			tunRoutesDown()
			return err
		}
		tunState.routes = append(tunState.routes, args)
	}
	// IPv6 follows the ipv6-support setting, like the inbound's address:
	// with it off the device has no IPv6 address and routing IPv6 into it
	// would black-hole every IPv6 connection.
	if tunIPv6Enabled() {
		for _, p := range tunHalfRoutes6 {
			args := []string{"-n", "add", "-inet6", p, tunGateway6}
			if _, err := run("route", args...); err != nil {
				log.Warn("tun: %v", err)
				continue
			}
			tunState.routes = append(tunState.routes, args)
		}
	}
	// Point every active network service at the DNS module on loopback.
	// The TUN gateway address does not work here: mDNSResponder scopes its
	// queries to the service's interface (en0), and a reply that arrives
	// through the utun is dropped for such a socket, so every lookup waits
	// out a 30-second timeout. Loopback delivery has no such scope. The
	// previous resolvers are kept for stop.
	out, err := run("networksetup", "-listallnetworkservices")
	if err != nil {
		tunRoutesDown()
		return err
	}
	tunState.dns = make(map[string][]string)
	for _, svc := range strings.Split(out, "\n")[1:] {
		svc = strings.TrimSpace(svc)
		if svc == "" || strings.HasPrefix(svc, "*") {
			continue
		}
		prev, _ := run("networksetup", "-getdnsservers", svc)
		var before []string
		for _, l := range strings.Split(prev, "\n") {
			if l = strings.TrimSpace(l); l != "" && !strings.Contains(l, " ") {
				before = append(before, l)
			}
		}
		if _, err := run("networksetup", "-setdnsservers", svc, "127.0.0.1"); err != nil {
			log.Warn("tun: %v", err)
			continue
		}
		tunState.dns[svc] = before
	}
	if data, err := json.Marshal(tunState.dns); err == nil {
		if err := os.WriteFile(tunDNSBackupPath(), data, 0o600); err != nil {
			log.Warn("tun: cannot save the DNS backup: %v", err)
		}
	}
	log.Info("tun: routes and DNS installed for %s", dev)
	return nil
}

func tunRoutesDown() {
	restoreDNS(tunState.dns)
	if len(tunState.dns) > 0 {
		_ = os.Remove(tunDNSBackupPath())
	}
	for i := len(tunState.routes) - 1; i >= 0; i-- {
		args := append([]string{}, tunState.routes[i]...)
		args[1] = "delete"
		if _, err := run("route", args...); err != nil {
			log.Warn("tun: undo failed: %v", err)
		}
	}
	tunState.routes, tunState.dns, tunState.device = nil, nil, ""
}

// tunCleanupResidual restores the resolvers a crashed run left pointed at
// 127.0.0.1. The routes went away with the utun; the DNS setting did not.
func tunCleanupResidual() {
	data, err := os.ReadFile(tunDNSBackupPath())
	if err != nil {
		return
	}
	var saved map[string][]string
	if err := json.Unmarshal(data, &saved); err == nil {
		log.Warn("tun: restoring the system DNS a previous run left behind")
		restoreDNS(saved)
	}
	_ = os.Remove(tunDNSBackupPath())
}
