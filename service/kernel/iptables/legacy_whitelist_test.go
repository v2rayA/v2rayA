package iptables

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type recordedCommand struct {
	name  string
	args  []string
	input string
}

func recordLegacyWhitelistCommands(t *testing.T, available map[string]bool, failSetMatch bool, failCommands ...string) *[]recordedCommand {
	t.Helper()
	oldAvailable := commandAvailable
	oldCommands := executeCommands
	oldWithInput := executeCommandWithInput
	var calls []recordedCommand
	commandAvailable = func(command string) bool { return available[command] }
	executeCommands = func(commands string, _ bool) error {
		calls = append(calls, recordedCommand{name: "commands", input: commands})
		return nil
	}
	executeCommandWithInput = func(command string, args []string, input string) error {
		calls = append(calls, recordedCommand{name: command, args: args, input: input})
		if failSetMatch && command != "ipset" && containsArg(args, "--match-set") {
			return errors.New("xt_set unavailable")
		}
		for _, failCommand := range failCommands {
			if command+" "+strings.Join(args, " ") == failCommand {
				return errors.New("command failed")
			}
		}
		return nil
	}
	t.Cleanup(func() {
		commandAvailable = oldAvailable
		executeCommands = oldCommands
		executeCommandWithInput = oldWithInput
	})
	return &calls
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func testWhitelistSetter(table string, ipv4, ipv6 []string) Setter {
	commands := "before\n" + legacyWhitelist4Marker + "\nbetween\n" + legacyWhitelist6Marker + "\nafter\n"
	return newLegacyWhitelistSetter(commands, table, ipv4, ipv6)
}

func TestLegacyWhitelistUsesIPSetInChainOrder(t *testing.T) {
	ipv4 := []string{"1.0.0.0/8", "2.0.0.0/8", "3.0.0.0/8"}
	ipv6 := []string{"2001:db8:1::/48", "2001:db8:2::/48", "2001:db8:3::/48"}
	calls := recordLegacyWhitelistCommands(t, map[string]bool{"ipset": true}, false)

	if err := testWhitelistSetter("mangle", ipv4, ipv6).run(false, func(command string) string { return command }); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 7 {
		t.Fatalf("calls = %#v", *calls)
	}
	if !strings.Contains((*calls)[0].input, "before") || !strings.Contains((*calls)[3].input, "between") || !strings.Contains((*calls)[6].input, "after") {
		t.Fatalf("command chunks are out of order: %#v", *calls)
	}
	want4 := "create v2raya_white4 hash:net family inet maxelem 65536\nflush v2raya_white4\nadd v2raya_white4 1.0.0.0/8\nadd v2raya_white4 2.0.0.0/8\nadd v2raya_white4 3.0.0.0/8\n"
	want6 := "create v2raya_white6 hash:net family inet6 maxelem 65536\nflush v2raya_white6\nadd v2raya_white6 2001:db8:1::/48\nadd v2raya_white6 2001:db8:2::/48\nadd v2raya_white6 2001:db8:3::/48\n"
	if (*calls)[1].name != "ipset" || !reflect.DeepEqual((*calls)[1].args, []string{"restore", "-!"}) || (*calls)[1].input != want4 {
		t.Errorf("IPv4 ipset call = %#v", (*calls)[1])
	}
	if (*calls)[4].name != "ipset" || !reflect.DeepEqual((*calls)[4].args, []string{"restore", "-!"}) || (*calls)[4].input != want6 {
		t.Errorf("IPv6 ipset call = %#v", (*calls)[4])
	}
	wantMatch4 := []string{"-w", "2", "-t", "mangle", "-A", "TP_RULE", "-m", "set", "--match-set", "v2raya_white4", "dst", "-j", "RETURN"}
	wantMatch6 := []string{"-w", "2", "-t", "mangle", "-A", "TP_RULE", "-m", "set", "--match-set", "v2raya_white6", "dst", "-j", "RETURN"}
	if (*calls)[2].name != "iptables" || !reflect.DeepEqual((*calls)[2].args, wantMatch4) {
		t.Errorf("IPv4 match call = %#v", (*calls)[2])
	}
	if (*calls)[5].name != "ip6tables" || !reflect.DeepEqual((*calls)[5].args, wantMatch6) {
		t.Errorf("IPv6 match call = %#v", (*calls)[5])
	}
	for _, call := range *calls {
		if call.name == "commands" && (strings.Contains(call.input, "1.0.0.0/8") || strings.Contains(call.input, "2001:db8:1::/48")) {
			t.Errorf("per-CIDR rule leaked into command chunks: %q", call.input)
		}
	}
}

func TestLegacyWhitelistFallsBackToRestore(t *testing.T) {
	ipv4 := []string{"1.0.0.0/8", "2.0.0.0/8", "3.0.0.0/8"}
	ipv6 := []string{"2001:db8:1::/48", "2001:db8:2::/48", "2001:db8:3::/48"}
	for _, tc := range []struct {
		name         string
		available    map[string]bool
		failSetMatch bool
	}{
		{name: "without ipset", available: map[string]bool{"iptables-restore": true, "ip6tables-restore": true}},
		{name: "without xt_set", available: map[string]bool{"ipset": true, "iptables-restore": true, "ip6tables-restore": true}, failSetMatch: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := recordLegacyWhitelistCommands(t, tc.available, tc.failSetMatch)
			if err := testWhitelistSetter("nat", ipv4, ipv6).run(false, func(command string) string { return command }); err != nil {
				t.Fatal(err)
			}
			var restores []recordedCommand
			for _, call := range *calls {
				if strings.HasSuffix(call.name, "tables-restore") {
					restores = append(restores, call)
				}
			}
			if len(restores) != 2 {
				t.Fatalf("restore calls = %#v", restores)
			}
			wantArgs := []string{"-w", "2", "--noflush"}
			if restores[0].name != "iptables-restore" || !reflect.DeepEqual(restores[0].args, wantArgs) || restores[0].input != "*nat\n-A TP_RULE -d 1.0.0.0/8 -j RETURN\n-A TP_RULE -d 2.0.0.0/8 -j RETURN\n-A TP_RULE -d 3.0.0.0/8 -j RETURN\nCOMMIT\n" {
				t.Errorf("IPv4 restore call = %#v", restores[0])
			}
			if restores[1].name != "ip6tables-restore" || !reflect.DeepEqual(restores[1].args, wantArgs) || restores[1].input != "*nat\n-A TP_RULE -d 2001:db8:1::/48 -j RETURN\n-A TP_RULE -d 2001:db8:2::/48 -j RETURN\n-A TP_RULE -d 2001:db8:3::/48 -j RETURN\nCOMMIT\n" {
				t.Errorf("IPv6 restore call = %#v", restores[1])
			}
		})
	}
}

