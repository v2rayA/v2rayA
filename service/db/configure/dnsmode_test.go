package configure

import "testing"

// The three modes separate "does the DNS module run" from "does the system
// send its queries to it". Only hijack does the second, so the two
// predicates must never collapse into one.
func TestDnsModeResolvesToModuleAndInterception(t *testing.T) {
	for _, tc := range []struct {
		mode         DnsMode
		service      bool
		interception bool
	}{
		{DnsModeOff, false, false},
		{DnsModeService, true, false},
		{DnsModeHijack, true, true},
	} {
		t.Run(string(tc.mode), func(t *testing.T) {
			setting := &Setting{DnsMode: tc.mode, DnsHijack: Yes}
			if got := ResolveDnsMode(setting); got != tc.mode {
				t.Fatalf("resolved %q, want %q", got, tc.mode)
			}
			if got := setting.DnsServiceEnabled(); got != tc.service {
				t.Fatalf("module enabled = %v, want %v", got, tc.service)
			}
			if got := setting.DnsInterceptionEnabled(); got != tc.interception {
				t.Fatalf("interception enabled = %v, want %v", got, tc.interception)
			}
		})
	}
}

// An installation from before the mode existed carries only the opt-out. An
// explicit "no" meant neither module nor interception; everything else meant
// both, which is what an absent field has to keep meaning.
func TestAbsentDnsModeDerivesFromTheLegacyOptOut(t *testing.T) {
	for _, tc := range []struct {
		name   string
		hijack DefaultYesNo
		want   DnsMode
	}{
		{"absent", "", DnsModeHijack},
		{"default", Default, DnsModeHijack},
		{"yes", Yes, DnsModeHijack},
		{"no", No, DnsModeOff},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setting := &Setting{DnsHijack: tc.hijack}
			if got := ResolveDnsMode(setting); got != tc.want {
				t.Fatalf("resolved %q, want %q", got, tc.want)
			}
		})
	}
}

// A value nobody chose must not turn interception on. It resolves to the
// safest mode and the API refuses to store it.
func TestUnknownDnsModeNeverEnablesInterception(t *testing.T) {
	setting := &Setting{DnsMode: DnsMode("hijackk"), DnsHijack: Yes}
	if got := ResolveDnsMode(setting); got != DnsModeOff {
		t.Fatalf("resolved %q, want %q", got, DnsModeOff)
	}
	if setting.DnsInterceptionEnabled() {
		t.Fatal("an unknown mode must not intercept")
	}
	if err := ValidateDnsMode(DnsMode("hijackk")); err == nil {
		t.Fatal("an unknown mode must be rejected by the API")
	}
	for _, mode := range []DnsMode{"", DnsModeOff, DnsModeService, DnsModeHijack} {
		if err := ValidateDnsMode(mode); err != nil {
			t.Fatalf("mode %q rejected: %v", mode, err)
		}
	}
}

// MigrateSetting writes the mode out from the opt-out so the next read does
// not have to derive it again, and never overrides a mode that is already set.
func TestMigrateSettingDerivesTheModeFromTheOptOut(t *testing.T) {
	old := &Setting{DnsHijack: No}
	MigrateSetting(old)
	if old.DnsMode != DnsModeOff {
		t.Fatalf("legacy opt-out migrated to %q, want %q", old.DnsMode, DnsModeOff)
	}

	fresh := &Setting{}
	MigrateSetting(fresh)
	if fresh.DnsMode != DnsModeHijack {
		t.Fatalf("empty setting migrated to %q, want %q", fresh.DnsMode, DnsModeHijack)
	}

	chosen := &Setting{DnsMode: DnsModeService, DnsHijack: Yes}
	MigrateSetting(chosen)
	if chosen.DnsMode != DnsModeService {
		t.Fatalf("migration overrode the explicit mode: %q", chosen.DnsMode)
	}
}

// The opt-out mirrors the mode rather than drifting from it, so a client that
// only knows dnsHijack reads the decision that is in force.
func TestMigrateSettingMirrorsTheModeIntoTheOptOut(t *testing.T) {
	for _, tc := range []struct {
		mode DnsMode
		want DefaultYesNo
	}{
		{DnsModeHijack, Yes},
		{DnsModeService, No},
		{DnsModeOff, No},
	} {
		setting := &Setting{DnsMode: tc.mode}
		MigrateSetting(setting)
		if setting.DnsHijack != tc.want {
			t.Fatalf("dnsMode %s left dnsHijack = %q, want %q", tc.mode, setting.DnsHijack, tc.want)
		}
	}
}

// A database written before the mode existed must not have the default filled
// in before the migration can read the opt-out: an installation that had
// turned interception off would get it back.
func TestStoredOptOutIsNotOverwrittenByTheModeDefault(t *testing.T) {
	legacy := &Setting{DnsHijack: No}
	if err := SetSetting(legacy); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetSetting(NewSetting()) })

	stored := GetSettingNotNil()
	if stored.DnsMode != DnsModeOff {
		t.Fatalf("stored dnsMode = %q, want %q", stored.DnsMode, DnsModeOff)
	}
	if stored.DnsInterceptionEnabled() {
		t.Fatal("the stored opt-out must survive the read")
	}
}
