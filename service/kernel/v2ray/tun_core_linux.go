package v2ray

import (
	"fmt"
	"net"
	"net/netip"
	"os/exec"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/kernel/iptables"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	tunSupported = true
	// tunBindsEgress says the core binds its sockets to the physical
	// interface (no socket mark here).
	tunBindsEgress = false
	tunDeviceName  = "tun0"

	// tunRouteTable holds the default route into the device; the rules
	// below send everything there that the main table cannot place on a
	// more specific route. Marked packets (0x80: the core's own sockets,
	// the direct outbound, the DNS module's upstreams) look up main in
	// full and leave through the physical interface, or fail with
	// ENETUNREACH rather than fall into the TUN when main has no route
	// for their family.
	tunRouteTable  = "2020"
	tunPrefMarked  = "7000"
	tunPrefUnreach = "7001"
	tunPrefNode    = "7005"
	tunPrefMain    = "7010"
	tunPrefTun     = "7020"
)

// tunRules lists the policy rules as `ip rule show` prints them; the same
// text is used to install, to recognise a leftover as ours, and to delete.
var tunRules = []string{
	tunPrefMarked + ":\tfrom all fwmark 0x80/0x80 lookup main",
	tunPrefUnreach + ":\tfrom all fwmark 0x80/0x80 unreachable",
	tunPrefMain + ":\tfrom all lookup main suppress_prefixlength 0",
	tunPrefTun + ":\tfrom all lookup " + tunRouteTable,
}

// tunUndo records the commands whose undo must run on stop, most recent
// last.
var tunUndo []string

func tunInstalled() bool { return len(tunUndo) > 0 }

func tunIPv6Enabled() bool { return iptables.IsIPv6Supported() }

func tunEgressInterface() string { return "" }

func ipCmd(args ...string) (string, error) {
	out, err := exec.Command("ip", args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("ip %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// tunConflicts reports rules in our preference range and routes in our
// table that are not ours. Ours are deleted so a crashed previous run does
// not block the next; anything else is the user's and stops the start.
func tunConflicts(family string) error {
	out, err := ipCmd(family, "rule", "show")
	if err != nil {
		return fmt.Errorf("cannot list policy rules: %w", err)
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			continue
		}
		pref := line[:colon]
		if pref < tunPrefMarked || pref > tunPrefTun || len(pref) != 4 {
			continue
		}
		ours := false
		for _, r := range tunRules {
			if strings.ReplaceAll(line, "\t", " ") == strings.ReplaceAll(r, "\t", " ") {
				ours = true
			}
		}
		// Node rules are ours when they name a single host, which is
		// how the kernel prints the ones tunRoutesUp installs; a rule
		// with a prefix at this preference is somebody else's.
		if pref == tunPrefNode && strings.HasSuffix(line, " lookup main") {
			if f := strings.Fields(line); len(f) >= 4 && f[len(f)-4] == "to" {
				if _, err := netip.ParseAddr(f[len(f)-3]); err == nil {
					ours = true
				}
			}
		}
		if !ours {
			return fmt.Errorf("policy rule %q occupies preference %s reserved for the TUN; remove it or choose another transparent proxy mode", line, pref)
		}
		// Delete by the full selector, as printed: a bare preference
		// would delete whichever rule the kernel lists first there.
		args := append([]string{family, "rule", "del", "pref", pref}, strings.Fields(line[colon+1:])...)
		if _, err := ipCmd(args...); err != nil {
			log.Warn("tun: leftover rule %q: %v", line, err)
		}
	}
	// An IPv6 table that was never populated does not exist yet and the
	// listing fails; that is an empty table, not a conflict.
	out, err = ipCmd(family, "route", "show", "table", tunRouteTable)
	if err != nil {
		out = ""
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "default dev "+tunDeviceName) {
			return fmt.Errorf("route %q occupies table %s reserved for the TUN; remove it or choose another transparent proxy mode", line, tunRouteTable)
		}
	}
	_, _ = ipCmd(family, "route", "flush", "table", tunRouteTable)
	return nil
}

func tunDevicePresent() bool {
	_, err := net.InterfaceByName(tunDeviceName)
	return err == nil
}

func tunRoutesUp(_ *Template, nodeIPs []string) error {
	tunUndo = nil
	families := []string{"-4"}
	if tunIPv6Enabled() {
		families = append(families, "-6")
	}
	for _, f := range families {
		if err := tunConflicts(f); err != nil {
			return err
		}
	}
	// The core creates the device while starting; afterStart runs once its
	// API answers, which is normally after that, but do not race it.
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := ipCmd("link", "show", tunDeviceName); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("device %s did not appear; the core did not start its TUN inbound", tunDeviceName)
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, f := range families {
		steps := [][]string{
			{f, "route", "add", "default", "dev", tunDeviceName, "table", tunRouteTable},
			{f, "rule", "add", "pref", tunPrefMarked, "fwmark", "0x80/0x80", "lookup", "main"},
			{f, "rule", "add", "pref", tunPrefUnreach, "fwmark", "0x80/0x80", "unreachable"},
		}
		// The proxy nodes themselves stay reachable through the physical
		// route for every process (an ssh session to one's own VPS must
		// not be sent through that VPS); these sit before the default
		// route suppression so the main table's default applies to them.
		for _, ip := range nodeIPs {
			if parsed := net.ParseIP(ip); parsed == nil || (f == "-6") != (parsed.To4() == nil) {
				continue
			}
			steps = append(steps, []string{f, "rule", "add", "pref", tunPrefNode, "to", ip, "lookup", "main"})
		}
		steps = append(steps,
			[]string{f, "rule", "add", "pref", tunPrefMain, "lookup", "main", "suppress_prefixlength", "0"},
			[]string{f, "rule", "add", "pref", tunPrefTun, "lookup", tunRouteTable},
		)
		for _, step := range steps {
			if _, err := ipCmd(step...); err != nil {
				tunRoutesDown()
				return err
			}
			tunUndo = append(tunUndo, strings.Join(step, " "))
		}
	}
	log.Info("tun: routes and policy rules installed for %s", tunDeviceName)
	return nil
}

// tunRoutesDown undoes the recorded commands in reverse. A failure is
// logged and the rest still run; the device itself vanishes with the core.
func tunRoutesDown() {
	for i := len(tunUndo) - 1; i >= 0; i-- {
		args := strings.Fields(tunUndo[i])
		for j, a := range args {
			if a == "add" {
				args[j] = "del"
				break
			}
		}
		if _, err := ipCmd(args...); err != nil {
			log.Warn("tun: undo failed: %v", err)
		}
	}
	tunUndo = nil
}

// tunCleanupResidual removes our rules and table entries left by an
// abnormal exit; it never touches anything that is not exactly ours.
func tunCleanupResidual() {
	for _, f := range []string{"-4", "-6"} {
		_ = tunConflicts(f)
	}
}