func TestLegacyWhitelistFallsBackWhenIPSetRestoreFails(t *testing.T) {
	ipv4 := []string{"1.0.0.0/8", "2.0.0.0/8"}
	ipv6 := []string{"2001:db8:1::/48", "2001:db8:2::/48"}
	calls := recordLegacyWhitelistCommands(t, map[string]bool{"ipset": true, "iptables-restore": true, "ip6tables-restore": true}, false, "ipset restore -!")

	if err := testWhitelistSetter("nat", ipv4, ipv6).run(false, func(command string) string { return command }); err != nil {
		t.Fatal(err)
	}
	var fallbackCalls []recordedCommand
	for _, call := range *calls {
		if (call.name == "ipset" && containsArg(call.args, "destroy")) || strings.HasSuffix(call.name, "tables-restore") {
			fallbackCalls = append(fallbackCalls, call)
		}
		if call.name == "commands" && strings.Contains(call.input, "-d ") {
			t.Errorf("unexpected per-line commands: %q", call.input)
		}
	}
	want := []recordedCommand{
		{name: "ipset", args: []string{"destroy", "v2raya_white4"}},
		{name: "iptables-restore", args: []string{"-w", "2", "--noflush"}, input: "*nat\n-A TP_RULE -d 1.0.0.0/8 -j RETURN\n-A TP_RULE -d 2.0.0.0/8 -j RETURN\nCOMMIT\n"},
		{name: "ipset", args: []string{"destroy", "v2raya_white6"}},
		{name: "ip6tables-restore", args: []string{"-w", "2", "--noflush"}, input: "*nat\n-A TP_RULE -d 2001:db8:1::/48 -j RETURN\n-A TP_RULE -d 2001:db8:2::/48 -j RETURN\nCOMMIT\n"},
	}
	if !reflect.DeepEqual(fallbackCalls, want) {
		t.Fatalf("fallback calls = %#v, want %#v", fallbackCalls, want)
	}
}

