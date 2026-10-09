package configure

import "testing"

// The DNS interception opt-out must not change what an existing installation
// does, so both a fresh setting and one stored before the field existed
// resolve to "on". FillEmpty leaves the zero value alone for bools only;
// DefaultYesNo is a string, so an absent field takes the default.
func TestDnsHijackDefaultsToOn(t *testing.T) {
	if got := NewSetting().DnsHijack; got != Yes {
		t.Fatalf("new setting dnsHijack = %q, want %q", got, Yes)
	}

	stored := NewSetting()
	if err := SetSetting(stored); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetSetting(NewSetting()) })

	if got := GetSettingNotNil().DnsHijack; got != Yes {
		t.Fatalf("stored setting dnsHijack = %q, want %q", got, Yes)
	}
}

// An older database holds no dnsHijack key at all. Decoding over it leaves the
// field empty, and the empty value has to mean "on" as well.
func TestDnsHijackAbsentFromAnOlderConfigKeepsInterception(t *testing.T) {
	setting := &Setting{}
	if err := SetSetting(setting); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetSetting(NewSetting()) })

	if got := GetSettingNotNil().DnsHijack; got != Yes {
		t.Fatalf("migrated dnsHijack = %q, want %q", got, Yes)
	}
}

func TestMigrateSettingKeepsAnExplicitOptOut(t *testing.T) {
	setting := NewSetting()
	setting.DnsHijack = No
	MigrateSetting(setting)
	if setting.DnsHijack != No {
		t.Fatalf("migration overrode the opt-out: %q", setting.DnsHijack)
	}

	// MigrateSetting runs on the request body, where an older GUI that does
	// not know the field leaves it empty; that must become "on", not "off".
	old := &Setting{}
	MigrateSetting(old)
	if old.DnsHijack != Yes {
		t.Fatalf("MigrateSetting dnsHijack = %q, want %q", old.DnsHijack, Yes)
	}
}
