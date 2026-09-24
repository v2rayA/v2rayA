package iptables

import (
	"fmt"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	legacyWhitelist4Marker = "__V2RAYA_WHITE4__"
	legacyWhitelist6Marker = "__V2RAYA_WHITE6__"
)

var (
	legacySetupValues   = getTproxySetupValues
	legacyIPv6Supported = IsIPv6Supported
)

type setterStep struct {
	commands  string
	whitelist *legacyWhitelist
}

type legacyWhitelist struct {
	binary   string
	table    string
	setName  string
	family   string
	cidrs    []string
	warnOnce *sync.Once
}

func newLegacyWhitelistSetter(commands, table string, ipv4, ipv6 []string) Setter {
	warnOnce := &sync.Once{}
	families := []struct {
		marker    string
		binary    string
		setName   string
		setFamily string
		cidrs     []string
	}{
		{legacyWhitelist4Marker, "iptables", "v2raya_white4", "inet", ipv4},
		{legacyWhitelist6Marker, "ip6tables", "v2raya_white6", "inet6", ipv6},
	}
	var steps []setterStep
	for _, family := range families {
		before, after, found := strings.Cut(commands, family.marker)
		if !found {
			continue
		}
		steps = append(steps, setterStep{commands: before})
		if len(family.cidrs) > 0 {
			steps = append(steps, setterStep{whitelist: &legacyWhitelist{
				binary:   family.binary,
				table:    table,
				setName:  family.setName,
				family:   family.setFamily,
				cidrs:    family.cidrs,
				warnOnce: warnOnce,
			}})
		}
		commands = after
	}
	steps = append(steps, setterStep{commands: commands})
	return Setter{steps: steps}
}

func (w *legacyWhitelist) run(stopAtError bool, rewrite func(string) string) error {
	if commandAvailable("ipset") {
		if err := executeCommandWithInput("ipset", []string{"restore", "-!"}, w.ipsetPayload()); err == nil {
			args := []string{"-w", "2", "-t", w.table, "-A", "TP_RULE", "-m", "set", "--match-set", w.setName, "dst", "-j", "RETURN"}
			if err := executeCommandWithInput(rewrite(w.binary), args, ""); err == nil {
				return nil
			}
		}
		_ = executeCommandWithInput("ipset", []string{"destroy", w.setName}, "")
	}

	w.warnOnce.Do(func() {
		log.Warn("transparent proxy IP whitelist matching remains linear; install ipset with xt_set support or use nftables")
	})
	restore := rewrite(w.binary + "-restore")
	if commandAvailable(restore) {
		if err := executeCommandWithInput(restore, []string{"-w", "2", "--noflush"}, w.restorePayload()); err == nil {
			return nil
		}
	}
	return executeCommands(rewrite(w.perLineCommands()), stopAtError)
}

func (w *legacyWhitelist) ipsetPayload() string {
	var payload strings.Builder
	fmt.Fprintf(&payload, "create %s hash:net family %s maxelem 65536\n", w.setName, w.family)
	fmt.Fprintf(&payload, "flush %s\n", w.setName)
	for _, cidr := range w.cidrs {
		fmt.Fprintf(&payload, "add %s %s\n", w.setName, cidr)
	}
	return payload.String()
}

func (w *legacyWhitelist) restorePayload() string {
	var payload strings.Builder
	fmt.Fprintf(&payload, "*%s\n", w.table)
	for _, cidr := range w.cidrs {
		fmt.Fprintf(&payload, "-A TP_RULE -d %s -j RETURN\n", cidr)
	}
	payload.WriteString("COMMIT\n")
	return payload.String()
}

func (w *legacyWhitelist) perLineCommands() string {
	var commands strings.Builder
	for _, cidr := range w.cidrs {
		fmt.Fprintf(&commands, "%s -w 2 -t %s -A TP_RULE -d %s -j RETURN\n", w.binary, w.table, cidr)
	}
	return commands.String()
}