func TestLegacyWhitelistFallsBackWhenRestoreFails(t *testing.T) {
	ipv4 := []string{"1.0.0.0/8", "2.0.0.0/8"}
	ipv6 := []string{"2001:db8:1::/48", "2001:db8:2::/48"}
	calls := recordLegacyWhitelistCommands(t, map[string]bool{"iptables-restore": true, "ip6tables-restore": true}, false, "iptables-restore -w 2 --noflush", "ip6tables-restore -w 2 --noflush")

	if err := testWhitelistSetter("mangle", ipv4, ipv6).run(false, func(command string) string { return command }); err != nil {
		t.Fatal(err)
	}
	var fallbackCalls []recordedCommand
	for _, call := range *calls {
		if strings.HasSuffix(call.name, "tables-restore") || (call.name == "commands" && strings.Contains(call.input, "-d ")) {
			fallbackCalls = append(fallbackCalls, call)
		}
	}
	want := []recordedCommand{
		{name: "iptables-restore", args: []string{"-w", "2", "--noflush"}, input: "*mangle\n-A TP_RULE -d 1.0.0.0/8 -j RETURN\n-A TP_RULE -d 2.0.0.0/8 -j RETURN\nCOMMIT\n"},
		{name: "commands", input: "iptables -w 2 -t mangle -A TP_RULE -d 1.0.0.0/8 -j RETURN\niptables -w 2 -t mangle -A TP_RULE -d 2.0.0.0/8 -j RETURN\n"},
		{name: "ip6tables-restore", args: []string{"-w", "2", "--noflush"}, input: "*mangle\n-A TP_RULE -d 2001:db8:1::/48 -j RETURN\n-A TP_RULE -d 2001:db8:2::/48 -j RETURN\nCOMMIT\n"},
		{name: "commands", input: "ip6tables -w 2 -t mangle -A TP_RULE -d 2001:db8:1::/48 -j RETURN\nip6tables -w 2 -t mangle -A TP_RULE -d 2001:db8:2::/48 -j RETURN\n"},
	}
	if !reflect.DeepEqual(fallbackCalls, want) {
		t.Fatalf("fallback calls = %#v, want %#v", fallbackCalls, want)
	}
}

func TestLegacyWhitelistFallsBackToPerLineCommands(t *testing.T) {
	calls := recordLegacyWhitelistCommands(t, nil, false)
	if err := testWhitelistSetter("mangle", []string{"1.0.0.0/8", "2.0.0.0/8", "3.0.0.0/8"}, []string{"2001:db8:1::/48", "2001:db8:2::/48", "2001:db8:3::/48"}).run(false, func(command string) string { return command }); err != nil {
		t.Fatal(err)
	}
	var cidrCommands []string
	for _, call := range *calls {
		if call.name == "commands" && strings.Contains(call.input, "-d ") {
			cidrCommands = append(cidrCommands, call.input)
		}
	}
	want := []string{
		"iptables -w 2 -t mangle -A TP_RULE -d 1.0.0.0/8 -j RETURN\niptables -w 2 -t mangle -A TP_RULE -d 2.0.0.0/8 -j RETURN\niptables -w 2 -t mangle -A TP_RULE -d 3.0.0.0/8 -j RETURN\n",
		"ip6tables -w 2 -t mangle -A TP_RULE -d 2001:db8:1::/48 -j RETURN\nip6tables -w 2 -t mangle -A TP_RULE -d 2001:db8:2::/48 -j RETURN\nip6tables -w 2 -t mangle -A TP_RULE -d 2001:db8:3::/48 -j RETURN\n",
	}
	if !reflect.DeepEqual(cidrCommands, want) {
		t.Fatalf("per-line commands = %#v", cidrCommands)
	}
}

func TestLegacyWhitelistSkipsEmptyFamilies(t *testing.T) {
	calls := recordLegacyWhitelistCommands(t, map[string]bool{"ipset": true, "iptables-restore": true, "ip6tables-restore": true}, false)
	if err := testWhitelistSetter("nat", nil, nil).run(false, func(command string) string { return command }); err != nil {
		t.Fatal(err)
	}
	for _, call := range *calls {
		if call.name != "commands" {
			t.Fatalf("empty whitelist executed %#v", call)
		}
	}
}

func TestLegacyCleanupDestroysWhitelistSetsAfterRules(t *testing.T) {
	for _, commands := range []string{(&legacyTproxy{}).GetCleanCommands().Cmds, (&legacyRedirect{}).GetCleanCommands().Cmds} {
		destroy4 := strings.LastIndex(commands, "ipset destroy v2raya_white4")
		destroy6 := strings.LastIndex(commands, "ipset destroy v2raya_white6")
		deleteRules := strings.LastIndex(commands, "-X TP_RULE")
		if destroy4 < deleteRules || destroy6 < deleteRules {
			t.Fatalf("cleanup order is wrong:\n%s", commands)
		}
	}
}

