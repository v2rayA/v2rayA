package configure

import "testing"

func TestNodeDNSMigration(t *testing.T) {
	for _, value := range []string{"", "auto", "https://1.1.1.1/dns-query"} {
		setting := &Setting{NodeDns: value}
		MigrateSetting(setting)
		want := value
		if want == "" {
			want = "auto"
		}
		if setting.NodeDns != want {
			t.Fatalf("%q migrated to %q", value, setting.NodeDns)
		}
	}
	if NewSetting().NodeDns != "auto" {
		t.Fatal("default node DNS is not auto")
	}
}