func TestRewriteIptablesBinariesIncludesRestore(t *testing.T) {
	commands := "iptables -L\nip6tables -L\niptables-restore\nip6tables-restore\n"
	if got, want := rewriteIptablesBinaries(commands, "legacy"), "iptables-legacy -L\nip6tables-legacy -L\niptables-legacy-restore\nip6tables-legacy-restore\n"; got != want {
		t.Errorf("legacy rewrite = %q, want %q", got, want)
	}
	if got, want := rewriteIptablesBinaries(commands, "nft"), "iptables-nft -L\nip6tables-nft -L\niptables-nft-restore\nip6tables-nft-restore\n"; got != want {
		t.Errorf("nft rewrite = %q, want %q", got, want)
	}
}

func TestLegacySetupPlacesSetRulesAtWhitelistPosition(t *testing.T) {
	oldValues := legacySetupValues
	oldIPv6 := legacyIPv6Supported
	ipv4 := []string{"1.0.0.0/8", "2.0.0.0/8", "3.0.0.0/8"}
	ipv6 := []string{"2001:db8:1::/48", "2001:db8:2::/48", "2001:db8:3::/48"}
	legacySetupValues = func() ([]string, []string, []string, error) {
		return nil, ipv4, ipv6, nil
	}
	legacyIPv6Supported = func() bool { return true }
	t.Cleanup(func() {
		legacySetupValues = oldValues
		legacyIPv6Supported = oldIPv6
	})

	for _, tc := range []struct {
		name    string
		setter  Setter
		table   string
		before4 string
		after4  string
		before6 string
		after6  string
	}{
		{
			name:    "tproxy",
			setter:  (&legacyTproxy{}).GetSetupCommands(),
			table:   "mangle",
			before4: "iptables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN",
			after4:  "iptables -w 2 -t mangle -A TP_RULE -j TP_MARK",
			before6: "ip6tables -w 2 -t mangle -A TP_RULE -m mark --mark 0x40/0xc0 -j RETURN",
			after6:  "ip6tables -w 2 -t mangle -A TP_RULE -j TP_MARK",
		},
		{
			name:    "redirect",
			setter:  (&legacyRedirect{}).GetSetupCommands(),
			table:   "nat",
			before4: "iptables -w 2 -t nat -A TP_RULE -m mark --mark 0x80/0x80 -j RETURN",
			after4:  "iptables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 52345",
			before6: "ip6tables -w 2 -t nat -A TP_RULE -m mark --mark 0x80/0x80 -j RETURN",
			after6:  "ip6tables -w 2 -t nat -A TP_RULE -p tcp -j REDIRECT --to-ports 52345",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := recordLegacyWhitelistCommands(t, map[string]bool{"ipset": true}, false)
			if err := tc.setter.run(false, func(command string) string { return command }); err != nil {
				t.Fatal(err)
			}
			assertSetRulePosition(t, *calls, "iptables", tc.table, "v2raya_white4", tc.before4, tc.after4)
			assertSetRulePosition(t, *calls, "ip6tables", tc.table, "v2raya_white6", tc.before6, tc.after6)
		})
	}
}

func assertSetRulePosition(t *testing.T, calls []recordedCommand, binary, table, setName, before, after string) {
	t.Helper()
	var matches []int
	for i, call := range calls {
		if call.name == binary && containsArg(call.args, setName) {
			matches = append(matches, i)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("%s match rules = %d in %#v", setName, len(matches), calls)
	}
	i := matches[0]
	if i < 2 || i+1 >= len(calls) ||
		calls[i-2].name != "commands" || !strings.Contains(calls[i-2].input, before) ||
		calls[i+1].name != "commands" || !strings.Contains(calls[i+1].input, after) {
		t.Fatalf("%s rule is out of position in %#v", setName, calls)
	}
	wantArgs := []string{"-w", "2", "-t", table, "-A", "TP_RULE", "-m", "set", "--match-set", setName, "dst", "-j", "RETURN"}
	if !reflect.DeepEqual(calls[i].args, wantArgs) {
		t.Errorf("%s match args = %#v, want %#v", setName, calls[i].args, wantArgs)
	}
}
